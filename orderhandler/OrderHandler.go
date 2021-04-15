package orderHandler

import (
	. "elevatorproject/elevatorManager"
	"fmt"
	"time"

	"../network/localip"
	. "..driver-go/elevio"
)

const (
	IDLE_CONN_TIMEOUT         = 2000  //ms
	ORDER_COMPLETION_MAX_TIME = 10000 //ms, 10 sec
)

type handlerMode int

const (
	SLAVE handlerMode = iota
	MASTER
)

//state
type OrderHandlerState struct {
	Mode           handlerMode
	Id             int
	AllOrders      OrderList
	ElevatorStates map[string]Elevator
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

	handler := OrderHandlerState{
		Id:             id,
		ElevatorStates: make(map[string]Elevator),
		AllOrders:      make(OrderList, 0),
		Mode:           SLAVE,
		LocalIP:        ip}

	//outputs
	Orders := make(chan Order)
	IPout := make(chan string)
	IDout := make(chan int)
	backupChan := make(chan OrderHandlerState)

	masterTimeoutTimer := time.NewTimer(IDLE_CONN_TIMEOUT * time.Millisecond)

	for {
		switch handler.Mode {
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
					handler.Mode = MASTER
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
				handler.AllOrders.OrderUpdate(newOrder)
			case newState := <-elevatorStateUpdate:
				for k, e := range handler.ElevatorStates {
					if e.ID == newState.ID {
						handler.ElevatorStates[k] = e
					}
				}
			case msg := <-aliveMsg:
			case req := <-connRequest:
			case err := <-connError:
			}

			//assign orders
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

func redistributeOrders(orders OrderList, elevatorStates []Elevator) {

	Distributer()
}

func masterDuplicate() bool {

}

func sendToPeer(peerID string, o Order) {

}

func connectToMaster() int {

}

func (e Elevator) toHRAFormat() HRAElevState {
	h := HRAElevState{}
	switch e.Behaviour {
	case EB_Idle:
		h.Behavior = "idle"
	case EB_DoorOpen:
		h.Behavior = "doorOpen"
	case EB_Moving:
		h.Behavior = "moving"
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

type HRAInput struct {
	HallOrder [][2]bool
	States    map[string]Elevator
}

func toHRAInput(allOrders OrderList, allStates map[string]Elevator) {
	input := HRAInput{}

	hallOrders, CabOrders := allOrders.OrderListToHRAFormat()

}
