package networking

import (
	"fmt"
	// "container/list"
	// "time"
	"net"
	// "elevatorproject/config"
	// "elevatorproject/network/bcast"
	// "../network/conn"
	// "../network/localip"
	// "../network/peers"
	// "elevatorproject/network/tcp"
	"encoding/json"
	"reflect"
	// "strings"
)

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func MasterListener(port int, errorChan chan<- error, chans ...interface{}) {
	checkArgs(chans...)

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp4", addr)

	if err != nil {
		errorChan <- err
		listener.Close()
		return
	}
	n := 0
	for range chans {
		n++
	}

	for {
		conn, _ := listener.Accept()
		ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
		fmt.Println("ip:", ip)
		go func() {
			selectCases := make([]reflect.SelectCase, n)
			typeNames := make([]string, n)
			for i, ch := range chans {
				selectCases[i] = reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(ch),
				}
				typeNames[i] = reflect.TypeOf(ch).Elem().String()
			}

			for {
				chosen, value, _ := reflect.Select(selectCases)
				buf, _ := json.Marshal(value.Interface())

				_, err = conn.Write([]byte(typeNames[chosen] + string(buf)))

				if err != nil {
					errorChan <- err
					return
				}
			}
		}()
	}
}

// Checks that args to Tx'er/Rx'er are valid:
//  All args must be channels
//  Element types of channels must be encodable with JSON
//  No element types are repeated
// Implementation note:
//  - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//    so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg#%d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg#%d and arg#%d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		switch elemType.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (arg#%d)",
				elemType.String(), i+1))
		case reflect.Map:
			if elemType.Key().Kind() != reflect.String {
				panic(fmt.Sprintf(
					"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (arg#%d)",
					elemType.String(), i+1))
			}
		}
	}
}
