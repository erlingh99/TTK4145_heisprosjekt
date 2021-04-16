package main

import (
	"fmt"
	// "time"
	"elevatorproject/elevatorManager"
	"elevatorproject/driver-go/elevio"
	"elevatorproject/config"
	// "./network/tcp"
	//"elevatorproject/Networking"
	// "./test"
)

func main() {
	elevio.Init("10.22.231.9:15657", config.N_FLOORS)
	fmt.Println("Starting elevatorManager")
	go elevatorManager.ElevatorManager()
	for {}
}
