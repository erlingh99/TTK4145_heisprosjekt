package elevatorManager

import (
	"../elevio"
 	"../config"
)

//Checking for request below current floor, if there is return true
func request_above(Elevator e) {
	for (f := e.floor; f < N_FLOORS; f++) {
		for (btn := elevio.ButtonType(0); btn < N_BUTTONS; btn++) {
			 if (e.request[f][btn]) {
				 return true
			 }
		}
	}
}

//Checking for request below current floor, if there is return true
func request_below(Elevator e) {
	for (f := 0; f < e.floor; f++) {
		for (btn := elevio.ButtonType(0); btn < N_BUTTONS; btn++) {
			 if (e.request[f][btn]) {
				 return true
			 }
		}
	}
}

//Choosing what direction the elevator should go dempendig on where the request is
func request_chooseDirection(Elevator e) {
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

	case MD_Down: //Since case for MC_Down and MD_Stop are equal they use the same check
	case MD_Stop:  //What order MD_Stop is checked in does not matter, therefore using same as MD_Down to save a case
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


func request_shouldStop(Elevator e) {
	switch (e.dirn) {
	case MD_Down:
		e.request[e.floor][BT_HallDown] || 
		e.request[e.floor][BT_Cab] 		||
		!request_below(e)
	case MD_Up:
		e.request[e.floor][BT_HallUp] || 
		e.request[e.floor][BT_Cab]	  ||
		!request_above(e)
	case MD_Stop:
	default:
		return true
	}
}


func request_clearAtCurrentFloor(Elevator e) {
	
}