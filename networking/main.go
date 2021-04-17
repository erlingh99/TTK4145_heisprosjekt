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
	// "../network/peers"
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

var MasterAddress net.Addr
var MasterIP string

var AddrRecvLastTime = time.Now()

var iAmMaster = false

// Used for accepting incoming connections
var NextOpenPort = 20003

// Used for connecting new channel to master
var ConnectPort = 20003

func Init(
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
	ordersIn 			chan<- orders.Order,

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

	// Make channels from network:
	receiverErrors := 					make(<-chan error)
	ordersFromNetwork := 				make(<-chan orders.Order)
	stringsFromNetwork :=				make(<-chan string)
	orderhandlerstatesFromNetwork := 	make(<-chan orderhandler.OrderHandlerState)
	orderTablesFromNetwork := 			make(<-chan map[string][][]bool)

	go tcp.Master(config.LISTEN_PORT, listenerErrors, ordersToNetwork, stringsToNetwork, orderhandlerstatesToNetwork, orderTablesToNetwork)
	go tcp.Slave()
	
	// Main loop
	for {
		select {
		case order := <-delegateOrders:
			if iAmMaster {
				// Send order to all elevators on the network
			}
		case ip := <-IPout:
			if iAmMaster {
				// Broadcast ip
			}
		case backup := <-backupChan:
			if iAmMaster {
				// Send backup to all elevators on the network
			}
		case order := <-ordersOut:
			if iAmMaster {
				// Send order to local orderhandler
				orderUpdate<- order
			} else {
				// Send order to master
			}
		}
	}

	// CreateSendIntChannel(config.LISTEN_PORT)
	// CreateRecvIntChannel(config.CONNECT_ADDR, config.CONNECT_PORT)
	// go BroadcastMyIp(config.BROADCAST_PORT)
	// go ListenForBroadcastedIP(config.BROADCAST_PORT)
	// go bcast.AddressReceiver(config.BROADCAST_PORT, addrRecvChannel, connectPortRecvChannel)
	// AddrRecvLastTime = time.Now()
	// go AcceptIncomingConnections()
	// go TimeoutController()
	// go ChannelReader()
}
