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
	floor     int
	requests  [config.N_FLOORS][config.N_BUTTONS]int
	behaviour ElevatorBehaviour
	dirn      elevio.MotorDirection
	obstruction bool
}
