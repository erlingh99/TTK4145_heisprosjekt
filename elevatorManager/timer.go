package elevatorManager

import (
	"fmt"
	"time"
)

var doorTimer *time.Timer

func timer_start(d time.Duration) {
	doorTimer = time.NewTimer(d)
}

func timer_timedOut() bool {
	select {
	case _ = <-doorTimer.C:
		fmt.Println("HAHARHERE")
		return true
	default:
		return false
	}
}

// func main() {
// 	timer_start(config.DOOR_TIMEOUT)
// 	fmt.Println("TIMER STARTED")

// 	for !timer_timedOut() {

// 	}
// 	fmt.Println("TIMER DONE")
// }
