package elevatorManager

import (
	"elevatorproject/config"
	"time"
	"fmt"
)


// Timer logic for the doortimer

var doorTimer *time.Timer



func timer_init() {
	doorTimer = time.NewTimer(config.DOOR_TIMEOUT)
	if !doorTimer.Stop() {
		<- doorTimer.C
	}
}

func timer_start() {
	doorTimer.Stop()
	select {
	case <-doorTimer.C:
	default: 
	}
	fmt.Println("started timer")
	doorTimer.Reset(config.DOOR_TIMEOUT)
}
 