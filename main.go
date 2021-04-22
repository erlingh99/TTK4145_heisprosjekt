package main

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	em "elevatorproject/elevatorManager"
	"elevatorproject/network/localip"
	"elevatorproject/utils"
	"time"

	"elevatorproject/network/peers"
	oh "elevatorproject/orderHandler"

	bcast_ack "elevatorproject/network/bcast_with_ackCtrl"
	"elevatorproject/orders"
	"flag"
	"fmt"
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

	ordersFromElevatorOut := make(chan orders.Order, 1)
	ordersFromElevatorIn := make(chan orders.Order, 1)


	elevStateChangeOut := make(chan em.Elevator, 1)
	elevStateChangeIn := make(chan em.Elevator, 1)


	backupOut := make(chan oh.DistributerState, 1)
	backupIn := make(chan oh.DistributerState, 1)

	ordersToElevatorsOut := make(chan map[string][config.N_FLOORS][config.N_BUTTONS]bool, 1)
	ordersToElevatorsIn := make(chan map[string][config.N_FLOORS][config.N_BUTTONS]bool, 1)

	broadcastReciever := make(chan string, 1) //obsolete
	enableIpBroadcast := make(chan bool, 1)   //obsolete


	elevDisconnect := make(chan string, 1)


	// Start orderHandler
	go oh.Distributer(elevatorID, ordersFromElevatorIn, elevStateChangeIn, broadcastReciever, backupIn, elevDisconnect, ordersToElevatorsOut, enableIpBroadcast, backupOut)

	// Start elevatorManager
	go em.ElevatorManager(elevatorID, ordersToElevatorsIn, ordersFromElevatorOut, elevStateChangeOut)




	peerIDs := make([]string, 0)
	waitingForAcks := make(bcast_ack.AckList, 0)

	AckRecieved := make(chan bcast_ack.AcknowledgeMsg)
	AckSend := make(chan bcast_ack.AcknowledgeMsg)
	AckNeeded := make(chan bcast_ack.AcknowledgeCtrl)


	//recvChans: ordersFromElevatorIn, elevStateChangeIn, backupIn, ordersToElevatorsIn, AckRecieved
	//sendChans: ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut, AckSend


	go bcast_ack.Transmitter(elevatorID, config.BCAST_PORT, AckNeeded, ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut, AckSend)
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
					fmt.Println("Lost peer " + lostPeer)
					peerIDs = utils.Remove(peerIDs, lostPeer)
					elevDisconnect <- lostPeer
				}
			}

		case <-msgResendTicker.C:
			//fmt.Println("ack ticker")
			for _, ack := range waitingForAcks {
				if time.Now().After(ack.SendTime.Add(time.Duration(ack.SendNum) * config.RESEND_RATE)) {
					// use reflect with ack.msg to resend on correct sendchan
					bcast_ack.ResendMsg(ack.Msg, ordersFromElevatorOut, elevStateChangeOut, backupOut, ordersToElevatorsOut, AckSend)
				}
			}
		
		case ack := <-AckRecieved:
			if ack.ElevID == elevatorID {
				continue
			}
			waitingForAcks.AckRecieved(&ack)			
			//fmt.Println("ack recv " + ack.ID)

		case ack := <-AckNeeded:
			// fmt.Println("ack needed")
			// fmt.Println(ack.Msg)
			ack.AcksNeeded = peerIDs		
			waitingForAcks.AddAck(&ack)
		}

		waitingForAcks.RemoveCompletedAcks()
		regElev := waitingForAcks.CheckForTimedoutSends()
		if len(regElev) == 0 {continue}
		//find what elevators are not responding
		for _, p := range peerIDs {
			if !utils.Contains(regElev, p) {
				//p is not responding
				peerIDs = utils.Remove(peerIDs, p)
				elevDisconnect <- p
				//might be a problem if good elevator is removed, because it will not reconnect with peers then

				fmt.Println("peer timedout" + p)
			}
		}
	}	
}

//TODO
//do they detect eachother?


//when is orderhandler master, when slave?

//how to check for bad elevator?

//using peers to remove missing acks, wab reconnects etcc

//acks needed. Everyone active, or just the ones active when msg was created?
//wab new and lost peers`?