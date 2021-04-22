package orders

import "elevatorproject/config"

type OrderList []*Order

func (ol *OrderList) ClearFinishedOrders() {
	for i := 0; i < len(*ol); i++ {
		if (*ol)[i].OrderCanBeDeleted() {
			(*ol)[i] = (*ol)[len(*ol)-1] //replace it with the last order since order is unimportant
			*ol = (*ol)[:len(*ol)-1]     //shave off the last element
			i--                          //have to check this index again since it is now a different order
		}
	}
}

func (ol *OrderList) OrderUpdate(o *Order) {	
	if o.Orderstate == COMPLETED {
		ol.clearOrdersAtFloor(o.Destination, o.OriginElevator)
		return
	}

	for i, order := range *ol {		
		if order.Equal(*o) { //orderUpdate not new
			(*ol)[i] = o
			return
		}
	}

	*ol = append(*ol, o)
}

func (ol *OrderList) OrderUpdateList(ol2 OrderList) {
	for _, v := range ol2 {
		ol.OrderUpdate(v)
	}
}

func (ol OrderList)clearOrdersAtFloor(f Floor, elevID string) {
	for _, o := range ol {
		if o.Destination == f && o.Ordertype == CAB && o.OriginElevator == elevID {
			o.Orderstate = COMPLETED
			//ol.OrderUpdate(o)				
		} else if o.Destination == f && o.Ordertype != CAB {
			o.Orderstate = COMPLETED	
			//ol.OrderUpdate(o)		
		}		
	}
}

func (ol OrderList) AllUnassignedAndTimedOut() (OrderList, OrderList, []string) {
	olUnassigned := make(OrderList, 0)
	olAsssigned := make(OrderList, 0)

	timedOutElevs := make([]string,0)

	for _, o := range ol {
		b := o.CheckForOrderTimeout()
		if o.Orderstate == UNASSIGNED || b {
			olUnassigned = append(olUnassigned, o)		
		} else if o.Orderstate == ASSIGNED {
			olAsssigned = append(olAsssigned, o)
		}

		if b {
			timedOutElevs = append(timedOutElevs, o.AssignedElevator)
			o.Orderstate = UNASSIGNED
		}
	}
	return olUnassigned, olAsssigned, timedOutElevs
}

func (ol OrderList) MarkAssignedElev(assignedOrders map[string][config.N_FLOORS][2]bool) {
	for elevID, orders := range assignedOrders {
		for f := range orders {
			for dir, b := range orders[f] {
				if b {
					o := ol.findHallOrder(f, dir)
					if o != nil {					
						o.Orderstate = ASSIGNED
						o.AssignedElevator = elevID
					}
				}
			}
		}
	}
}

func (ol OrderList) findHallOrder(floor int, dir int) *Order{
	for _, o := range ol {
		if o.Ordertype == OrderType(dir) && o.Destination == Floor(floor) {
			return o
		}
	}
	return nil
}
