//Global constants for the elevator
package config

import (
	"time"
)

var N_FLOORS = 3
var N_BUTTONS = 3
var DOOR_TIMEOUT = 3 * time.Second
var POLLRATE = 25 * time.Millisecond