package elevatorManager

import (
	"elevatorproject/driver-go/elevio"
	"time"

	"elevatorproject/config"
	//"elevatorproject/orders"
)


var elevator = Elevator{Floor: 			-1, 
						Behaviour: 		EB_Idle,
			 			Dirn: 			elevio.MD_Stop,
						Obstruction:	false,
						LastChange:  	time.Now()}

func setAllLights() {
	for floor := 0; floor < config.N_FLOORS; floor++ {
		for btn := elevio.ButtonType(0); btn < 3; btn++ {
			elevio.SetButtonLamp(btn, floor, true)
		}
	}
}

func fsm_onInitBetweenFloor() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = EB_Moving
}

func fsm_onOrdersRecieved(reqFloor int, reqBtn elevio.ButtonType) {
	//ADD printing for debug

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if (elevator.Floor == reqFloor) {
			timer_start()
		} else {
			elevator.Requests[reqFloor][reqBtn] = true
			elevio.SetButtonLamp(reqBtn, reqFloor, true)
		}
	
	case EB_Moving:
		elevator.Requests[reqFloor][reqBtn] = false
		elevio.SetButtonLamp(reqBtn, reqFloor, true)
	
	case EB_Idle:
		if (elevator.Floor == reqFloor) {
			elevio.SetDoorOpenLamp(true)
			timer_start()
			elevator.Behaviour = EB_DoorOpen
		} else {
			elevator.Requests[reqFloor][reqBtn] = true
			elevio.SetButtonLamp(reqBtn, reqFloor, true)
			elevator.Dirn = request_chooseDirection()
			elevio.SetMotorDirection(elevator.Dirn)
			elevator.Behaviour = EB_Moving
		}
	default:

	}
}

func fsm_onFloorArrival(newFloor int) {

	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case EB_Moving:
		if request_shouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = request_clearAtCurrentFloor()
			timer_start()
			//setAllLights()
			elevator.Behaviour = EB_DoorOpen
			//send order complete
		}
	default:
	}

}

func fsm_onDoorTimeout() {	

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if elevator.Obstruction {
			timer_start()
			return
		}
		
		elevator.Dirn = request_chooseDirection()

		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.Dirn)

		if (elevator.Dirn == elevio.MD_Stop) {
			elevator.Behaviour = EB_Idle
		} else {
			elevator.Behaviour = EB_Moving
		}
	default:
	}
}

func fsm_setCabLights(cabLights [config.N_FLOORS]bool) {
	for floor := 0; floor < config.N_FLOORS; floor ++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, cabLights[floor])
	}
}

func fsm_setHallLights(hallLights [config-N_FLOORS][config.N_BUTTONS - 1]bool) {
	for floor := 0; floor < config.N_FLOORS; floor ++ {
		for btn := elevio.ButtonType(0); btn < config.N_BUTTONS - 1; btn ++ {
			elevio.SetButtonLamp(btn, floor, cabLights[floor][btn])
		}
	}
}