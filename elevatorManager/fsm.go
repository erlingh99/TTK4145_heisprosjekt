
package elevatorManager

import (
	"elevatorproject/driver-go/elevio"

	"elevatorproject/config"
)


var elevator = Elevator{floor: 		-1, 
						behaviour: 	EB_Idle,
			 			dirn: 		elevio.MD_Stop,
						obstruction:false}

func setAllLights() {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, true)
		}
	}
}

func fsm_onInitBetweenFloor() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.dirn = elevio.MD_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(reqFloor int, reqBtn elevio.ButtonType) {
	//ADD printing for debug

	switch elevator.behaviour {
	case EB_DoorOpen:
		if (elevator.floor == reqFloor) {
			timer_start(config.DOOR_TIMEOUT)
		} else {
			elevator.requests[reqFloor][reqBtn] = 1
			elevio.SetButtonLamp(reqBtn, reqFloor, true)
		}
	
	case EB_Moving:
		elevator.requests[reqFloor][reqBtn] = 1
		elevio.SetButtonLamp(reqBtn, reqFloor, true)
	
	case EB_Idle:
		if (elevator.floor == reqFloor) {
			elevio.SetDoorOpenLamp(true)
			timer_start(config.DOOR_TIMEOUT)
			elevator.behaviour = EB_DoorOpen
		} else {
			elevator.requests[reqFloor][reqBtn] = 1
			elevio.SetButtonLamp(reqBtn, reqFloor, true)
			elevator.dirn = request_chooseDirection()
			elevio.SetMotorDirection(elevator.dirn)
			elevator.behaviour = EB_Moving
		}
	default:

	}
}

func fsm_onFloorArrival(newFloor int) {

	elevator.floor = newFloor
	elevio.SetFloorIndicator(elevator.floor)

	switch elevator.behaviour {
	case EB_Moving:
		if(request_shouldStop()) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = request_clearAtCurrentFloor()
			timer_start(config.DOOR_TIMEOUT)
			//setAllLights()
			elevator.behaviour = EB_DoorOpen
		}
	default:
	}

}

func fsm_onDoorTimeout() {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		elevator.dirn = request_chooseDirection()

		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.dirn)

		if (elevator.dirn == elevio.MD_Stop) {
			elevator.behaviour = EB_Idle
		} else {
			elevator.behaviour = EB_Moving
		}
	default:
	}
}
