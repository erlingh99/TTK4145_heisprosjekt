package elevatorManager

import "../elevio/elevator_io.go"
import "time"

const ObstructionTimeout = 1*time.Second

func GoToFloor(floor int, Elevator el) {
	if (el.dirn != obstructed) {
		if (floor == el.curFloor) {

		}
		else if (floor < el.curFloor) {
			SetMotorDirection(MD_Up)

		}
		else if (floor > el.curFloor) {

		}

	else (
		time.Sleep(ObstructionTimeout)
		GoToFloor(floor, el)
	)
}