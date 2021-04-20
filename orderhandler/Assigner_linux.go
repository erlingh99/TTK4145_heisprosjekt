// +build linux

package orderHandler

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"elevatorproject/config"
)

func Assigner(input HRAInput) (map[string][config.N_FLOORS][2]bool, error) {

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("json.Marshal error: %v\n", err)
		return nil, err
	}

	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner", "-i", string(jsonBytes)).Output()
	if err != nil {
		fmt.Printf("exec.Command error: %v\n", err)
		return nil, err
	}

	
	output := make(map[string][config.N_FLOORS][2]bool)
	err = json.Unmarshal(retvals, &output)
	if err != nil {
		fmt.Printf("json.Unmarshal error: %v\n", err)
		return nil, err
	}

	return output, nil
}