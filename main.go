package main

import (
	"fmt"
	// "time"
	"elevatorproject/elevatorManager"
	"elevatorproject/driver-go/elevio"
	"elevatorproject/config"
	// "./network/tcp"
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

	// Start orderHandler
	go orderHandler.Init(chan1, chan2, chan3)

	// Start elevatorManager
	go elevatorManager.Init(chan4, chan5)

	// Start networking
	go networking.Init(chan1, chan2, chan3, chan4, chan5)
}
