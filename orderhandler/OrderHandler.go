package main

import (
	//"../elevatorManager"
	"../network/bcast"
	"../network/localip"
	"container/list"
	"encoding/json"
	"fmt"
	"time"
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
	ElevatorStates []Elevator
	LocalIP        string
}

func OrderHandler(orderUpdate <-chan Order,
	elevatorStateUpdate <-chan Elevator,
	aliveMsg <-chan string,
	checkpoint <-chan OrderHandlerState,
	connRequest <-chan string,
	connError <-chan error) {

	id := connectToMaster()
	ip, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Error no internet connection: ", err)
	}

	handler := OrderHandler{
		Id:             id,
		ElevatorStates: make(map[string]Elevator),
		AllOrders:      list.New(),
		State:          SLAVE,
		LocalIP:        ip}

	//outputs
	Orders := make(chan Order)
	IPout := make(chan string)
	IDout := make(chan int)
	backupChan := make(chan OrderHandlerState)

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

			case cp := <-checkpoint:
				handler.AllOrders = cp.AllOrders
				handler.ElevatorStates = cp.ElevatorStates
			case <-connError:
				//cannot connect to master
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
			redistributeOrders(handler.AllOrders, handler.ElevatorStates)

			//check for master duplicate
			if masterDuplicate() {
				handler.State = SLAVE
				connectToMaster()
				//reset timer?
			}
		}
	}
}

func redistributeOrders(orders list.List, elevatorStates []Elevator) {

	Distributer()
}

func masterDuplicate() bool {

}

func sendToPeer(peerID string, o Order) {

}

func connectToMaster() int {
	bcast.Receiver(PORT)
}

func (e Elevator) toHRAFormat() HRAElevState {
	h := HRAElevState{}
	switch e.Behavior {
	case EB_IDLE:
		h.Behavior = "idle"
	case EB_DoorOpen:
		h.behavior = "doorOpen"
	case EB_Moving:
		h.behavior = "moving"
	}

	switch e.Dirn {
	case MD_Up:
		h.Direction = "up"
	case MD_Stop:
		h.Direction = "stop"
	case MD_Down:
		h.Direction = "down"
	}
	h.Floor = e.Floor
	h.CabRequests = make([]bool, e.NumFloors)
	for f, reqs := range e.Requests {
		h.CabRequests[f] = reqs[BT_Cab]
	}
	return h
}

type HRAElevState struct {
	Behavior    string `json:"behavior"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}
