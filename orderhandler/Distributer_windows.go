package orderHandler

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func Distributer(input HRAInput) (map[string][][2]bool, error) {

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("json.Marshal error: ", err)
		return nil, err
	}

	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner.exe", "-i", string(jsonBytes)).Output()
	if err != nil {
		fmt.Printf("exec.Command error: ", err)
		return nil, err
	}

	output := make(map[string][][2]bool)
	err := json.Unmarshal(retvals, &output)
	if err != nil {
		fmt.Printf("json.Unmarshal error: ", err)
		return nil, err
	}

	return output, nil
}
