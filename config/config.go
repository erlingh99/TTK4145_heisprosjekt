//Global constants for the elevator
package config

import (
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3
const DOOR_TIMEOUT = 3 * time.Second
const POLLRATE = 25 * time.Millisecond

const MAX_RESENDS = 3
const RESEND_RATE = 300 * time.Millisecond


const CONNECT_ADDR = "192.168.1.176"
const CONNECT_PORT = 20002

const PEER_PORT =  20005
const BCAST_PORT = 20010

const LISTEN_PORT = 20002

const BROADCAST_PORT = 20001

const REQUEST_CONNECTION_PORT = 20003

const TCP_TX_PORT_MASTER = 20006
const TCP_RX_PORT_MASTER = 20007

const IDLE_CONN_TIMEOUT = 2 * time.Second

const MASTER_BROADCAST_INTERVAL = 1 * time.Second
const MASTER_BROADCAST_LISTEN_TIMEOUT = 5 * time.Second

const SLAVE_ALIVE_MSG_INTERVAL = 1 * time.Second
const SLAVE_ALIVE_MSG_LISTEN_TIMEOUT = 5 * time.Second
