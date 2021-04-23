package elevatorManager

import (
	"elevatorproject/config"
	"elevatorproject/orders"
	"elevatorproject/driver-go/elevio"
	"time"	
)

type ElevatorBehaviour int
const (
	EB_Moving	ElevatorBehaviour = iota - 1
	EB_DoorOpen
	EB_Idle
)

func (e ElevatorBehaviour) String() string {
	switch e {
	case EB_Moving: return "EB_Moving"
	case EB_DoorOpen: return "EB_DoorOpen"
	case EB_Idle: return "EB_Idle"
	default: return "Unknown"
	}
}

type Elevator struct {
	Floor     		int
	Requests  		[config.N_FLOORS][config.N_BUTTONS]bool
	Behaviour 		ElevatorBehaviour
	Dirn      		elevio.MotorDirection
	Obstruction 	bool
	ID        		string //unique identifier
	LastChange		time.Time //useful for ordering messages arriving in disorder
}

// Changing from the Requests type to a list with Orders type
func (e Elevator) OrdersFromElevRequests() orders.OrderList{
	subList := make(orders.OrderList, 0)
	for i := 0; i<config.N_FLOORS; i++ {
		for j := 0; j<config.N_BUTTONS; j++ {
			if e.Requests[i][j] {
				o := orders.Order{Orderstate: 		orders.UNASSIGNED,
									OriginElevator: e.ID,
									Destination: 	orders.Floor(i),
									Ordertype:		orders.OrderType(j),
									Timestamp: 		time.Now()}
				if j == 2 {
					o.Orderstate = orders.ASSIGNED
				}

				subList = append(subList, &o)
			}
		}
	}
	return subList
}