package orderHandler

import (
	em "elevatorproject/elevatorManager"
	"elevatorproject/orders"
	"elevatorproject/config"
	"elevatorproject/utils"
	"fmt"
	"time"
)

type distributerMode int
const (
	SLAVE distributerMode = iota
	MASTER
)

func (d distributerMode) String() string {
	switch d {
	case MASTER: return "MASTER"
	case SLAVE: return "SLAVE"
	default: return "Unknown"		
	}
}

type DistributerState struct {
	Mode           	distributerMode
	ID             	string
	AllOrders      	orders.OrderList
	ElevatorStates 	map[string]em.Elevator
	Timestamp		time.Time
}

func Distributer(	ID 					string,
	
					orderUpdate 		<-chan orders.Order,		//orders from elevators
					elevatorStateUpdate <-chan em.Elevator,			//states from elevators
					broadcastRx			<-chan string,				//alivemsg from master
					checkpoint 			<-chan DistributerState,	//checkpoint from master
					elevDisconnect 		<-chan string,				//error reaching elevator "elevDisconnect"

					delegateOrders 		chan<- map[string][config.N_FLOORS][config.N_BUTTONS]bool,	//master delegates to elevators
					enableIpBroadcast	chan<- bool,				//enable broadcast of ip (only master broadcasts)
					backupChan 			chan<- DistributerState) {	//send backup to slaves

	handler := DistributerState{
				ElevatorStates: make(map[string]em.Elevator),
				AllOrders:      make(orders.OrderList, 0),
				Mode:           SLAVE,
				ID:				ID,
				Timestamp: 		time.Now()}

	masterTimeoutTimer := time.NewTimer(config.IDLE_CONN_TIMEOUT)

	for {
		switch handler.Mode {
		case SLAVE:
			select {
			
			case <-broadcastRx:
				// fmt.Println("alive message")
				if !masterTimeoutTimer.Stop() {
					select{
					case <-masterTimeoutTimer.C:
					default:
					}
				}
				masterTimeoutTimer.Reset(config.IDLE_CONN_TIMEOUT)
			
			case <-masterTimeoutTimer.C: //master has disconnected
				fmt.Println("masterTimeout")
				handler.Mode = MASTER
				fmt.Println(handler.Mode)
				enableIpBroadcast <- true
				fmt.Println(handler.Mode)
			
			case cp := <-checkpoint:
				if cp.Timestamp.After(handler.Timestamp) {
					// fmt.Println("checkpoint recieved")
					handler.AllOrders = cp.AllOrders
					handler.ElevatorStates = cp.ElevatorStates
					handler.Timestamp = cp.Timestamp
				}

			case <-elevDisconnect:
				fmt.Println("Error with connection to Master")
				handler.Mode = MASTER
				fmt.Println(handler.Mode)
				masterTimeoutTimer.Stop()
				enableIpBroadcast <- true
				fmt.Println(handler.Mode)				
		
			case <-orderUpdate: //Do nothing, master responsibility
			case <-elevatorStateUpdate:
			}

		case MASTER:
			select {
			case newOrder := <-orderUpdate:
				handler.AllOrders.OrderUpdate(&newOrder)
			case newState := <-elevatorStateUpdate:

				if _, exists := handler.ElevatorStates[newState.ID]; !exists {
					fmt.Println("new elevator registered: " + newState.ID)
					handler.ElevatorStates[newState.ID] = newState
					//handle possible new orders
					newOrders := newState.OrdersFromElevRequests()
					handler.AllOrders.OrderUpdateList(newOrders)

				} else if newState.LastChange.After(handler.ElevatorStates[newState.ID].LastChange) {
					//fmt.Println("elevatorState recieved: " + newState.ID)
					handler.ElevatorStates[newState.ID] = newState
				}
			
			case msg := <-broadcastRx: //some other master exist
				if msg == handler.ID {
					continue
				}
				handler.Mode = SLAVE
				fmt.Println(handler.Mode)
				enableIpBroadcast <- false
				masterTimeoutTimer.Stop()
				select {
				case <-masterTimeoutTimer.C:
				default:						
				}									
				masterTimeoutTimer.Reset(config.IDLE_CONN_TIMEOUT)
				fmt.Println(handler.Mode)
			
			case elevID := <-elevDisconnect:
				fmt.Println("Connection error with slave " + elevID)
				delete(handler.ElevatorStates, elevID)	
									
			case <-checkpoint:
				continue //do nothing, slave responsibility
			}


			fmt.Println(handler.AllOrders)
			ordersToAssign, assignedorders, elevsWithProbs := handler.AllOrders.AllUnassignedAndTimedOut()
			fmt.Println(ordersToAssign)
			fmt.Println(assignedorders)

			for _, elevID := range elevsWithProbs {
				delete(handler.ElevatorStates, elevID)
				fmt.Println("Elevator has problems, removed from elevs: " + elevID)
				for _, order := range handler.AllOrders {
					if order.AssignedElevator == elevID {
						order.Orderstate = orders.UNASSIGNED
						order.AssignedElevator = ""
						order.Timestamp = time.Now()
					}
				}
			}

			delegatedOrders, err := redistributeOrders(ordersToAssign, assignedorders, handler.ElevatorStates)
			if err != nil {
				continue
			} 
			handler.Timestamp = time.Now()
			handler.AllOrders.ClearFinishedOrders()		
			delegateOrders <- delegatedOrders
			backupChan <- handler
			//fmt.Println("Orders delegated, backup sent")			
		}
	}
}

func redistributeOrders(Unassigned 		orders.OrderList, 
						assigned 		orders.OrderList,
						elevatorStates 	map[string]em.Elevator) (map[string][config.N_FLOORS][config.N_BUTTONS]bool, error) {

	input := toHRAInput(Unassigned, assigned, elevatorStates)

	fmt.Println(input.States)

	hallOrders, err := Assigner(input)
	Unassigned.MarkAssignedElev(hallOrders)

	
	for _, order := range assigned {
		if order.Ordertype == orders.CAB {
			continue
		} else if _, exist := hallOrders[order.AssignedElevator]; exist {
			temp := hallOrders[order.AssignedElevator]
			temp[int(order.Destination)][int(order.Ordertype)] = true
			hallOrders[order.AssignedElevator] = temp
		}
	}

	sharedLights := [config.N_FLOORS][config.N_BUTTONS]bool {}
	for _, orders := range hallOrders {
		for f := range orders {
			for i, b := range orders[f] {
				if b {
					sharedLights[f][i] = true
				}
			}
		}
	}

	if err != nil {		
		return nil, err
	}

	delegatedOrders := make(map[string][config.N_FLOORS][config.N_BUTTONS]bool)

	for k, elev := range elevatorStates {
		elevOrders := utils.Mux(hallOrders[k], input.States[k].CabRequests)		
		delegatedOrders[elev.ID] = elevOrders		
	}
	
	delegatedOrders["HallLights"] = sharedLights 
	return delegatedOrders, nil
}