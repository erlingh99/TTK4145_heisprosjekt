package switcher


import (	
	"elevatorproject/network/localip"
	oh "elevatorproject/orderHandler"
	em "elevatorproject/elevatorManager"
	"elevatorproject/orders"
	"elevatorproject/config"
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

	for {
		_, err := localip.LocalIP()
		if err != nil {
			hasInternet = false
		} else {
			hasInternet = true
		}

		select {
		case msg := <- ordersFromElevatorOut:
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
			if hasInternet {
				ordersToElevatorsOut2 <- msg
			} else {
				ordersToElevatorsIn <- msg
			}
		}
	}
}