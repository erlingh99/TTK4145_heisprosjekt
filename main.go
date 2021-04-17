package main

import (

	// "time"
	"elevatorproject/config"
	em "elevatorproject/elevatorManager"
	"elevatorproject/network/localip"
	"elevatorproject/network/peers"
	"elevatorproject/networking"
	oh "elevatorproject/orderHandler"
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

	ordersUpdate := make(chan orders.Order)
	elevStateChange := make(chan em.ElevState)
	backupChan := make(chan oh.DistributerState)

	orders := make(chan map[string][][]bool)

	enableIpBroadcast := make(chan bool)	
	broadcastReciever := make(chan string)

	go peers.Transmitter(config.BCAST_PORT, localip, enableIpBroadcast) //can use peers to broadcast since only one thing is broadcasted
	go peers.Receiver(config.BCAST_PORT, broadcastReciever)

	


	// Start orderHandler
	go oh.Distributer(elevatorID, broad, chan2, chan3, broadcastReciever)

	// Start elevatorManager
	go em.ElevatorManager(orderChange<-, <-orders, id)

	// Start networking
	go networking.Init(chan1, chan2, chan3, chan4, chan5)



	//for oversikt
	availabilityChan := make(chan bool)
	peerUpdateChannel := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PEER_PORT, id ,availabilityChan) //maybe not use this. Not really neccessary of network module alerts orderhandler of new conns and broken conns
	go peers.Receiver(config.PEER_PORT, peerUpdateChannel)

	for {
		select {
			case p<-peerpeerUpdateChannel:		
				fmt.Println("PEER UPDATE")
				fmt.Printf("* Peers: %v\n", p.Peers)
				fmt.Printf("* New: %v\n", p.New)
				fmt.Printf("* Lost: %v\n", p.Lost)
		}
	}
}
