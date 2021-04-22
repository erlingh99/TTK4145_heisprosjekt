package elevatorManager

import (
	"elevatorproject/utils"
	"elevatorproject/config"
	"elevatorproject/driver-go/elevio"
	"elevatorproject/orders"
	"fmt"
	"time"
)

func ElevatorManager(ID 		string,					
					ordersIn 	<-chan map[string][config.N_FLOORS][config.N_BUTTONS]bool,
					orderOut 	chan<- orders.Order,
					shareState  chan<- Elevator) {

	fmt.Printf("Elevator Manager started: %s\n", ID)	

	drvButtons := make(chan elevio.ButtonEvent)
	drvFloors  := make(chan int)

	drvObstr   := make(chan bool)
	drvStop    := make(chan bool)	

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)


	timer_init()
	fsm_onInit(ID)
	
	shareStateTicker := time.NewTicker(config.ELEV_SHARE_STATE_INTERVAL)

    for {
		//Main loop checking for inputs on any of the channels
        select {
			//When button is pressed on the elevator
			case v := <- drvButtons:
				fmt.Println("Button pressed")				
				if v.Floor == elevator.Floor && elevator.Behaviour != EB_Moving {
					fsm_openDoor()
					continue
				}

				//make order
				o := orders.NewOrder(v, elevator.ID)
				orderOut <- o
				fmt.Printf("new order: %v\n", o)
			
			//When elevator passes a floor sensor
			case f := <- drvFloors:				
				fmt.Printf("Hit floor %d\n",f)
				fsm_onFloorArrival(f)
				
				//Update master that order is completed if a floor with order is hit
				if elevator.Behaviour == EB_DoorOpen{
					o := orders.Order{
									Orderstate:  	orders.COMPLETED,
									Ordertype:   	orders.CAB,
									Destination: 	orders.Floor(f),
									Timestamp:   	time.Now(),
									OriginElevator:	elevator.ID}
					orderOut <- o
					fmt.Printf("order complete %v\n", o.Destination)
				}
			
				
			//When there is an obstruction
			case b := <- drvObstr:
				fmt.Printf("Obstruction: %v\n", b)
				//If obsturction, the timer will be restared until the obstruction is gone
				elevator.Obstruction = b
				if !b && elevator.Behaviour == EB_DoorOpen{
					timer_start()
				}				
			
			//when there comes a new orders from the master elevator
			case newOrders := <-ordersIn:				
				if newOrders[elevator.ID] != elevator.Requests {
					//fmt.Println(newOrders)
					fmt.Println("new orders recieved")
				}
				elevator.Requests = newOrders[elevator.ID]

				_, cabLights := utils.Demux(newOrders[elevator.ID])				
				hallLights, _ := utils.Demux(newOrders["HallLights"])
				

				//check for recieved order on the floor we are at
				if (elevator.Requests[elevator.Floor][0] || elevator.Requests[elevator.Floor][1]) && elevator.Behaviour != EB_Moving {
					fsm_openDoor()
					o := orders.Order{
						Orderstate:  	orders.COMPLETED,
						Ordertype:   	orders.CAB,
						Destination: 	orders.Floor(elevator.Floor),
						Timestamp:   	time.Now(),
						OriginElevator:	elevator.ID}
					orderOut <- o
					fmt.Printf("order complete %v\n", o.Destination)
				}		
					
				//set lights
				fsm_setCabLights(cabLights)
				fsm_setHallLights(hallLights)

				//start elevator
				fsm_onOrdersRecieved()
				
			
			//When the stopbutton is pressed, this is not implemented
        	case <- drvStop:
				fmt.Println("stop button not implemented")

			//When the doortimer is done
			case <-doorTimer.C:
				fmt.Println("Doortimeout")
				fsm_onDoorTimeout()	
			
			case <- shareStateTicker.C:
				shareState <- elevator				
        }
		elevator.LastChange = time.Now()				
	}
}
