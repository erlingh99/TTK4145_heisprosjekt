package localip

import (
	"net"
	"strings"
	"elevatorproject/config"
	"fmt"
)

var localIP string

func LocalIP() (string, error) {
	d := net.Dialer{Timeout: config.NET_TIMEOUT}
	conn, err := d.Dial("tcp", "8.8.8.8:53")
	fmt.Println(string([]byte{8, 8, 8, 8}))

	if err != nil {
		return "", err
	}

	defer conn.Close()
	localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	return localIP, nil
}
