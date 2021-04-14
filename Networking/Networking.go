package Networking

import (
	"fmt"
	"container/list"
	"../config"
	// "../network/bcast"
	// "../network/conn"
	// "../network/localip"
	// "../network/peers"
	"../network/tcp"
)

var sendIntChannels = list.New()
var recvIntChannels = list.New()

func Init() {
	// go SpamChannel(sendNumChan)
	fmt.Println("Init Networking")
	CreateSendIntChannel(config.LISTEN_PORT)
	CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
}

func SendOnce() {
	recvIntChannels.Front().Value.(chan int) <- 1234
}

func ReadOnce() {
	fmt.Println(<-recvIntChannels.Front().Value.(chan int))
}

func CreateSendIntChannel(port int) {
	sendChannel := make(chan int)
	errChannel := make(chan error)
	go tcp.Master(port, errChannel, sendChannel)
	sendIntChannels.PushBack(sendChannel)
}

func CreateRecvIntChannel(addr string, port int) {
	recvChannel := make(chan int)
	errChannel := make(chan error)
	go tcp.Slave(addr, port, errChannel, recvChannel)
	recvIntChannels.PushBack(recvChannel)
}

func SpamChannel(channel chan int) {
	for {
		channel <- 69
		channel <- 420
	}
}
