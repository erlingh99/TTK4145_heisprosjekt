package Networking

import (
	"fmt"
	"container/list"
	"time"
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	// "../network/conn"
	// "../network/localip"
	// "../network/peers"
	"elevatorproject/network/tcp"
)

var sendIntChannels = list.New()
var sendIntErrChannels = list.New()

var recvIntChannels = list.New()
var recvIntErrChannels = list.New()

func Init() {
	// go SpamChannel(sendNumChan)
	fmt.Println("Init Networking")
	CreateSendIntChannel(config.LISTEN_PORT)
	CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
	go BroadcastMyIp(config.BROADCAST_PORT)
	go ListenForBroadcastedIP(config.BROADCAST_PORT)
}

func SendRepeated() {
	for {
		time.Sleep(time.Second)
		fmt.Println("halla")
		sendIntChannels.Front().Value.(chan int) <- 1234
	}
}

func ReadOnce() {
	fmt.Println(<-recvIntChannels.Front().Value.(chan int))
}

func CreateSendIntChannel(port int) {
	sendChannel := make(chan int)
	errChannel := make(chan error)
	go tcp.Master(port, errChannel, sendChannel)
	sendIntChannels.PushBack(sendChannel)
	sendIntErrChannels.PushBack(errChannel)
	go func () {
		for {
			fmt.Println("Error fra sendchannel:", <-errChannel)
		}
	}()
}

func CreateRecvIntChannel(addr string, port int) {
	recvChannel := make(chan int)
	errChannel := make(chan error)
	go tcp.Slave(addr, port, errChannel, recvChannel)
	recvIntChannels.PushBack(recvChannel)
	recvIntErrChannels.PushBack(errChannel)
	go func () {
		for {
			fmt.Println("Error fra recvchannel:", <-errChannel)
		}
	}()
}

func BroadcastMyIp(port int) {
	channel := make(chan string)
	go bcast.Transmitter(port, channel)
	for {
		time.Sleep(time.Second)
		channel <- "ip goes here"
	}
}

func ListenForBroadcastedIP(port int) {
	channel := make(chan string)
	go bcast.Receiver(port, channel)
	for {
		fmt.Println("Recieved broadcast:", <-channel)
	}
}

func SpamChannel(channel chan int) {
	for {
		channel <- 69
		channel <- 420
	}
}
