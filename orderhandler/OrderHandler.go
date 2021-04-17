package orderHandler

import (
	em "elevatorproject/elevatorManager"
	"elevatorproject/orders"
	"elevatorproject/network/localip"
	"elevatorproject/network/peers"
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

//state
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

					delegateOrders 		chan<- map[string][][]bool,	//master delegates to elevators
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
				} else if newState.Timestamp.After(handler.ElevatorStates[newState.ID].Timestamp) {
					fmt.Println("elevatorState recieved " + newState.ID)
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
				if _, ok := handler.ElevatorStates[elevID]; ok {
					delete(handler.ElevatorStates, elevID)
				}					
			}

			handler.Timestamp = time.Now()
			delegatedOrders := redistributeOrders(handler.AllOrders, handler.ElevatorStates)

			delegateOrders <- delegatedOrders //need to change assigned elevator in Order struct
			backupChan <- handler
			fmt.Println("Orders delegated, backup sent")
			fmt.Println(delegatedOrders)
		}
	}
}

func redistributeOrders(orders orders.OrderList, elevatorStates map[string]em.Elevator) map[string][][]bool {
	input := toHRAInput(orders, elevatorStates)
	hallOrders, err := Assigner(input)

	if err != nil {
		fmt.Printf("Error distributing orders: %e", err)
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
	HallOrder [][2]bool 					`json:"hallRequests"`
	States    map[string]em.HRAElevState   	`json:"states"`
}

func toHRAInput(allOrders orders.OrderList, allStates map[string]em.Elevator) HRAInput {
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


//TODO

//input ouput
//connections