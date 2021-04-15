//INCLUDE TODO INCLUDE TODO INCLUDE
package elevatorManager

import "elevator_io.go"
//!!!!!!!!!!!!!!!!!
//Make when ready

//static ElevOutputDevice


var Elevator elevator

func setAllLights(Elevator es) {
	for (floor := 0; floor < N_FLOORS; floor++) {
		for (btn := elevio.ButtonType(0); btn < 3; btn++) {
			elevio.SetButtonLamp(b, f, false)
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

	switch(elevator.behaviour) {
	case EB_DoorOpen:
		if (elevator.floor == floor) {
			timer_start(config.DOOR_TIMEOUT)
		}
		else {
			elevator.requests[floor][btn] = 1
		}
	case EB_Moving:
		elevator.requests[btn][floor] = 1
	case EB_Idle:
		if (elevator.floor == floor) {
			
		}
	default:

	}
}