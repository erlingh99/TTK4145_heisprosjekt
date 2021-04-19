package orders

import "elevatorproject/config"

type OrderList []Order

func (ol *OrderList) ClearFinishedOrders() {
	for i := 0; i < len(*ol); i++ {
		if (*ol)[i].OrderCanBeDeleted() {
			(*ol)[i] = (*ol)[len(*ol)-1] //replace it with the last order since order is unimportant
			*ol = (*ol)[:len(*ol)-1]     //shave off the last element
			i--                          //have to check this index again since it is now a different order
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

func (ol *OrderList) OrderUpdate(o Order) {
	for i, order := range *ol {
		if order.Equal(o) { //orderUpdate not new
			(*ol)[i] = o
			return
		}
	}
	*ol = append(*ol, o) //new order
}

//do all uncompleted orders, or leave assigned and not timed out orders alone?
func (ol OrderList) OrderListToHRAFormat() ([config.N_FLOORS][2]bool, map[string][config.N_FLOORS]bool) {
	hallOrders := [config.N_FLOORS][2]bool{}
	cabOrders := make(map[string][config.N_FLOORS]bool)

	for _, order := range ol {
		if order.Orderstate == COMPLETED {
			continue
		}

		switch order.Ordertype {
		case CAB:
			cabs := cabOrders[order.AssignedElevatorID]
			cabs[order.Destination] = true
			cabOrders[order.AssignedElevatorID] = cabs
		default:
			hallOrders[order.Destination][order.Ordertype] = true
		}
	}

	return hallOrders, cabOrders
}
