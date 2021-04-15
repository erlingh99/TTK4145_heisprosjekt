package elevatorManager

import (
	"elevatorproject/elevio"
 	"elevatorproject/config"
)

//Checking for request below current floor, if there is return true
func request_above(Elevator e) bool {
	for (f := e.floor; f < config.N_FLOORS; f++) {
		for (btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++) {
			 if (e.requests[f][btn]) {
				 return true
			 }
		}
	}
}

//Checking for request below current floor, if there is return true
func request_below(Elevator e) bool {
	for (f := 0; f < e.floor; f++) {
		for (btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++) {
			 if (e.requests[f][btn]) {
				 return true
			 }
		}
	}
}

//Choosing what direction the elevator should go dempendig on where the request is
func request_chooseDirection(Elevator e) elevio.MotorDirection{
	switch (e.dirn) {
	case MD_Up:
		if (request_above(e)) {
			return MD_Up
		}
		else if (request_below(e)) {
			return MD_Down
		}
		else {
			return MD_Stop
		}

	case MD_Down:
		if (request_below(e)) {
			return MD_Up
		}
		else if (request_stop(e)) {
			return MD_Down
		}
		else {
			return MD_Stop
		}
	default: return MD_Stop
	}
	case MD_Stop:  
		if (request_below(e)) {
			return MD_Up
		}
		else if (request_stop(e)) {
			return MD_Down
		}
		else {
			return MD_Stop
		}
	default: return MD_Stop
	}
}


func request_shouldStop(Elevator e) bool {
	switch (e.dirn) {
	case MD_Down:
		e.requests[e.floor][BT_HallDown] || 
		e.requests[e.floor][BT_Cab] 		||
		!request_below(e)
	case MD_Up:
		e.requests[e.floor][BT_HallUp] || 
		e.requests[e.floor][BT_Cab]	  ||
		!request_above(e)
	case MD_Stop:
	default:
		return true
	}
}


//Asuming everyone enters a elevator, even if the elevator is going the wrong way to start with
func request_clearAtCurrentFloor(Elevator e) Elevator {
	for (btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++) {
		e.requests[e.floor][btn] = 0
	}
	return e
}