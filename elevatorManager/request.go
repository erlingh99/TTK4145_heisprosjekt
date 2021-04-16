package elevatorManager

import (
	"elevatorproject/driver-go/elevio"
 	"elevatorproject/config"
	 "fmt"
)

//Checking for request below current floor, if there is return true
func request_above() bool {
	for f := elevator.Floor + 1; f < config.N_FLOORS; f++ {
		for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
			 if (elevator.Requests[f][btn] != 0) {
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
			 if (elevator.Requests[f][btn] != 0) {
				fmt.Println("Returning true")
				return true
			}
		}
	}
	fmt.Println("Returning false")
	return false
}

//Choosing what direction the elevator should go dempendig on where the request is
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

	case elevio.MD_Down:
		if (request_below()) {
			return elevio.MD_Down
		} else if (request_above()) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	case elevio.MD_Stop:  
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
		(elevator.Requests[elevator.Floor][elevio.BT_HallDown] != 0)|| 
		(elevator.Requests[elevator.Floor][elevio.BT_Cab] != 0)	    ||
		!request_below() )
	case elevio.MD_Up:
		return (
		(elevator.Requests[elevator.Floor][elevio.BT_HallUp] != 0) || 
		(elevator.Requests[elevator.Floor][elevio.BT_Cab] != 0)	 ||
		!request_above())
	case elevio.MD_Stop:
	default:
		return true
	}
	return true
}


//Asuming everyone enters a elevator, even if the elevator is going the wrong way to start with
func request_clearAtCurrentFloor() Elevator {
	for btn := elevio.ButtonType(0); btn < config.N_BUTTONS; btn++ {
		 elevator.Requests[elevator.Floor][btn] = 0
		 elevio.SetButtonLamp(btn, elevator.Floor, false)
	}
	return elevator
}