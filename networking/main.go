package networking

import (
	"fmt"
	"container/list"
	"time"
	"net"
	"elevatorproject/config"
	"elevatorproject/network/bcast"
	// "../network/conn"
	// "../network/localip"
	"elevatorproject/network/peers"
	"elevatorproject/network/tcp"
	"elevatorproject/orders"
	"elevatorproject/orderhandler"
)

var sendIntChannels = list.New()
var sendIntErrChannels = list.New()

var recvIntChannels = list.New()
var recvIntErrChannels = list.New()

var addrRecvChannel = make(chan net.Addr)
var addrSendChannel = make(chan int)
var connectPortRecvChannel = make(chan int)

var MasterIP string
var newMasterNotificationCh = make(chan string)

var AddrRecvLastTime = time.Now()

var iAmMaster = false

// Used for accepting incoming connections
var NextOpenPort = 20003

// Used for connecting new channel to master
var ConnectPort = 20003

func Init(
	elevatorID 			string,

	// Channels to orderhandler:
	orderUpdate 		chan<- orders.Order,
	elevatorStateUpdate chan<- Elevator,
	aliveMsg 			chan<- string,
	checkpoint 			chan<- orderhandler.OrderHandlerState,
	connRequest 		chan<- string,
	connError 			chan<- error,
	peerUpdate			chan<- peers.PeerUpdate,

	// Channels from orderhandler:
	delegateOrders 		<-chan map[string][][]bool,
	IPout 				<-chan string,
	backupChan 			<-chan orderhandler.OrderHandlerState,
	
	// Channels to elevatormanager:
	ordersTablesIn		chan<- map[string][][]bool,

	// Channels from elevatormanager:
	ordersOut			<-chan orders.Order
	) {
	fmt.Println("Init Networking")
	
	// Make channels to network:
	listenerErrors := 					make(chan<- error)
	ordersToNetwork := 					make(chan<- orders.Order)
	stringsToNetwork :=					make(chan<- string)
	orderhandlerstatesToNetwork := 		make(chan<- orderhandler.OrderHandlerState)
	orderTablesToNetwork := 			make(chan<- map[string][][]bool)

	go tcp.Master(config.LISTEN_PORT, listenerErrors, ordersToNetwork, stringsToNetwork, orderhandlerstatesToNetwork, orderTablesToNetwork)

	// Declare channels from network:
	var receiverErrors 					<-chan error
	var ordersFromNetwork 				<-chan orders.Order
	var stringsFromNetwork 				<-chan string
	var orderhandlerstatesFromNetwork 	<-chan orderhandler.OrderHandlerState
	var orderTablesFromNetwork 			<-chan map[string][][]bool

	// go tcp.Slave(MasterIP, config.LISTEN_PORT, receiverErrors, ordersFromNetwork, stringsFromNetwork, orderhandlerstatesFromNetwork, orderTablesFromNetwork)

	// Spawn side threads
	go ListenForBroadcastedIP(config.BROADCAST_PORT)
	go TimeoutController()
	
	// Main loop
	for {
		select {
		// Messages from orderhandler:
		case orderTable := <-delegateOrders:
			if iAmMaster {
				// Send orderTable to all elevators on the network
				orderTablesToNetwork<- orderTable
			}
		case ip := <-IPout:
			if iAmMaster {
				// Broadcast ip
				// This is did in timeoutController instead
			}
		case backup := <-backupChan:
			if iAmMaster {
				// Send backup to all elevators on the network
				orderhandlerstatesToNetwork<- backup
			}
		
		// Messages from elevatormanager:
		case order := <-ordersOut:
			if iAmMaster {
				// Send order to local orderhandler
				orderUpdate<- order
			} else {
				// Send order to master
				ordersToNetwork<- order
			}
		
		// Messages from network:
		case order := <-ordersFromNetwork:
			// Order update from an evevatormanager, send to orderhandler
			orderUpdate<- order
		case str := <-stringsFromNetwork:
			// Shouldn't really get any strings, ignore
		case state := <-orderhandlerstatesFromNetwork:
			// This is a backup from master, send to orderhandler
			checkpoint<- state
		case orderTable := <-orderTablesFromNetwork:
			// This is a set of delegated orders, send to elevatormanager
			ordersTablesIn<- orderTable
		
		// New master:
		case MasterIP = <-newMasterNotificationCh:
			// Make new channels from network
			receiverErrors = 					make(<-chan error)
			ordersFromNetwork = 				make(<-chan orders.Order)
			stringsFromNetwork =				make(<-chan string)
			orderhandlerstatesFromNetwork = 	make(<-chan orderhandler.OrderHandlerState)
			orderTablesFromNetwork = 			make(<-chan map[string][][]bool)
	
			// Spawn slave thread
			go tcp.Slave(MasterIP, config.LISTEN_PORT, receiverErrors, ordersFromNetwork, stringsFromNetwork, orderhandlerstatesFromNetwork, orderTablesFromNetwork)
			 

		// No actions on channels:
		default:

			time.Sleep(time.Second)
		}
	}

	// CreateSendIntChannel(config.LISTEN_PORT)
	// CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
	// go BroadcastMyIp(config.BROADCAST_PORT)
	// go bcast.AddressReceiver(config.BROADCAST_PORT, addrRecvChannel, connectPortRecvChannel)
	// AddrRecvLastTime = time.Now()
	// go AcceptIncomingConnections()
	// go ChannelReader()
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
			message := "I am your master"
			// message := NextOpenPort
			fmt.Printf("Broadcasting message: %d\n", message)
			addrSendChannel <- message
		}
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
		if addrString != MasterIP {
			newMasterNotificationCh<- addrString
		}
	}
}
