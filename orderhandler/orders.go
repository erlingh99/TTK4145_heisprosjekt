package main

import (
	"time"
)

const ORDER_MAX_AGE = 20000 //ms, how long to keep a record of old orders

type floor int

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

type Order struct {
	Orderstate         OrderState
	Ordertype          OrderType
	Destination        floor
	Origin             floor
	Timestamp          time.Time
	AssignedElevatorID int
	Id                 int
}

func (o Order) OrderCanBeDeleted() bool {
	return o.Orderstate == COMPLETED && o.CheckForOrderTimeout()
}

func (o Order) CheckForOrderTimeout() bool {
	return time.Now().Sub(o.Timestamp).Milliseconds() > ORDER_MAX_AGE
}
