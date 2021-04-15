//INCLUDE TODO INCLUDE TODO INCLUDE
package elevatorManager

import (
	"elevatorproject/driver-go/elevio"

	"elevatorproject/config"
)

var elevator Elevator

func setAllLights(Elevator es) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(b, f, true)
		}
	}
}

func fsm_onInitBetweenFloor() {
	SetMotorDirection(MD_Down)
	elevator.dirn = MD_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(int floor, ButtonType btn) {
	//ADD printing for debug

	switch elevator.behaviour {
	case EB_DoorOpen:
		if elevator.floor == floor {
			timer_start(config.DOOR_TIMEOUT)
		} else {
			elevator.requests[floor][btn] = 1
		}
	case EB_Moving:
		elevator.requests[floor][btn] = 1
	case EB_Idle:
		if elevator.floor == floor {
			elevio.SetDoorOpenLamp(true)
			timer_start(config.DOOR_TIMEOUT)
			elevator.behaviour = EB_DoorOpen
		} else {
			elevator.requests[floor][btn] = 1
			elevator.dirn = request_chooseDirection(elevator)
			elevio.MotorDirection(elevator.dirn)
			elevator.behaviour = EB_Moving
		}
	default:

	}
	setAllLights(elevator)

}

func fsm_onFloorArrival(int newFloor) {

	elevator.floor = newFloor
	elevio.SetFloorIndicator(elevator.floor)

	switch elevator.behaviour {
	case EB_Moving:
		if request_shouldStop(elevator) {
			elevio.SetMotorDirection(MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = request_clearAtCurrentFloor(elevator)
			timer_start(config.DOOR_TIMEOUT)
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	default:
	}

}

func fsm_onDoorTimeout() {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		elevator.dirn = request_chooseDirection(elevator)

		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.Dirn)

		if elevator.Dirn == elevio.MD_Stop {
			elevator.Behaviour = EB_Idle
		} else {
			elevator.Behaviour = EB_Moving
		}
	default:
	}
}
