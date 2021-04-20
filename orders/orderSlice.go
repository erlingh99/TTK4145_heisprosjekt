package orders

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
