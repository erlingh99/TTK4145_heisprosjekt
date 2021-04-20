package elevatorManager

import (
	"elevatorproject/config"
	"time"
	"fmt"
)


var doorTimer *time.Timer
//var timerStarted = false


func timer_init() {
	doorTimer = time.NewTimer(config.DOOR_TIMEOUT)
	if !doorTimer.Stop() {
		<- doorTimer.C
	}
}

func timer_start() {

	//timerStarted = true
	doorTimer.Stop()
	select {
	case <-doorTimer.C:
	default: 
	}
	fmt.Println("started timer")
	doorTimer.Reset(config.DOOR_TIMEOUT)
}
 
/*
func timer_timedOut() bool {
	if timerStarted {
		select {
		case <-doorTimer.C:
			timerStarted = false
			return true
		default:
			return false
		}
	}
	return false
}
*/