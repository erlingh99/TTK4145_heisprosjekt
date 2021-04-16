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
	Floor     		int
	Requests  		[config.N_FLOORS][config.N_BUTTONS]int
	Behaviour 		ElevatorBehaviour
	Dirn      		elevio.MotorDirection
	Obstruction 	bool
	ID        		string //unique identifier
}

func (e Elevator) ToHRAFormat(cabOrders []bool) (HRAElevState, error) {
	//if e.Available { //aka ikke stopp knappen trykket inn
	//	return HRAElevState{}, fmt.Errorf("Elevator not available")
	//}

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
	case elevio.MD_Up:
		h.Direction = "up"
	case elevio.MD_Stop:
		h.Direction = "stop"
	case elevio.MD_Down:
		h.Direction = "down"
	}
	h.Floor = e.Floor
	h.CabRequests = cabOrders

	return h, nil
}

type HRAElevState struct {
	Behavior    string `json:"behavior"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}
