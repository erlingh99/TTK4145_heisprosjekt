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
	Requests  [config.N_FLOORS][config.N_BUTTONS]int
	Behaviour ElevatorBehaviour
	Dirn      elevio.MotorDirection
	ID        string //unique identifier
}
