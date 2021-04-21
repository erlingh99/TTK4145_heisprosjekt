package tcp

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
)

/*
 * Functions in this file:
 *
 *


	- Master, tar i mot tilkoblinger, kan sende og skrive på de

	- Slave, kobler til, kan sende og skrive på de
 */
 
 type IdIpPair struct {
	 ID string
	 IP string
 }


 func Master(txPortMaster int, rxPortMaster int, errChan chan<- error, stopChan <-chan bool, sendChans []interface{},  recvChans []interface{}) {
	ipToIdMap := make(map[string]string)
	conLostIPCh := make(<-chan string)
	idIpPairCh	:= make(<-chan IdIpPair)

	stopTransmitterCh := make(chan bool, 1)
	stopRecieverCh := make(chan bool, 1)
	
	go MasterTransmitter(txPortMaster, errChan, stopTransmitterCh, conLostIPCh, sendChans...)
	go MasterReceiver(rxPortMaster, errChan, stopRecieverCh, idIpPairCh, recvChans...)

	for {
		select {
		case ipLost := <-conLostIPCh:
			delete(ipToIdMap, ipLost)
			//Send to orderhandler on the right channel
		case idIp := <-idIpPairCh:
			ipToIdMap[idIp.IP] = idIp.ID
		case <-stopChan:
			stopTransmitterCh<- true
			stopRecieverCh<- true
			return
		}
	}
 }

 func Slave(ip string, txPortMaster int, rxPortMaster int, errChan chan<- error, stopChan <-chan bool, sendChans []interface{},  recvChans []interface{}) {
	go SlaveTransmitter(ip, rxPortMaster, errChan, stopChan, sendChans)
	go SlaveReceiver(ip, txPortMaster, errChan, stopChan, recvChans)
 }


// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func SlaveReceiver(ip string, port int, errorChan chan<- error, stopChan <-chan bool, chans ...interface{}) {
	checkArgs(chans...)

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp4", addr)

	if err != nil {
		errorChan <- err
		return
	}

	var buf [1024]byte
	for {
		n, e := conn.Read(buf[0:])
		select {
		case <-stopChan:
			stopChan<- true
			return
		default:
		}
		if e != nil {
			fmt.Printf("tcp.Receiver(:%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
			errorChan <- err
			conn.Close()
			return
		}
		for _, ch := range chans {
			T := reflect.TypeOf(ch).Elem()
			typeName := T.String()
			if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
				v := reflect.New(T)
				json.Unmarshal(buf[len(typeName):n], v.Interface())

				reflect.Select([]reflect.SelectCase{{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(ch),
					Send: reflect.Indirect(v),
				}})
			}
		}
	}
}

// Encodes received values from `chans` into type-tagged JSON, then transmits them
// to `ip` on `port`
func SlaveTransmitter(ip string, port int, errorChan chan<- error, stopChan <-chan bool, chans ...interface{}) {
	checkArgs(chans...)

	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp4", addr)

	if err != nil {
		errorChan <- err
		return
	}

	n := 0
	for range chans {
		n++
	}

	for {
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
			select {
			case <-stopChan:
				stopChan<- true
				return
			default:
			}
			buf, _ := json.Marshal(value.Interface())

			_, err = conn.Write([]byte(typeNames[chosen] + string(buf)))
			// for i := 0; i < N_PLEX; i++ {
				
			// 	if err == nil break				
			// }

			

			if err != nil {
				errorChan <- err
				conn.Close()
				return
			}
		}
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func MasterTransmitter(port int, errorChan chan<- error, stopChan <-chan bool, conLostIPCh chan<- string, chans ...interface{}) {
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

	var conns []net.Conn
	ipToIdMap := make(map[string]string)
	
	stopAccepterCh := make(chan bool)
	go func() {
		for {
			conn, _ := listener.Accept()
			select {
			case <-stopAccepterCh:
				// Stop
				return
			default:
			}
			conns = append(conns, conn)
		}
	}()

	selectCases := make([]reflect.SelectCase, n)
	typeNames := make([]string, n)
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}
	selectCases = append(selectCases, reflect.SelectCase{
		Dir: reflect.SelectRecv,
		Chan: reflect.ValueOf(stopChan)})

	for {
		chosen, value, _ := reflect.Select(selectCases)

		if chosen == n {
			// Stop
			stopAccepterCh<- true
			return
		}

		for _, conn := range conns {
			buf, _ := json.Marshal(value.Interface())
	
			i, err := conn.Write([]byte(typeNames[chosen] + string(buf)))
	
			if err != nil {
				// Here we should have a way to remove conns that are no longer in use, but it will look a little messy,
				// and I don't think we will face any issues with it. Better to just restart from time to time...
				// (the below code is not fully functional)
				// errorChan <- err
				ip := conn.RemoteAddr().(*net.TCPAddr).IP.String()
				conLostIPCh <- ip
				conn.Close()
				conns = append(conns[:i], conns[i+1:]...)
				// return
			}
		}
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func MasterReceiver(port int, errorChan chan<- error, stopChan chan bool, idIpPairCh chan<- IdIpPair, chans ...interface{}) {
	checkArgs(chans...)

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp4", addr)

	if err != nil {
		errorChan <- err
		listener.Close()
		return
	}

	var buf [1024]byte

	for {
		conn, _ := listener.Accept()
		select {
		case <-stopChan:
			stopChan<- true
			return
		default:
		}
		go func() {
			for {
				n, e := conn.Read(buf[0:])
				select {
				case <-stopChan:
					stopChan<- true
					return
				default:
				}
				if e != nil {
					fmt.Printf("tcp.Receiver(:%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
					errorChan <- err
					conn.Close()
					return
				}
				ip := conn.RemoteAddr().(*TCPAddr).IP.String()
				for _, ch := range chans {
					T := reflect.TypeOf(ch).Elem()
					typeName := T.String()
					if strings.HasPrefix(string(buf[0:n])+"{", typeName) {
						v := reflect.New(T)
						json.Unmarshal(buf[len(typeName):n], v.Interface())
						
						if typeName == "elevatormanger.Elevator" {
							ipIdPair := IdIpPair{IP: ip, ID: v.Interface().ID}
							idIpPairCh <- ipIdPair
						}

						reflect.Select([]reflect.SelectCase{{
							Dir:  reflect.SelectSend,
							Chan: reflect.ValueOf(ch),
							Send: reflect.Indirect(v),
						}})
					}
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
