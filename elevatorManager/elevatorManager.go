package elevatorManager
<<<<<<< HEAD

import "./elevio"
import "fmt"

func ElevatorManager() {
=======

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"fmt"
	"elevatorproject/config"
)

func elevatorManager() {
	fmt.Println(("Started"))

	if elevio.PollFloorSensor {
		fsm_onInitBetweenFloor()
	}

>>>>>>> 5cac751806bfb8ff7f7b70f66201ad2538b836c2
	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
<<<<<<< HEAD

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)
			if a == numFloors-1 {
				d = elevio.MD_Down
			} else if a == 0 {
				d = elevio.MD_Up
			}
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
=======
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
>>>>>>> 5cac751806bfb8ff7f7b70f66201ad2538b836c2
}
