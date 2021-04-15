package orderHandler

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func Distributer(input string) {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("json.Marshal error: ", err)
	}

	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner.exe", "-i", string(jsonBytes)).Output()
}
