package elevatorManager

import (
	"time"
)


var doorTimer *time.Timer
var timerStarted = false

func timer_start(d time.Duration) {
	doorTimer = time.NewTimer(d)
	timerStarted = true
}

func timer_timedOut() bool {
	if timerStarted {
		select {
		case _ = <-doorTimer.C:
			timerStarted = false
			return true
		default:
			return false
		}
	}
	return false
}
