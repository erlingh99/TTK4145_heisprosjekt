package orderHandler

import (
	. "elevatorproject/elevatorManager"
	"fmt"
	"time"

	"../network/localip"
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
				if handler.Id == 0 { //should prob find soemthing else
					handler.Mode = MASTER
				} else {
					masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)
				}

			case cp := <-checkpoint: //should it always be accepted? ID check
				handler.AllOrders = cp.AllOrders
				handler.ElevatorStates = cp.ElevatorStates
			case <-connError:
				handler.Mode = MASTER
			}

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
				if msg != handler.LocalIP {
					handler.Mode = SLAVE
					if !masterTimeoutTimer.Stop() {
						<-masterTimeoutTimer.C
					}
					masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)
					//connect to other master msg=ip
					//continue
				}
			case req := <-connRequest:
				//add to peers, give priority
			case err := <-connError:
				//remove from peers
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

func redistributeOrders(orders OrderList, elevatorStates map[string]Elevator) {
	input := toHRAInput(orders, elevatorStates)
	output, err := Distributer(input)

	if err != nil {
		fmt.Printf("Error distributing orders ", err)
		return
	}

	for k := range elevatorStates {
		sendToPeer(k, output[k])
	}
}

func sendToPeer(peerID string, orders [][2]bool) {

}

func connectToMaster() int {

}

type HRAInput struct {
	HallOrder [][2]bool
	States    map[string]HRAElevState
}

func toHRAInput(allOrders OrderList, allStates map[string]Elevator) HRAInput {
	input := HRAInput{}

	hallOrders, CabOrders := allOrders.OrderListToHRAFormat()

	input.HallOrder = hallOrders
	for k, elev := range allStates {
		states, err := elev.ToHRAFormat(CabOrders[k])
		if err == nil {
			input.States[k] = states
		}
	}
	return input
}
