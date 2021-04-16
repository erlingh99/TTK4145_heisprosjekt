package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"fmt"
	"time"
)

func ElevatorManager() {
	fmt.Println(("Elevator Manager started"))
	if elevator.floor== -1 {
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
        case v := <- drvButtons:
			fmt.Println("Button pressed")
			if (prevRequests[v.Floor][v.Button] == 0) {
				fsm_onRequestButtonPress(v.Floor, v.Button)
			}
            prevRequests[v.Floor][v.Button] = 1
            
        case f := <- drvFloors:
            if (f != -1 && f != prevFloorSensor) {
				fmt.Println("Hit floor")
				fsm_onFloorArrival(f)
			}
            
            
        // case a := <- drvObstr:
            
        // case a := <- drvStop:
        }

	
		if (timer_timedOut()) {
			fsm_onDoorTimeout()
		}
		time.Sleep(config.POLLRATE)
    }
}
