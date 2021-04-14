package main

import (
	"os/exec"
)

func Distributer() {
	retvals, err := exec.Command("./hall_request_assigner/hall_request_assigner", "-i", "")
}
