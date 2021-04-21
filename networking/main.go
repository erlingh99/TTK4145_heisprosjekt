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
// var addrSendChannel = make(chan int)
var connectPortRecvChannel = make(chan int)

var MasterIP string
var newMasterNotificationCh = make(chan string)

var AddrRecvLastTime = time.Now()

var iAmMaster = false

var SlaveIPs []string
var SlaveLastAliveMsgs []time.Time

var myID string
var enableBroadcast = false

// Used for accepting incoming connections
var NextOpenPort = 20003

// Used for connecting new channel to master
var ConnectPort = 20003

type AliveMessage struct {
	IP string
}

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
	broadCastEnableCh 	<-chan string,
	backupChan 			<-chan orderhandler.OrderHandlerState,
	
	// Channels to elevatormanager:
	ordersTablesIn		chan<- map[string][][]bool,

	// Channels from elevatormanager:
	ordersOut			<-chan orders.Order
	) {
	fmt.Println("Init Networking")

	myID = elevatorID
	
	// Make channels to network:
	ordersToNetwork := 					make(chan<- orders.Order)
	stringsToNetwork :=					make(chan<- string)
	orderhandlerstatesToNetwork := 		make(chan<- orderhandler.OrderHandlerState)
	orderTablesToNetwork := 			make(chan<- map[string][][]bool)
	aliveMessagesToNetwork :=			make(chan<- AliveMessage)

	// Make channels to get back info from network master:
	listenerErrors := 					make(chan<- error)
	newConnCh :=						make(chan<- string)

	go tcp.Master(config.LISTEN_PORT, listenerErrors, ordersToNetwork, stringsToNetwork, orderhandlerstatesToNetwork, orderTablesToNetwork, aliveMessagesToNetwork)

	// Declare channels from network:
	var receiverErrors 					<-chan error
	var ordersFromNetwork 				<-chan orders.Order
	var stringsFromNetwork 				<-chan string
	var orderhandlerstatesFromNetwork 	<-chan orderhandler.OrderHandlerState
	var orderTablesFromNetwork 			<-chan map[string][][]bool
	var aliveMessagesFromNetwork		<-chan AliveMessage

	// go tcp.Slave(MasterIP, config.LISTEN_PORT, receiverErrors, ordersFromNetwork, stringsFromNetwork, orderhandlerstatesFromNetwork, orderTablesFromNetwork)

	// Spawn side threads
	// go ListenForBroadcastedIP(config.BROADCAST_PORT)
	go bcast.Receiver(config.BROADCAST_PORT, aliveMsg)
	go Broadcaster()
	go TimeoutController()
	go SlaveTimeoutController(aliveMessagesFromNetwork, aliveMessagesToNetwork)
	
	// Main loop
	for {
		select {
		// Messages from orderhandler:
		case orderTable := <-delegateOrders:
			if iAmMaster {
				// Send orderTable to all elevators on the network
				orderTablesToNetwork<- orderTable
			}
		case enableBroadcast = <-broadCastEnableCh:
			// Broadcast ip
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

		// Info from network master:
		case newSlaveIP := <-newConnCh:
			if iAmMaster {
				// Add ip to slave list
				SlaveIPs = append(SlaveIPs, newSlaveIP)
				SlaveLastAliveMsgs = append(SlaveLastAliveMsgs, time.Now())
			}
		
		// New master:
		case MasterIP = <-newMasterNotificationCh:
			// Make new channels from network
			receiverErrors = 					make(<-chan error)
			ordersFromNetwork = 				make(<-chan orders.Order)
			stringsFromNetwork =				make(<-chan string)
			orderhandlerstatesFromNetwork = 	make(<-chan orderhandler.OrderHandlerState)
			orderTablesFromNetwork = 			make(<-chan map[string][][]bool)
			aliveMessagesFromNetwork =			make(<-chan AliveMessage)
	
			// Spawn slave thread
			go tcp.Slave(MasterIP, config.LISTEN_PORT, receiverErrors, ordersFromNetwork, stringsFromNetwork, orderhandlerstatesFromNetwork, orderTablesFromNetwork, aliveMessagesFromNetwork)
			 

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

func Broadcaster() {
	bcastCh := make(chan string)
	go bcast.Transmitter(config.BROADCAST_PORT, bcastCh)
	for {
		time.Sleep(config.MASTER_BROADCAST_INTERVAL)

		if enableBroadcast {
			bcastCh<- myID
		}
	}
}

func SlaveTimeoutController(aliveMsgCh <-chan AliveMessage) {
	go func() {
		for {
			aliveMsg := <-aliveMsgCh:
			ip := aliveMsg.IP
			var slaveNr int
			for i, el := range SlaveIPs {
				if el == ip {
					slaveNr = i
					break
				}
			}
			SlaveLastAliveMsgs[slaveNr] = time.Now()
		}
	}()
	for {
		time.Sleep(config.SLAVE_ALIVE_MSG_INTERVAL)
		
		slavesToRemove := make([]int)
		for i, lastAliveMsg := range SlaveLastAliveMsgs {
			if iAmMaster && time.Now().After(lastAliveMsg.Add(config.SLAVE_ALIVE_MSG_LISTEN_TIMEOUT)) {
				// Timeout
				fmt.Println("Timeout. Assume slave has died")
				slavesToRemove = append(slavesToRemove, i)
			}
		}

		// Iterate backwards because we are deleting things as we go
		for i := len(slavesToRemove) - 1; i >= 0; i-- {
			SlaveIPs = append(SlaveIPs[:i], SlaveIPs[i+1:]...)
			SlaveLastAliveMsgs = append(SlaveLastAliveMsgs[:i], SlaveLastAliveMsgs[i+1:]...)
		}
	
		if !iAmMaster {
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
