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
