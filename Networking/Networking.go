package Networking

import (
	"fmt"
	"container/list"
	"time"
	"net"
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

var MasterAddress net.Addr

func Init() {
	// go SpamChannel(sendNumChan)
	fmt.Println("Init Networking")
	// CreateSendIntChannel(config.LISTEN_PORT)
	// CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
	// go BroadcastMyIp(config.BROADCAST_PORT)
	// go ListenForBroadcastedIP(config.BROADCAST_PORT)
	go NetworkingMainThread()
}

func NetworkingMainThread() {
	iAmMaster := false
	addrRecvChannel := make(chan net.Addr)
	addrSendChannel := make(chan string)
	go bcast.AddressReceiver(config.BROADCAST_PORT, addrRecvChannel)
	addrRecvLastTime := time.Now()
	for {
		time.Sleep(config.MASTER_BROADCAST_INTERVAL)
		select {
		case addr := <-addrRecvChannel:
			// Long way:
			// tcpAddr, _ := net.ResolveTCPAddr("tcp", addr.String())
			// addrString := tcpAddr.IP.String()
			
			// Short way:
			// addrString := addr.(*net.UDPAddr).IP.String()
			// fmt.Println("Recieved broadcast in Networking main thread:", addrString)
			if !IsItMyAddress(addr.(*net.UDPAddr)) {
				fmt.Println("There is another master at", addr)
				addrRecvLastTime = time.Now()
				MasterAddress = addr
				iAmMaster = false
			}
		default:
			// Just keep going
		}
		
		if !iAmMaster && time.Now().After(addrRecvLastTime.Add(config.MASTER_BROADCAST_LISTEN_TIMEOUT)) {
			// Timeout
			fmt.Println("Timeout. I make myself master")
			iAmMaster = true
			go bcast.Transmitter(config.BROADCAST_PORT, addrSendChannel)
		}

		if iAmMaster {
			message := "I am your master"
			fmt.Printf("Broadcasting message: %s\n", message)
			addrSendChannel <- message
		}	
	}
}

func IsItMyAddress(addr *net.UDPAddr) bool {
	localAddrs, _ := net.InterfaceAddrs()
	for _, localAddr := range localAddrs {
		if addr.IP.String() == localAddr.(*net.IPNet).IP.String() {
			return true
		}
	}
	return false
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
		channel <- "I am your master"
	}
}

func ListenForBroadcastedIP(port int) {
	channel := make(chan net.Addr)
	go bcast.AddressReceiver(port, channel)
	for {
		addr := <-channel
		// Long way:
		// tcpAddr, _ := net.ResolveTCPAddr("tcp", addr.String())
		// addrString := tcpAddr.IP.String()

		// Short way:
		addrString := addr.(*net.UDPAddr).IP.String()
		fmt.Println("Recieved broadcast:", addrString)
		MasterAddress = addr
	}
}

func SpamChannel(channel chan int) {
	for {
		channel <- 69
		channel <- 420
	}
}
