//Global constants for the elevator
package config

import (
	"time"
)

var N_FLOORS = 3
var N_BUTTONS = 3
var DOOR_TIMEOUT = 3 * time.Second
var POLLRATE = 25 * time.Millisecond


const CONNECT_ADDR = "192.168.1.176"
const CONNECT_PORT = 20002

const LISTEN_PORT = 20002

const BROADCAST_PORT = 20001

const MASTER_BROADCAST_INTERVAL = 1 * time.Second
const MASTER_BROADCAST_LISTEN_TIMEOUT = 3 * time.Second
