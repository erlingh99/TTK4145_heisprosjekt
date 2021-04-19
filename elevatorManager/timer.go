package elevatorManager

import (
	"elevatorproject/config"
	"time"
)


var doorTimer *time.Timer
//var timerStarted = false


func timer_init() *time.Timer {
	doorTimer := time.NewTimer(config.DOOR_TIMEOUT)
	if !doorTimer.Stop() {
		<- doorTimer.C
	}
	return doorTimer
}

func timer_start() {

	//timerStarted = true

	if !doorTimer.Stop() {
		<-doorTimer.C
	}
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