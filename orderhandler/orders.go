package orderHandler

import (
	. "elevatorproject/driver-go/elevio"
	"time"
)

const ORDER_MAX_AGE = 20000 //ms, time before an uncompleted order is reassigned
var numberOfOrders int = 0

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
	ID                 int
}

func NewOrder(be ButtonEvent, elevID string) Order {
	o := Order{
		Orderstate:  UNASSIGNED,
		Ordertype:   OrderType(be.Button),
		Destination: floor(be.Floor),
		Timestamp:   time.Now(),
		ID:          numberOfOrders}

	numberOfOrders++
	if be.Button == BT_Cab {
		o.AssignedElevatorID = elevID
		o.Orderstate = ASSIGNED
	}
	return o
}

func (o Order) OrderCanBeDeleted() bool {
	return o.Orderstate == COMPLETED
}

func (o Order) CheckForOrderTimeout() bool {
	return time.Now().Sub(o.Timestamp).Milliseconds() > ORDER_MAX_AGE && !o.OrderCanBeDeleted()
}
