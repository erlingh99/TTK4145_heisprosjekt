package switcher


import (	
	"fmt"
	"elevatorproject/network/localip"
	oh "elevatorproject/orderHandler"
	em "elevatorproject/elevatorManager"
	"elevatorproject/orders"
	"elevatorproject/config"
	"time"
)


func Switcher(	//inputs
				ordersFromElevatorOut 		<-chan orders.Order,
				elevStateChangeOut 			<-chan em.Elevator,
				backupOut 					<-chan oh.DistributerState,
				ordersToElevatorsOut 		<-chan map[string][config.N_FLOORS][config.N_BUTTONS]bool,
				//outputs with internet
				ordersFromElevatorOut2 		chan<- orders.Order,
				elevStateChangeOut2 		chan<- em.Elevator,
				backupOut2					chan<- oh.DistributerState,
				ordersToElevatorsOut2		chan<- map[string][config.N_FLOORS][config.N_BUTTONS]bool,
				//outputs without internet
				ordersFromElevatorIn        chan<- orders.Order,
				elevStateChangeIn			chan<- em.Elevator,
				backupIn					chan<- oh.DistributerState,
				ordersToElevatorsIn			chan<- map[string][config.N_FLOORS][config.N_BUTTONS]bool) {

	hasInternet := false

	go func() {
		for {
			ip, err := localip.LocalIP()
			fmt.Println("localip:", ip, err)
			if err != nil {
				hasInternet = false
			} else {
				hasInternet = true
			}
			time.Sleep(config.NET_TIMEOUT)
		}
	}()

	for {


		select {
		case msg := <- ordersFromElevatorOut:
			fmt.Println("hasInternet 1:", hasInternet)

			if hasInternet {
				ordersFromElevatorOut2<-msg
			} else {
				ordersFromElevatorIn<-msg
			}
		case msg := <- elevStateChangeOut:
			if hasInternet {
				elevStateChangeOut2<-msg
			} else {
				elevStateChangeIn<-msg
			}
		case msg := <- backupOut:
			if hasInternet {
				backupOut2 <-msg
			} else {
				backupIn<-msg
			}
		case msg := <- ordersToElevatorsOut:
			fmt.Println("hasInternet 2:", hasInternet)
			if hasInternet {
				ordersToElevatorsOut2 <- msg
			} else {
				fmt.Println("ordToElev:", msg)
				ordersToElevatorsIn <- msg
			}
		}
	}
}