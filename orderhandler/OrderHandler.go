package main

import (
	"container/list"
	"time"

	"../network/bcast"
	"../network/localip"
)

const (
	IDLE_CONN_TIMEOUT         = 2000  //ms
	ORDER_COMPLETION_MAX_TIME = 10000 //ms, 10 sec
)

type handlerState int

const (
	SLAVE handlerState = iota
	MASTER
)

//state
type OrderHandlerState struct {
	State          handlerState
	Id             int
	AllOrders      list.List
	ElevatorStates []elevator
	LocalIP        string
}

func OrderHandler(orderUpdate <-chan Order,
	elevatorStateUpdate <-chan Elevator,
	aliveMsg <-chan string,
	checkpointIn <-chan OrderHandlerState,
	connRequest <-chan string,
	connError <-chan error) {

	id := connectToMaster()
	ip, err := localip.LocalIP()
	if err != nil {
		//do  something
	}

	handler := OrderHandler{
		Id:             id,
		ElevatorStates: make(map[string]Elevator),
		AllOrders:      list.New(),
		State:          SLAVE,
		LocalIP:        ip}

	//outputs
	Orders := make(chan<- Order)
	IPout := make(chan<- string)
	IDout := make(chan<- int)
	backupChan := make(chan<- OrderHandlerState)

	masterTimeoutTimer := time.NewTimer(IDLE_CONN_TIMEOUT * time.Millisecond)

	for {
		switch handler.State {
		case SLAVE:
			select {
			case <-aliveMsg:
				if !masterTimeoutTimer.Stop() {
					<-masterTimeoutTimer.C
				}
				masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)

			case <-masterTimeoutTimer.C: //master has disconnected
				handler.Id--
				if handler.Id == 0 {
					handler.State = MASTER
				} else {
					masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)
				}

			case cp := <-checkpointIn:
				handler.AllOrders = cp.AllOrders
				handler.ElevatorStates = cp.ElevatorStates
			case <-connError:
			}

			//do slave stuff on change

		case MASTER:
			select {
			case newOrder := <-orderUpdate:
			case newState := <-elevatorStateUpdate:
			case msg := <-aliveMsg:
			case req := <-connRequest:
			case err := <-connError:
			}

			//calculate order
			redistributeOrders(handler.AllOrders, peers)

			//check for master duplicate
			if masterDuplicate() {
				handler.State = SLAVE
				connectToMaster()
				//reset timer?
			}
		}
	}
}

func redistributeOrders() {
	Distributer()
}

func masterDuplicate() bool {

}

func sendToPeer(peerID string, o Order) {

}

func connectToMaster() int {
	bcast.Receiver(PORT)
}
