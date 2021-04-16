package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	
)

type ElevatorBehaviour int

const (
	EB_Idle     ElevatorBehaviour = 1
	EB_DoorOpen                   = 0
	EB_Moving                     = -1
)

type Elevator struct {
	floor     		int
	requests  		[config.N_FLOORS][config.N_BUTTONS]int
	behaviour 		ElevatorBehaviour
	dirn      		elevio.MotorDirection
	obstruction 	bool
	ID        		string //unique identifier
}

func (e Elevator) ToHRAFormat(cabOrders []bool) (HRAElevState, error) {
	//if e.Available { //aka ikke stopp knappen trykket inn
	//	return HRAElevState{}, fmt.Errorf("Elevator not available")
	//}

	h := HRAElevState{}
	switch e.behaviour {
	case EB_Idle:
		h.Behavior = "idle"
	case EB_DoorOpen:
		h.Behavior = "doorOpen"
	case EB_Moving:
		h.Behavior = "moving"
	}

	switch e.dirn {
	case elevio.MD_Up:
		h.Direction = "up"
	case elevio.MD_Stop:
		h.Direction = "stop"
	case elevio.MD_Down:
		h.Direction = "down"
	}
	h.Floor = e.floor
	h.CabRequests = cabOrders

	return h, nil
}

type HRAElevState struct {
	Behavior    string `json:"behavior"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}
