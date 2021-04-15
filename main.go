package main

import (
	"fmt"
	// "time"
	"elevatorproject/config"
	// "./network/tcp"
	"elevatorproject/Networking"
	// "./test"
)

func main() {
	fmt.Println("Number of floors:", config.N_FLOORS);
	Networking.Init()
	// Networking.ReadOnce()
	// go test.Test()
	// go test.Test()
	// recvChannel := make(chan int)
	// recvErrorChannel := make(chan error)
	// sendChannel := make(chan int)
	// sendErrorChannel := make(chan error)
	// IMPORTANT NOTE on how to use:
	// You can only send messages from master to slave and not from slave to master
	// go tcp.Slave(config.CONNECT_ADDR, config.CONNECT_PORT, recvErrorChannel, recvChannel)
	// go tcp.Master(config.LISTEN_PORT, sendErrorChannel, sendChannel)

	// for {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	sendChannel <- 69
	// 	select {
	// 	case a := <-recvChannel:
	// 		a = a
	// 		fmt.Println("Testmessage:", a)
	// 	case a := <-recvErrorChannel:
	// 		a = a
	// 		fmt.Println("Recieve errormessage:", a)
	// 	case a := <-sendErrorChannel:
	// 		a = a
	// 		fmt.Println("Send errormessage:", a)
	// 	}
	// }

	for {
		// Don't die
	}
}
