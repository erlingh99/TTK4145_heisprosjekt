package orderHandler

import (
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
	ID                 int
}

func (o Order) OrderCanBeDeleted() bool {
	return o.Orderstate == COMPLETED
}

func (o Order) CheckForOrderTimeout() bool {
	return time.Now().Sub(o.Timestamp).Milliseconds() > ORDER_MAX_AGE && !o.OrderCanBeDeleted()
}

type OrderList []Order

func (ol OrderList) ClearFinishedOrders() {
	for i := 0; i < len(ol); i++ {
		if ol[i].OrderCanBeDeleted() {
			ol[i] = ol[len(ol)-1] //replace it with the last order since order is unimportant
			ol = ol[:len(ol)-1]   //shave off the last element
			i--                   //have to check this index again since it is now a different order
		}
	}
}

func (ol OrderList) FindAllUnassignedAndTimedoutOrders() OrderList {
	subList := make([]Order, 0)

	for _, order := range ol {
		if order.Orderstate == UNASSIGNED || order.CheckForOrderTimeout() {
			subList = append(subList, order)
		}
	}
	return subList
}

func (ol OrderList) OrderUpdate(o Order) {
	for i, order := range ol {
		if order.ID == o.ID { //orderUpdate not new
			ol[i] = o
			return
		}
	}
	ol = append(ol, o) //new order
}

//do all uncompleted orders, or leave assigned and not timed out orders alone?
func (ol OrderList) OrderListToHRAFormat() ([][2]bool, map[string][]bool) {
	hallOrders := [][2]bool{}
	cabOrders := make(map[string][]bool)

	for _, order := range ol {
		if order.Orderstate == COMPLETED {
			continue
		}

		switch order.Ordertype {
		case CAB:
			cabOrders[order.AssignedElevatorID][order.Destination] = true
		default:
			hallOrders[order.Destination][order.Ordertype] = true
		}
	}

	return hallOrders, cabOrders
}
