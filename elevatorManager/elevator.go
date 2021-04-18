package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"time"
	
)

type ElevatorBehaviour int
const (
	EB_Moving	ElevatorBehaviour = iota -1
	EB_DoorOpen
	EB_Idle
)

type Elevator struct {
	Floor     		int
	Requests  		[config.N_FLOORS][config.N_BUTTONS]int
	Behaviour 		ElevatorBehaviour
	Dirn      		elevio.MotorDirection
	Obstruction 	bool
	ID        		string //unique identifier
	Timestamp		time.Time
}