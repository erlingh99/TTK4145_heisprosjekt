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

var addrRecvChannel = make(chan net.Addr)
var addrSendChannel = make(chan int)
var connectPortRecvChannel = make(chan int)

var MasterAddress net.Addr
var MasterIP string

var AddrRecvLastTime = time.Now()

var iAmMaster = false

// Used for accepting incoming connections
var NextOpenPort = 20003

// Used for connecting new channel to master
var ConnectPort = 20003

func Init() {
	// go SpamChannel(sendNumChan)
	fmt.Println("Init Networking")
	// CreateSendIntChannel(config.LISTEN_PORT)
	// CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
	// go BroadcastMyIp(config.BROADCAST_PORT)
	// go ListenForBroadcastedIP(config.BROADCAST_PORT)
	go bcast.AddressReceiver(config.BROADCAST_PORT, addrRecvChannel, connectPortRecvChannel)
	AddrRecvLastTime = time.Now()
	go AcceptIncomingConnections()
	go TimeoutController()
	go ChannelReader()
}

func TimeoutController() {
	for {
		time.Sleep(config.MASTER_BROADCAST_INTERVAL)
		
		if !iAmMaster && time.Now().After(AddrRecvLastTime.Add(config.MASTER_BROADCAST_LISTEN_TIMEOUT)) {
			// Timeout
			fmt.Println("Timeout. I make myself master")
			iAmMaster = true
			go bcast.Transmitter(config.BROADCAST_PORT, addrSendChannel)
		}
	
		if iAmMaster {
			// message := "I am your master"
			message := NextOpenPort
			fmt.Printf("Broadcasting message: %d\n", message)
			addrSendChannel <- message
		}
	}
}

func ChannelReader() {
	for {
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
				AddrRecvLastTime = time.Now()
				MasterAddress = addr
				MasterIP = addr.(*net.UDPAddr).IP.String()
				iAmMaster = false
				channel := make(chan string)
				go ConnectSendChannelToMaster(channel)
				go func() {
					time.Sleep(time.Second)
					channel <- "Halla!"
				}()
			}
		case port := <- connectPortRecvChannel:
			ConnectPort = port
		default:

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

func ConnectSendChannelToMaster(channel interface{}) {
	// sendChannel := make(chan int)
	errChannel := make(chan error)
	go tcp.Transmitter(MasterIP, ConnectPort, errChannel, channel)
	for {
		fmt.Println("Error in ConnectSendChannelToMaster():", <-errChannel)
	}
}

// func RequestConnection(ip string) int {
// 	addrString := fmt.Sprintf("%s:%d", ip, config.REQUEST_CONNECTION_PORT)
// 	conn, err := net.Dial("tcp4", addrString)
// 	var buf [1024]byte
// 	n, e := conn.Read(buf[0:])
// 	if e != nil {
// 		fmt.Printf("RequestConnection(:%d, ...):ReadFrom() failed: \"%+v\"\n", config.REQUEST_CONNECTION_PORT, e)
// 		// errorChan <- err
// 		conn.Close()
// 		return 0
// 	}
// 	// tcp.Transmitter(addr.IP.String(), , )
// }

func AcceptIncomingConnections() {
	for {
		channel := make(chan string)
		errChannel := make(chan error)
		NextOpenPort++
		go tcp.Receiver(NextOpenPort, errChannel, channel)
		select {
		case message := <-channel:
			fmt.Println("Got message:", message)
			go func() {
				for {
					fmt.Println("Got message:", <-channel)
				}
			}()
		}
	}
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
