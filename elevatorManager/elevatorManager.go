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
					orderOut 	chan<- orders.Order,
					ordersIn 	<-chan map[string][config.N_FLOORS][config.N_BUTTONS]bool,
					shareState  chan<- Elevator) {

	fmt.Printf("Elevator Manager started: %s\n" + ID)	
	fsm_onInitBetweenFloor() //rename til bare fsm_init og sende med ID istedet for å sette her?
	elevator.ID = ID

	drvButtons := make(<-chan elevio.ButtonEvent)
	drvFloors  := make(<-chan int)
	drvObstr   := make(<-chan bool)
	drvStop    := make(<-chan bool)
	

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)


	doorTimeout := timer_init()

    for {
		//Main loop checking for inputs on any of the channels
        select {
			//When button is pressed on the elevator
			case v := <- drvButtons:
				fmt.Println("Button pressed")
				//make order
				if v.Floor == elevator.Floor && elevator.Behaviour != EB_Moving {
					continue
				}

				o := orders.NewOrder(v, elevator.ID)
				orderOut <- o
			
			//When elevator passes a floor sensor
			case f := <- drvFloors:
				fmt.Println("Hit floor")

				//Update master that order is completed if a floor with order is hit
				if request_shouldStop() {
					o := Order{
								Orderstate:  	orders.COMPLETED,
								Ordertype:   	orders.BT_Cab,
								Destination: 	f,
								Timestamp:   	time.Now(),
								OriginElevator:	elevator.ID}
					orderOut <- o	
				}
				
				fsm_onFloorArrival(f)
				
			//When there is an obstruction
			case b := <- drvObstr:
				fmt.Printf("Obstruction: %v\n", b)
				//If obsturction, the timer will be restared until the obstruction os gone
				elevator.Obstruction = b
				if b && elevator.Behaviour == EB_DoorOpen{
					timer_start()
				}				
			
			//when there comes a new orders from the master elevator
			case newOrders := <-ordersIn:
				elevator.Requests = newOrders[elevator.ID]//maybe not overwrite?

				_, cabLights := combine.Demux(newOrders[elevator.ID])				
				hallLights, _ := combine.Demux(newOrders["HallLights"])

				//set lights
				fsm_setCabLights(cabLights)
				fsm_setHallLights(hallLights)

				//start elevator
				fsm_onOrdersRecieved()
			
			//When the stopbutton is pressed, this is not implemented
        	case <- drvStop:
				fmt.Println("stop button not implemented")

			//When the doortimer is done
			case <-doorTimeout.C:
				fsm_onDoorTimeout()			
        }
		
		elevator.LastChange = time.Now()
		shareState <- elevator
				
	}
}
