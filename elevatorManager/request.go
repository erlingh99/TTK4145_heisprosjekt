package elevatorManager

import (
	"elevatorproject/driver-go/elevio"
 	"elevatorproject/config"	
)

//Checking for request above current floor, if there is return true
func request_above() bool {
	for f := elevator.Floor + 1; f < config.N_FLOORS; f++ {
		for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
			if elevator.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

//Checking for request below current floor, if there is return true
func request_below() bool {
	for f := 0; f < elevator.Floor; f++ {
		for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
			if elevator.Requests[f][btn] {				
				return true
			}
		}
	}
	return false
}

//Choosing what direction the elevator should go dependig on where the request is
func request_chooseDirection() elevio.MotorDirection{
	switch (elevator.Dirn) {
	case elevio.MD_Up:
		if (request_above()) {
			return elevio.MD_Up
		} else if (request_below()) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}

	case elevio.MD_Down, elevio.MD_Stop:
		if (request_below()) {
			return elevio.MD_Down
		} else if (request_above()) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	default: return elevio.MD_Stop
	}
}


func request_shouldStop() bool {
	switch (elevator.Dirn) {
	case elevio.MD_Down:
		return (
		elevator.Requests[elevator.Floor][elevio.BT_HallDown] || 
		elevator.Requests[elevator.Floor][elevio.BT_Cab]	  ||
		!request_below() )
	case elevio.MD_Up:
		return (
		elevator.Requests[elevator.Floor][elevio.BT_HallUp]   || 
		elevator.Requests[elevator.Floor][elevio.BT_Cab]	  ||
		!request_above())	 
	default: //elevio.MD_Stop:
		return true
	}
}


//Asuming everyone enters a elevator, even if the elevator is going the wrong way to start with
func request_clearAtCurrentFloor() Elevator {
	for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
		elevator.Requests[elevator.Floor][btn] = false
		elevio.SetButtonLamp(btn, elevator.Floor, false)
	}
	return elevator
}