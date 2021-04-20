package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"fmt"
	"time"
)


var elevator = Elevator{Floor: 			-1, 
						Behaviour: 		EB_Idle,
			 			Dirn: 			elevio.MD_Stop,
						Obstruction:	false,
						LastChange:  	time.Now()}

func fsm_onInit(ID string) {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.Dirn = elevio.MD_Down
	elevator.Behaviour = EB_Moving
	elevator.ID = ID
}

func fsm_onFloorArrival(newFloor int) {

	elevator.Floor = newFloor
	elevio.SetFloorIndicator(elevator.Floor)

	switch elevator.Behaviour {
	case EB_Moving:
		if request_shouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			//elevator = request_clearAtCurrentFloor()
			timer_start()
			elevator.Behaviour = EB_DoorOpen
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

func fsm_setHallLights(hallLights [config.N_FLOORS][config.N_BUTTONS - 1]bool) {
	for floor := 0; floor < config.N_FLOORS; floor ++ {
		for btn := elevio.ButtonType(0); btn < config.N_BUTTONS - 1; btn ++ {
			elevio.SetButtonLamp(btn, floor, hallLights[floor][btn])
		}
	}
}

func fsm_openDoor() {
	timer_start()
	elevator.Behaviour = EB_DoorOpen
	elevio.SetDoorOpenLamp(true)
}


func fsm_onOrdersRecieved() {
	if elevator.Behaviour != EB_DoorOpen {
		elevator.Dirn = request_chooseDirection()
		elevio.SetMotorDirection(elevator.Dirn)
		if elevator.Dirn != elevio.MD_Stop {
			elevator.Behaviour = EB_Moving
		}		
	}
}