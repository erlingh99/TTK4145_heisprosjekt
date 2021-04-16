package orderHandler

import (
	. "elevatorproject/elevatorManager"
	"fmt"
	"time"

	"elevatorproject/network/localip"
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
	ID             int
	AllOrders      OrderList
	ElevatorStates map[string]Elevator
	LocalIP        string
}

func OrderHandler(orderUpdate <-chan Order,
	elevatorStateUpdate <-chan Elevator,
	aliveMsg <-chan string,
	checkpoint <-chan OrderHandlerState,
	connRequest <-chan string,
	connError <-chan error,

	delegateOrders chan<- map[string][][]bool,
	IPout chan<- string,
	IDout chan<- int,
	backupChan chan<- OrderHandlerState) {

	id := connectToMaster()
	ip, err := localip.LocalIP()
	if err != nil {
		fmt.Printf("Error no internet connection: ", err)
	}

	handler := OrderHandlerState{
		ID:             id,
		ElevatorStates: make(map[string]Elevator),
		AllOrders:      make(OrderList, 0),
		Mode:           SLAVE,
		LocalIP:        ip}

	masterTimeoutTimer := time.NewTimer(IDLE_CONN_TIMEOUT * time.Millisecond)

	IPoutTicker := time.NewTicker(50 * time.Millisecond)

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
				handler.ID--
				if handler.ID == 0 { //should prob find soemthing else
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
				if msg == handler.LocalIP {
					continue
				}
				handler.Mode = SLAVE
				if !masterTimeoutTimer.Stop() {
					<-masterTimeoutTimer.C
				}
				masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)
				//connect to other master msg=ip
				//continue

			case req := <-connRequest:
				//add to peers, give priority/id
			case err := <-connError:
				//remove from peers
			case <-IPoutTicker.C:
				IPout <- handler.LocalIP
				continue
			}

			delegatedOrders := redistributeOrders(handler.AllOrders, handler.ElevatorStates)

			delegateOrders <- delegatedOrders //need to change assigned elevator in Order struct
			backupChan <- handler
		}
	}
}

func redistributeOrders(orders OrderList, elevatorStates map[string]Elevator) map[string][][]bool {
	input := toHRAInput(orders, elevatorStates)
	hallOrders, err := Distributer(input)

	if err != nil {
		fmt.Printf("Error distributing orders ", err)
		return nil
	}

	delegatedOrders := make(map[string][][]bool)

	for k, elev := range elevatorStates {
		elevOrders := combineOrders(hallOrders[k], input.States[k].CabRequests)
		if elevOrders != nil {
			delegatedOrders[elev.ID] = elevOrders
		}
	}
	return delegatedOrders
}

func combineOrders(hallOrders [][2]bool, cabOrders []bool) [][]bool {
	if hallOrders == nil || cabOrders == nil {
		return nil
	}

	combinedOrders := [][]bool{}
	for f, v := range cabOrders {
		combinedOrders[f] = append(hallOrders[f][:], v)
	}
	return combinedOrders
}

func connectToMaster() error {
	return fmt.Errorf("Not implemented yet")
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
