package main

import (

	// "time"
	"elevatorproject/config"
	. "elevatorproject/elevatorManager"
	"elevatorproject/network/localip"
	"elevatorproject/network/peers"
	"elevatorproject/networking"
	. "elevatorproject/orderHandler"
	"flag"
	"fmt"
	"os"
)

func main() {
	// elevio.Init("localhost:15657", config.N_FLOORS)
	// fmt.Println("Starting elevatorManager")
	// go elevatorManager.ElevatorManager()
	// for {}


	// Create communication channels
	chan1 := make(chan int)
	chan2 := make(chan string)
	chan3 := make(chan Order)
	chan4 := make(chan ElevState)
	chan5 := make(chan OrderHandlerState)

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	

	availabilityChan := make(chan bool)
	peerUpdateChannel := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PEER_PORT, id ,availabilityChan)
	go peers.Receiver(config.PEER_PORT, peerUpdateChannel)

	// Start orderHandler
	go OrderHandler(chan1, chan2, chan3, peerUpdateChannel)

	// Start elevatorManager
	go ElevatorManager(chan4, chan5, elevID)

	// Start networking
	go networking.Init(chan1, chan2, chan3, chan4, chan5)
}
