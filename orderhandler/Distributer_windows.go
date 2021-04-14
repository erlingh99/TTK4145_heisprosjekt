package main

import (
	"os/exec"
)

func Distributer(inpt string) {


	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner.exe", "-i", string(jsonBytes)).Output()
}
