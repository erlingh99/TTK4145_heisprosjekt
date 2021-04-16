package main

import (
	"fmt"
	// "time"
	. "elevatorproject/elevatorManager"
	. "elevatorproject/orderHandler"
	"elevatorproject/driver-go/elevio"
	"elevatorproject/config"
	"elevatorproject/network/peers"
	"elevatorproject/networking"
	// "./test"
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


	var elevID string
	flag.Var(&elevID, "n", "Name of elevator")

	if elevID == nil {
		elevID, err = localip.LocalIP()
	}

	

	availabilityChan := make(chan bool)
	peerUpdateChannel := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PEER_PORT, elevID,availabilityChan)
	go peers.Receiver(config.PEER_PORT, peerUpdateChannel)

	// Start orderHandler
	go OrderHandler(chan1, chan2, chan3, peerUpdateChannel)

	// Start elevatorManager
	go ElevatorManager(chan4, chan5, elevID)

	// Start networking
	go networking.Init(chan1, chan2, chan3, chan4, chan5)
}
