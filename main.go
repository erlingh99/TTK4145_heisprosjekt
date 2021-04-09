package main

import (
	"fmt"

	"./network/tcp"
)

func main() {
	testChannel := make(chan int)
	errorChannel := make(chan error)
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
