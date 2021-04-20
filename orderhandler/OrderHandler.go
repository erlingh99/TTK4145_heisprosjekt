package orderHandler

import (
	em "elevatorproject/elevatorManager"
	"elevatorproject/orders"
	"elevatorproject/config"
	"elevatorproject/combine"
	"fmt"
	"time"
)

const (
	IDLE_CONN_TIMEOUT         = 2000  //ms
	ORDER_COMPLETION_MAX_TIME = 10000 //ms, 10 sec
)

type distributerMode int
const (
	SLAVE distributerMode = iota
	MASTER
)

type DistributerState struct {
	Mode           	distributerMode
	ID             	string
	AllOrders      	orders.OrderList
	ElevatorStates 	map[string]em.Elevator
	Timestamp		time.Time
}

func Distributer(	ID 					string,
	
					orderUpdate 		<-chan orders.Order,		//orders from elevators
					elevatorStateUpdate <-chan em.Elevator,			//states, from elevators
					broadcastRx			<-chan string,				//alivemsg from master
					checkpoint 			<-chan DistributerState,	//checkpoint from master
					//newConnection 		<-chan string,				//new Elevator has conneceted (master responsibilty) LET NETWORK DEAL WITH
					elevDisconnect 		<-chan string,				//connection lost (master responsibility)
					//peerUpdate			<-chan peers.PeerUpdate,	

					delegateOrders 		chan<- map[string][config.N_FLOORS][config.N_BUTTONS]bool,	//master delegates to elevators
					enableIpBroadcast	chan<- bool,				//enable broadcast of ip (only master broadcasts)
					backupChan 			chan<- DistributerState) {	//send backup to slaves

	handler := DistributerState{
				ElevatorStates: make(map[string]em.Elevator),
				AllOrders:      make(orders.OrderList, 0),
				Mode:           SLAVE,
				ID:				ID,
				Timestamp: 		time.Now()}

	masterTimeoutTimer := time.NewTimer(IDLE_CONN_TIMEOUT * time.Millisecond)

	for {
		switch handler.Mode {
		case SLAVE:
			select {
			case <-broadcastRx:
				if !masterTimeoutTimer.Stop() {
					<-masterTimeoutTimer.C
				}
				masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)

			case <-masterTimeoutTimer.C: //master has disconnected
				fmt.Println("masterTimeout")
				handler.Mode = MASTER
				enableIpBroadcast <- true

			case cp := <-checkpoint:
				if cp.Timestamp.After(handler.Timestamp) {
					fmt.Println("checkpoint recieved")
					handler.AllOrders = cp.AllOrders
					handler.ElevatorStates = cp.ElevatorStates
				}

			case <-elevDisconnect:
				fmt.Println("Error with connection to Master")
				handler.Mode = MASTER
				masterTimeoutTimer.Stop()
				enableIpBroadcast <- true				
			}

		case MASTER:
			select {
			case newOrder := <-orderUpdate:
				handler.AllOrders.OrderUpdate(newOrder)
			case newState := <-elevatorStateUpdate:

				if _, exists := handler.ElevatorStates[newState.ID]; !exists {
					fmt.Println("new elevator registered: " + newState.ID)
					handler.ElevatorStates[newState.ID] = newState
					//handle possible new orders
					newOrders := OrdersFromElev(newState)
					handler.AllOrders.OrderUpdateList(newOrders)

				} else if newState.LastChange.After(handler.ElevatorStates[newState.ID].LastChange) {
					fmt.Println("elevatorState recieved: " + newState.ID)
					handler.ElevatorStates[newState.ID] = newState
				}

			case msg := <-broadcastRx: //some other master exist
				if msg == handler.ID {
					continue
				}
				handler.Mode = SLAVE
				enableIpBroadcast <- false
				if !masterTimeoutTimer.Stop() {
					<-masterTimeoutTimer.C
				}
				masterTimeoutTimer.Reset(IDLE_CONN_TIMEOUT * time.Millisecond)
				//connect to other master msg=ip
				//continue

			case elevID := <-elevDisconnect:
				fmt.Println("Connection error with slave " + elevID)
				delete(handler.ElevatorStates, elevID)				
			}

			delegatedOrders, err := redistributeOrders(handler.AllOrders, handler.ElevatorStates)
			if err != nil {
				//panic?
				//do what
				//def not send backup and delegate
				continue
			}
			handler.AllOrders.ClearFinishedOrders()
			handler.Timestamp = time.Now()			
			delegateOrders <- delegatedOrders //need to change assigned elevator in Order struct
			backupChan <- handler
			fmt.Println("Orders delegated, backup sent")
			fmt.Println(delegatedOrders)
		}
	}
}

func redistributeOrders(orders orders.OrderList, elevatorStates map[string]em.Elevator) (map[string][config.N_FLOORS][config.N_BUTTONS]bool, error) {
	input := toHRAInput(orders, elevatorStates)

	lights := input.HallOrder

	hallOrders, err := Assigner(input)

	if err != nil {
		fmt.Printf("Error distributing orders: %e", err)
		return nil, err
	}

	delegatedOrders := make(map[string][config.N_FLOORS][config.N_BUTTONS]bool)

	for k, elev := range elevatorStates {
		elevOrders := combine.Mux(hallOrders[k], input.States[k].CabRequests)
		
		delegatedOrders[elev.ID] = elevOrders
		
	}

	//want to send lights on a [4][3]bool chan
	delegatedOrders["HallLights"] = combine.Mux(lights, [config.N_FLOORS]bool{false, false, false, false}) 
	return delegatedOrders, nil
}

func OrdersFromElev(elev em.Elevator) orders.OrderList{

	subList := make(orders.OrderList, 0)
	for i := 0; i<config.N_FLOORS; i++ {
		for j := 0; j<config.N_BUTTONS; j++ {
			if elev.Requests[i][j] {
				o := orders.Order{Orderstate: 		orders.UNASSIGNED,
									OriginElevator: elev.ID,
									Destination: 	orders.Floor(i),
									Timestamp: 		time.Now()}
				switch j {
				case 0: o.Ordertype = orders.HALL_UP
				case 1: o.Ordertype = orders.HALL_DOWN
				case 2: o.Ordertype = orders.CAB
				}
				subList = append(subList, o)
			}
		}
	}
	return subList
}


//TODO

//input ouput
//connections
//order fields, assign values or remove?
//order complete messages