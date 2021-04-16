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
	elevio.Init("localhost:15657", config.N_FLOORS)
	fmt.Println("Starting elevatorManager")
	go elevatorManager.ElevatorManager()
	for {}
}
