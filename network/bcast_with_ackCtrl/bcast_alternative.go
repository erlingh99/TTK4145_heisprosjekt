package bcast_with_ackCtrl

import (
	"elevatorproject/config"
	"elevatorproject/network/conn"
	"elevatorproject/utils"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"
)

type AcknowledgeCtrl struct { 
	Msg			interface{}
	ID 			string
	AcksNeeded	[]string
	AcksRecvd	[]string
	SendTime	time.Time
	SendNum		int
}

type AcknowledgeMsg struct {
	ID			string
	ElevID		string
}

type AckList []*AcknowledgeCtrl

func (al *AckList) AddAck(a *AcknowledgeCtrl) {
	for _, ack := range *al {
		if ack.ID == a.ID {
			ack.SendNum++
		}
	}
	*al = append(*al, a)
}

func (al *AckList) AckRecieved(a *AcknowledgeMsg) {
	for _, ack := range *al {
		if ack.ID == a.ID {
			if !utils.Contains(ack.AcksRecvd, a.ElevID) {
				ack.AcksRecvd = append(ack.AcksRecvd, a.ElevID) //only add if it is not already there
			}
		}
	}
}

func (al *AckList) RemoveCompletedAcks() {
	al2 := make(AckList, 0)
	for _, ack := range *al {
		if !utils.StringArrEqual(ack.AcksNeeded, ack.AcksRecvd) {
			al2 = append(al2, ack)
		}
	}	
	*al = al2
}

func (al *AckList) CheckForTimedoutSends() []string {
	missingAcks := make([]string, 0)

	for _, acks := range *al {
		if acks.SendNum > config.MAX_RESENDS {
			fmt.Printf("Couldn't send %+v\n", acks.Msg)
			missing := utils.StringArrDiff(acks.AcksNeeded, acks.AcksRecvd)
			//fmt.Println("Acks missing: ", missingAcks)			
			missingAcks = append(missingAcks, missing...)			

			//mark element as completed, so it is removed			
			acks.AcksRecvd = acks.AcksNeeded
		}
	}
	return missingAcks
}

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func Transmitter(id string, port int, AckNeeded chan<- AcknowledgeCtrl, chans ...interface{}) {
	checkArgs(chans...)

	n := 0
	for range chans {
		n++
	}

	selectCases := make([]reflect.SelectCase, n)
	typeNames := make([]string, n)
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for {
		chosen, value, _ := reflect.Select(selectCases)
		buf, _ := json.Marshal(value.Interface())
		conn.WriteTo([]byte(typeNames[chosen]+string(buf)), addr)

		if typeNames[chosen] != "bcast_with_ackCtrl.AcknowledgeMsg" {
			//Make ackCtrl for the message
			AckNeeded <- AcknowledgeCtrl{	Msg: 		value,
											ID:			string(buf), 
											SendTime: 	time.Now(),
											SendNum: 	1,
											AcksRecvd:	make([]string, 1)}
		}		
	}
}

func ResendMsg(msg interface{}, sendChans ...interface{}){
	for _, ch := range sendChans {
		if reflect.TypeOf(ch).Elem() == reflect.TypeOf(msg) {
			sendCase := reflect.SelectCase{	Dir:  reflect.SelectSend,
											Send: reflect.ValueOf(msg),
											Chan:  reflect.ValueOf(ch)}
			reflect.Select([]reflect.SelectCase{sendCase})			
			return
		}
	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(id string, port int, ackSend chan<- AcknowledgeMsg, chans ...interface{}) {
	checkArgs(chans...)

	var buf [1024]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])

		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
			continue
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

				if  reflect.TypeOf(ch).Elem().String() != "bcast_with_ackCtrl.AcknowledgeMsg" {
					//send ackMessage
					ackSend <- AcknowledgeMsg{	ID:			string(buf[0:n]), 											
												ElevID: 	id}

				}
			}
		}
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