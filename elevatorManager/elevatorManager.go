package elevatorManager

import (
	"elevatorproject/combine"
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"elevatorproject/orders"
	"fmt"
	"time"
)

func ElevatorManager(ID 		string,
					ordersOut 	chan<- orders.Order,
					ordersIn 	<-chan map[string][config.N_FLOORS][config.N_BUTTONS]bool,
					shareState  chan<- Elevator) {

	fmt.Printf("Elevator Manager started: %s\n" + ID)	
	if elevator.Floor== -1 { //vil alltid være tilfellet?
		fsm_onInitBetweenFloor() //rename til bare fsm_init og sende med ID istedet for å sette her?
	}
	elevator.ID = ID

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors  := make(chan int)
	drvObstr   := make(chan bool)
	drvStop    := make(chan bool)
	

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)


	doorTimeout := timer_init()

    for {
		/*
		if timer_timedOut() {
			fmt.Println("DOOR TIME OUT")
			fsm_onDoorTimeout()
		}
		
		if elevator.Obstruction && elevator.Behaviour == EB_DoorOpen{
			timer_start()
		}*/

        select {
			case v := <- drvButtons:
				fmt.Println("Button pressed")
				fsm_onRequestButtonPress(v.Floor, v.Button)
				//make order?
				o := orders.NewOrder(v, elevator.ID)
				ordersOut <- o

				fmt.Println(elevator.Requests)
				
			case f := <- drvFloors:
				fmt.Println("Hit floor")
				//if should stop, send order update?
				fsm_onFloorArrival(f)
		
			case b := <- drvObstr:
				fmt.Printf("Obstruction: %v\n", b)
				elevator.Obstruction = b
				if b && elevator.Behaviour == EB_DoorOpen{
					timer_start()
				}				

			case newOrders := <-ordersIn:
				elevator.Requests = newOrders[elevator.ID]//maybe not overwrite?

				_, cabLights := combine.Demux(newOrders[elevator.ID])				
				hallLights, _ := combine.Demux(newOrders["HallLights"])

				//set lights
				setCabLights(cabLights)
				setHallLights(hallLights)

        	case <- drvStop:
				fmt.Println("stop button not implemented")

			case <-doorTimeout.C:
				fsm_onDoorTimeout()			
        }
		
		elevator.LastChange = time.Now()
		shareState <- elevator
				
		//time.Sleep(config.POLLRATE) //should implement with real timer timing out in select case, so states wont send all the time
    }
}
