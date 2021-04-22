package netCheck

import (
	"time"
	"config"
)

func netChecker(status chan bool) {
	prev := false

	tick := time.NewTicker(config.NET_TICK_INTERVAL)


	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})

	conn.SetDeadline(time.Now().Add(config.NET_TICK_INTERVAL))

	if err != nil {
		return "", err
	}
		defer conn.Close()


	for {
		select {
			case <-tick.C
				if prev {
					prev = false
					status <- prev
				}
			case 
				tick.reset()

		}
	}
}