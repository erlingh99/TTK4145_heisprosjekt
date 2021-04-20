package orderHandler

import (
	em "elevatorproject/elevatorManager"
	"elevatorproject/config"
	"elevatorproject/orders"
	"elevatorproject/driver-go/elevio"
)

type HRAInput struct {
	HallOrder [config.N_FLOORS][2]bool 		`json:"hallRequests"`
	States    map[string]HRAElevState   	`json:"states"`
}

func toHRAInput(allOrders orders.OrderList, allStates map[string]em.Elevator) HRAInput {
	input := HRAInput{}

	hallOrders, CabOrders := OrderListToHRAFormat(allOrders)

	input.HallOrder = hallOrders
	for k, elev := range allStates {
		states, err := ElevToHRAFormat(elev, CabOrders[k])
		if err == nil {
			input.States[k] = states
		}
	}
	return input
}

type HRAElevState struct {
	Behavior    string 					`json:"behavior"`
	Floor       int    					`json:"floor"`
	Direction   string 					`json:"direction"`
	CabRequests [config.N_FLOORS]bool 	`json:"cabRequests"`
}

func ElevToHRAFormat(e em.Elevator, cabOrders [config.N_FLOORS]bool) (HRAElevState, error) {

	h := HRAElevState{}
	switch e.Behaviour {
	case em.EB_Idle:		h.Behavior = "idle"
	case em.EB_DoorOpen:	h.Behavior = "doorOpen"
	case em.EB_Moving:		h.Behavior = "moving"
	}

	switch e.Dirn {
	case elevio.MD_Up: 		h.Direction = "up"
	case elevio.MD_Stop:	h.Direction = "stop"
	case elevio.MD_Down:	h.Direction = "down"
	}
	h.Floor = e.Floor
	h.CabRequests = cabOrders

	return h, nil
}


func OrderListToHRAFormat(ol orders.OrderList) ([config.N_FLOORS][2]bool, map[string][config.N_FLOORS]bool) {
	hallOrders := [config.N_FLOORS][2]bool{}
	cabOrders := make(map[string][config.N_FLOORS]bool)

	for _, order := range ol {
		if order.Orderstate == orders.COMPLETED {
			continue
		}

		switch order.Ordertype {
		case orders.CAB:
			cabs := cabOrders[order.AssignedElevatorID]
			cabs[order.Destination] = true
			cabOrders[order.AssignedElevatorID] = cabs
		default:
			hallOrders[order.Destination][order.Ordertype] = true
		}
	}

	return hallOrders, cabOrders
}