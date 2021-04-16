package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"fmt"
	"time"
)

func ElevatorManager() {
	fmt.Println(("Elevator Manager started"))
	if elevator.Floor== -1 {
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


    for {
		if timer_timedOut() {
			fmt.Println("DOOR TIME OUT")
			fsm_onDoorTimeout()
		}
        select {
			case v := <- drvButtons:
				fmt.Println("Button pressed")
				fsm_onRequestButtonPress(v.Floor, v.Button)
				fmt.Println(elevator.Requests)
				
			case f := <- drvFloors:
				fmt.Println("Hit floor")
				fsm_onFloorArrival(f)
		
			case b := <- drvObstr:
				elevator.Obstruction = b
				fmt.Println(b)

        //case b := <- drvStop:

		default:
        }
		time.Sleep(config.POLLRATE)

		if elevator.Obstruction && elevator.Behaviour == EB_DoorOpen{
			timer_start(config.DOOR_TIMEOUT)
		}
		
    }

}
