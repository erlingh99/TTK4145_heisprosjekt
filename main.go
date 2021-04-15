package main

import (
	"fmt"
	"./config"
	"./network/tcp"
)

func main() {
	fmt.Println("Number of floors:", config.N_FLOORS);
	fmt.Println("yo")
	// IMPORTANT NOTE on how to use:
	// You can only send messages from master to slave and not from slave to master
	go tcp.Slave("10.100.23.167", 20001, errorChannel, testChannel)
	// go tcp.Master(20001, errorChannel, testChannel)
	fmt.Println("yo2")

	for {
		// testChannel <- 69
		select {
		case a := <-testChannel:
			a = a
			fmt.Println("Testmessage")
		case a := <-errorChannel:
			a = a
			fmt.Println("Errormessage")
		}
	}
}
