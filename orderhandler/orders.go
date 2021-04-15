package main

import (
	"time"
)

const ORDER_MAX_AGE = 20000 //ms, time before an uncompleted order is reassigned

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
	AssignedElevatorID string
	Id                 int
}

func (o Order) OrderCanBeDeleted() bool {
	return o.Orderstate == COMPLETED
}

func (o Order) CheckForOrderTimeout() bool {
	return time.Now().Sub(o.Timestamp).Milliseconds() > ORDER_MAX_AGE
}

type OrderList List.list

func (ol *OrderList) ClearFinishedOrders() {

}

func (ol *OrderList) FindAllUnassignedAndTimedoutOrders() OrderList {
	subList := OrderList.new()

	for p_order := ol.Front(); p_order != nil; p_order = p_order.Next() {
		if p_order.Orderstate == UNASSIGNED || p_order.CheckForOrderTimeout() {
			subList.PushBack(*p_order)			
		}
	}
	return subList
}

func (ol *OrderList) OrderUpdate(o Order) {
	for p_order := ol.Front(); p_order != nil; p_order = p_order.Next() {
		if p_order.ID == newOrder.ID { //orderUpdate not new
			*p_order = newOrder
			return
		}
	}
	ol.PushBack(newOrder) //new order
}

func (ol *OrderList) 