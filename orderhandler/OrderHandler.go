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
					//broadcastRx			<-chan string,				//alivemsg from master
					checkpoint 			<-chan DistributerState,	//checkpoint from master
					elevDisconnect 		<-chan string,				//error reaching elevator "elevDisconnect"

					delegateOrders 		chan<- map[string][config.N_FLOORS][config.N_BUTTONS]bool,	//master delegates to elevators
					//enableIpBroadcast	chan<- bool,				//enable broadcast of ip (only master broadcasts)
					backupChan 			chan<- DistributerState) {	//send backup to slaves

	handler := DistributerState{
				ElevatorStates: make(map[string]em.Elevator),
				AllOrders:      make(orders.OrderList, 0),
				Mode:           SLAVE,
				ID:				ID,
				Timestamp: 		time.Now()}

	masterTimeoutTimer := time.NewTimer(config.IDLE_CONN_TIMEOUT)

	for {
		fmt.Println(handler.Mode)
		switch handler.Mode {
		case SLAVE:
			select {
			/*
			case <-broadcastRx:
				fmt.Println("alive message")
				if !masterTimeoutTimer.Stop() {
					<-masterTimeoutTimer.C
				}
				masterTimeoutTimer.Reset(config.IDLE_CONN_TIMEOUT)
			*/
			case <-masterTimeoutTimer.C: //master has disconnected
				fmt.Println("masterTimeout")
				handler.Mode = MASTER
				//enableIpBroadcast <- true
			
			case cp := <-checkpoint:
				if cp.Timestamp.After(handler.Timestamp) {
					fmt.Println("checkpoint recieved")
					handler.AllOrders = cp.AllOrders
					handler.ElevatorStates = cp.ElevatorStates
					handler.Timestamp =cp.Timestamp
				}

			case <-elevDisconnect:
				fmt.Println("Error with connection to Master")
				handler.Mode = MASTER
				//masterTimeoutTimer.Stop()
				//enableIpBroadcast <- true				
		
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
			/*
			case msg := <-broadcastRx: //some other master exist
				if msg == handler.ID {
					continue
				}
				handler.Mode = SLAVE
				enableIpBroadcast <- false
				masterTimeoutTimer.Stop()
				select {
				case <-masterTimeoutTimer.C:
				default:						
				}									
				masterTimeoutTimer.Reset(config.IDLE_CONN_TIMEOUT)
			*/
			case elevID := <-elevDisconnect:
				fmt.Println("Connection error with slave " + elevID)
				delete(handler.ElevatorStates, elevID)	
									
			case <-checkpoint:
				continue //do nothing, slave responsibility
			}

			delegatedOrders, err := redistributeOrders(handler.AllOrders, handler.ElevatorStates)
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

func redistributeOrders(orders orders.OrderList, elevatorStates map[string]em.Elevator) (map[string][config.N_FLOORS][config.N_BUTTONS]bool, error) {
	input := toHRAInput(orders, elevatorStates)
	sharedLights := input.HallOrder

	hallOrders, err := Assigner(input)

	if err != nil {		
		return nil, err
	}

	delegatedOrders := make(map[string][config.N_FLOORS][config.N_BUTTONS]bool)

	for k, elev := range elevatorStates {
		elevOrders := utils.Mux(hallOrders[k], input.States[k].CabRequests)		
		delegatedOrders[elev.ID] = elevOrders		
	}

	//want to send lights on a [4][3]bool chan
	delegatedOrders["HallLights"] = utils.Mux(sharedLights, [config.N_FLOORS]bool{false, false, false, false}) 
	return delegatedOrders, nil
}