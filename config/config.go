//Global constants for the elevator
package config

import (
	"time"
)

const N_FLOORS = 3
const N_BUTTONS = 3
const DOOR_TIMEOUT = 3 * time.Second
const POLLRATE = 25 * time.Millisecond

const CONNECT_ADDR = "192.168.1.176"
const CONNECT_PORT = 20002

const LISTEN_PORT = 20002

const BROADCAST_PORT = 20001

const MASTER_BROADCAST_INTERVAL = 1 * time.Second
const MASTER_BROADCAST_LISTEN_TIMEOUT = 3 * time.Second
