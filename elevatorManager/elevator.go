package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"time"	
)

type ElevatorBehaviour int
const (
	EB_Moving	ElevatorBehaviour = iota - 1
	EB_DoorOpen
	EB_Idle
)

func (e ElevatorBehaviour) String() string {
	switch e {
	case EB_Moving: return "EB_Moving"
	case EB_DoorOpen: return "EB_DoorOpen"
	case EB_Idle: return "EB_Idle"
	default: return "Unknown"
	}
}

type Elevator struct {
	Floor     		int
	Requests  		[config.N_FLOORS][config.N_BUTTONS]bool
	Behaviour 		ElevatorBehaviour
	Dirn      		elevio.MotorDirection
	Obstruction 	bool
	ID        		string //unique identifier
	LastChange		time.Time //useful for ordering messages arriving in disorder
}