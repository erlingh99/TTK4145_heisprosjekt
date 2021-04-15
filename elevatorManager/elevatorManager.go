package elevatorManager

import (
	"config"
	"elevio"
	"fmt"

	"../config"
)

func elevatorManager() {
	fmt.Println(("Started"))

	if elevio.PollFloorSensor {
		fsm_onInitBetweenFloor()
	}

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors  := make(chan int)
	drvObstr   := make(chan bool)
	drvStop    := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	var prevRequests [config.N_FLOORS][config.N_BUTTONS]int
	var prevFloorSensor int
    for {
        select {
        case v := <- drv_buttons:
			fmt.Println("Button pressed")
			if (a != prev[v.Floor][v.Button]) {
				fsm_onRequestButtonPress(v.Floor, v.Button)
			}
            prev[v.Floor][v.Button] = 1
            
        case f := <- drv_floors:
            if (f != -1 && f != prevFloorSensor) {
				fsm_onFloorArrival(f)
				fmt.Println("Hit floor")
			}
            
            
        case a := <- drv_obstr:
            
        case a := <- drv_stop:
        }

	
		if (timer_timedOut()) {
			fsm_onDoorTimeout()
		}
    }
}
