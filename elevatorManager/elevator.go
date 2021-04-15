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
	Floor     int
	Requests  [config.N_FLOORS][config.N_BUTTONS]bool
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
	ID        string //unique identifier
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
	case elevio.MD_Up:
		h.Direction = "up"
	case elevio.MD_Stop:
		h.Direction = "stop"
	case elevio.MD_Down:
		h.Direction = "down"
	}
	h.Floor = e.Floor
	h.CabRequests = make([]bool, config.N_FLOORS)
	for f, reqs := range e.Requests {
		h.CabRequests[f] = reqs[elevio.BT_Cab]
	}
	return h
}

type HRAElevState struct {
	Behavior    string `json:"behavior"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}
