package orderHandler

import (
	"testing"

	"elevatorproject/driver-go/elevio"
	. "elevatorproject/elevatorManager"
	"elevatorproject/orders"
)


func TestOrderHandler(t *testing.T) {

	
	elev1 := Elevator{	Floor: 2,
						Behaviour: EB_Moving,
						Dirn: 0,
						Obstruction: false,
						ID: "heis1"}

	elev2 := Elevator{	Floor: 1,
						Behaviour: EB_Idle,
						Dirn: 1,
						Obstruction: false,
						ID: "heis2"}


	o1 := orders.NewOrder(elevio.ButtonEvent{Floor: 1, Button: elevio.BT_Cab}, "heis1")
	o2 := orders.NewOrder(elevio.ButtonEvent{Floor: 3, Button: elevio.BT_Cab}, "heis2")
	o3 := orders.NewOrder(elevio.ButtonEvent{Floor: 0, Button: elevio.BT_HallUp}, "heis1")
	o4 := orders.NewOrder(elevio.ButtonEvent{Floor: 2, Button: elevio.BT_HallDown}, "heis2")
	o5 := orders.NewOrder(elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallDown}, "heis1")
	//t.Log("orders created")
	ol := orders.OrderList{o1,o2,o3,o4}
	//t.Log("list created")
	ol.OrderUpdate(o5)
	//t.Log("list appended")

	
	hall, cab := OrderListToHRAFormat(ol)

	//t.Log("hra created")

	t.Log(cab)

	st := make(map[string]HRAElevState)

	st[elev1.ID], _ = ElevToHRAFormat(elev1, cab[elev1.ID])
	st[elev2.ID], _ = ElevToHRAFormat(elev2, cab[elev2.ID])

	input := HRAInput{
		HallOrder: hall,
		States: st,
	}

	output, _ := Assigner(input)

	//t.Log(err)	
	t.Log(output)

	o3.Orderstate = orders.COMPLETED
	ol.OrderUpdate(o3)

	hall, cab = OrderListToHRAFormat(ol)

	//t.Log("hra2 created")

	t.Log(cab)

	input = HRAInput{
		HallOrder: hall,
		States: st,
	}

	output, _ = Assigner(input)
	//t.Log(err)

	t.Log(output)
}
