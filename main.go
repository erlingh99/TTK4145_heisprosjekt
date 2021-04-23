package main

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	em "elevatorproject/elevatorManager"
	"elevatorproject/network/bcast"
	"elevatorproject/network/localip"
	"elevatorproject/network/switcher"
	"elevatorproject/utils"
	"time"

	"elevatorproject/network/peers"
	oh "elevatorproject/orderHandler"

	bcast_ack "elevatorproject/network/bcast_with_ackCtrl"
	"elevatorproject/orders"
	"flag"
	"fmt"
	"math/rand"
	"os"
)


func main() {

	var elevatorID string
	flag.StringVar(&elevatorID, "id", "", "id of this peer")
	flag.Parse()	

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if elevatorID == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		elevatorID = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	elevio.Init("localhost:15657", config.N_FLOORS)


	//out = channel the msg is originally sent on
	//out2 = channel the msg is passed into if internet connection exist, gos to network
	//In = channel the msg is passed into if internet connection exist, gos to local module

	ordersFromElevatorOut := make(chan orders.Order, 1)
	ordersFromElevatorOut2 := make(chan orders.Order, 1)
	ordersFromElevatorIn := make(chan orders.Order, 1)


	elevStateChangeOut := make(chan em.Elevator, 1)
	elevStateChangeOut2 := make(chan em.Elevator, 1)
	elevStateChangeIn := make(chan em.Elevator, 1)


	backupOut := make(chan oh.DistributerState, 1)
	backupOut2 := make(chan oh.DistributerState, 1)
	backupIn := make(chan oh.DistributerState, 1)

	ordersToElevatorsOut := make(chan map[string][config.N_FLOORS][config.N_BUTTONS]bool, 1)
	ordersToElevatorsOut2 := make(chan map[string][config.N_FLOORS][config.N_BUTTONS]bool, 1)
	ordersToElevatorsIn := make(chan map[string][config.N_FLOORS][config.N_BUTTONS]bool, 1)

	broadcastReciever := make(chan string, 1)
	enableIpBroadcast := make(chan bool, 1)


	elevDisconnect := make(chan string, 1)

	// Start orderHandler
	go oh.Distributer(elevatorID, ordersFromElevatorIn, elevStateChangeIn, broadcastReciever, backupIn, elevDisconnect, ordersToElevatorsOut, enableIpBroadcast, backupOut)

	// Start elevatorManager
	go em.ElevatorManager(elevatorID, ordersToElevatorsIn, ordersFromElevatorOut, elevStateChangeOut)

	go imMasterAlertBcast(config.MASTER_BCAST_PORT, enableIpBroadcast, elevatorID)
	go bcast.Receiver(config.MASTER_BCAST_PORT, broadcastReciever)

	peerIDs := make([]string, 0)
	waitingForAcks := make(bcast_ack.AckList, 0)

	AckRecieved := make(chan bcast_ack.AcknowledgeMsg)
	AckSend := make(chan bcast_ack.AcknowledgeMsg)
	AckNeeded := make(chan bcast_ack.AcknowledgeCtrl)


	//recvChans: ordersFromElevatorIn, elevStateChangeIn, backupIn, ordersToElevatorsIn, AckRecieved
	//sendChans: ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut, AckSend
	
	go switcher.Switcher(ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut,
						ordersFromElevatorOut2, elevStateChangeOut2, backupOut2, ordersToElevatorsOut2,
						ordersFromElevatorIn, elevStateChangeIn, backupIn, ordersToElevatorsIn)		


	go bcast_ack.Transmitter(elevatorID, config.BCAST_PORT, AckNeeded, ordersFromElevatorOut2, elevStateChangeOut2, backupOut2, ordersToElevatorsOut2, AckSend)
	go bcast_ack.Receiver(elevatorID, config.BCAST_PORT, AckSend, ordersFromElevatorIn, elevStateChangeIn, backupIn, ordersToElevatorsIn, AckRecieved)

	peerChan := make(chan peers.PeerUpdate)
	transmitEnable := make(chan bool)
	go peers.Transmitter(config.PEER_PORT, elevatorID, transmitEnable)
	go peers.Receiver(config.PEER_PORT, peerChan)

	msgResendTicker := time.NewTicker(config.RESEND_RATE)

	for {
		select {
		case p := <-peerChan:		
			fmt.Println("PEER UPDATE")
			fmt.Printf("* Peers: %v\n", p.Peers)
			fmt.Printf("* New: %v\n", p.New)
			fmt.Printf("* Lost: %v\n", p.Lost)

			peerIDs = append(peerIDs, p.New)


			if len(p.Lost) > 0 {		
				for _, lostPeer := range p.Lost {
					if lostPeer != elevatorID {
						fmt.Println("Lost peer " + lostPeer)
						peerIDs = utils.Remove(peerIDs, lostPeer)
						elevDisconnect <- lostPeer
					}
				}
			}

		case <-msgResendTicker.C:
			//fmt.Println("ack ticker")
			for _, ack := range waitingForAcks {
				//find messages to resend
				if time.Now().After(ack.SendTime.Add(time.Duration(ack.SendNum) * config.RESEND_RATE)) {
					// use reflect with ack.msg to resend on correct sendchan
					bcast_ack.ResendMsg(ack.Msg, ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut, AckSend)
				}
			}
		
		case ack := <-AckRecieved:
			waitingForAcks.AckRecieved(&ack)			

		case ack := <-AckNeeded:
			ack.AcksNeeded = peerIDs		
			waitingForAcks.AddAck(&ack)
		}

		waitingForAcks.RemoveCompletedAcks()
		missingAcks := waitingForAcks.CheckForTimedoutSends()
		if len(missingAcks) == 0 {continue}
		//find what elevators are not responding
		for _, v := range missingAcks {
			peerIDs = utils.Remove(peerIDs, v)
			elevDisconnect <- v					
			fmt.Println("missing acks from: " + v + "\n ^has been removed")							
		}
	}	
}

func imMasterAlertBcast(port int, enableBcast <-chan bool, id string) {
	enabled := false
	bcastCh := make(chan string)
	go bcast.Transmitter(port, bcastCh)
	for {
		select {
		case b := <-enableBcast:
			enabled = b
		default:
			if enabled {
				bcastCh<- id
			}
			// sleep some random time to avoid master -> slave -> master cycle when 2 elevators compete for master
			time.Sleep(time.Duration(config.MASTER_BROADCAST_INTERVAL_MIN + rand.Intn(config.MASTER_BROADCAST_INTERVAL_MAX - config.MASTER_BROADCAST_INTERVAL_MIN)) * time.Millisecond)
		}
	}
}