package orders

import (
	"elevatorproject/driver-go/elevio"
	"time"
)

const ORDER_MAX_AGE = 20000 //ms, time before an uncompleted order is reassigned

type OrderState int
const (
	UNASSIGNED OrderState = iota
	ASSIGNED
	COMPLETED
)

type OrderType int
const (
	HALL_UP OrderType = iota
	HALL_DOWN
	CAB
)

type floor int

type Order struct {
	Orderstate         OrderState
	Ordertype          OrderType
	Destination        floor
	Timestamp          time.Time
	AssignedElevatorID string
	OriginElevator     string
}

func NewOrder(be elevio.ButtonEvent, elevID string) Order {
	o := Order{
		Orderstate:  	UNASSIGNED,
		Ordertype:   	OrderType(be.Button),
		Destination: 	floor(be.Floor),
		Timestamp:   	time.Now(),
		OriginElevator:	elevID}
	
	if be.Button == elevio.BT_Cab {
		o.AssignedElevatorID = elevID
		o.Orderstate = ASSIGNED
	}
	return o
}

func (o Order) OrderCanBeDeleted() bool {
	return o.Orderstate == COMPLETED
}

func (o Order) CheckForOrderTimeout() bool {
	return time.Since(o.Timestamp).Milliseconds() > ORDER_MAX_AGE && !o.OrderCanBeDeleted()
}

func (o Order) Equal(o2 Order) bool {
	if o.Ordertype != o2.Ordertype {		
		return false
	}

	if o.Ordertype == CAB {
		if o.AssignedElevatorID != o2.AssignedElevatorID {
			return false
		}
	}

	if o.Destination != o2.Destination {
		return false
	}


	return true
}
