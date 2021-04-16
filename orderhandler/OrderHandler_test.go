package orderHandler

import (
	"encoding/json"
	
	"os/exec"
	"testing"

	. "elevatorproject/elevatorManager")


func TestOrderHandler(t *testing.T) {

	
	elev1 := Elevator{Floor: 2,
					Behaviour: EB_Moving,
					Dirn: 0,
					Obstruction: false,
					ID: "heis1"}
	elev2 := Elevator{Floor: 1,
						Behaviour: EB_Idle,
						Dirn: 1,
						Obstruction: false,
						ID: "heis2"}

	st := make(map[string]HRAElevState)

	st[elev1.ID], _ = elev1.ToHRAFormat([]bool{false, false, true, false})
	st[elev2.ID], _ = elev2.ToHRAFormat([]bool{true, false, false, false})

	input := HRAInput{
		HallOrder: [][2]bool{{false, true}, {true, false}, {false, false}, {false, false}},
		States: st,
	}

	jsonBytes, err := json.Marshal(input)
	t.Log(string(jsonBytes))

	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner.exe", "-i", string(jsonBytes)).Output()
	t.Log(err)

	
	output := make(map[string][][2]bool)
	err = json.Unmarshal(retvals, &output)
	t.Log(err)
	

	for k,v := range output {
		t.Log(k)
		t.Log(v)
	}



	//t.Log(output[elev1.ID])
	//t.Log(output[elev2.ID])
}
