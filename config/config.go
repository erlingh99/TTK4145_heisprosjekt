//Global constants for the elevator
package config

import (
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3

//ack consts
const MAX_RESENDS = 3
const RESEND_RATE = 300 * time.Millisecond

//order consts
const ORDER_MAX_EXECUTION_TIME = 15 * time.Second

//elev consts
const DOOR_TIMEOUT = 3 * time.Second
const POLLRATE = 25 * time.Millisecond
const ELEV_SHARE_STATE_INTERVAL = 2 * time.Second

//bcast ports
const PEER_PORT =  20005
const BCAST_PORT = 20010

//Conn timeout consts
const IDLE_CONN_TIMEOUT = 2 * time.Second
const NET_TIMEOUT = 2 * time.Second


//alive msg
const MASTER_BCAST_PORT = 20008
const MASTER_BROADCAST_INTERVAL_MIN = 100  // ms
const MASTER_BROADCAST_INTERVAL_MAX = 1500 // ms