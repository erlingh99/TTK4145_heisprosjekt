package networking_bcast

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"
	"elevatorproject/network/conn"
)

type Acknowledgement struct {
	Msg 		interface{}
	ID 			string
	elevID 		[]string
	sendTime	time.Time
	sendNum		int
}

type ackList []*Acknowledgement

func (al ackList) AddAck(a *Acknowledgement) error {
	//if it exist increment resendnum
	//if resendnum exceeds limit remove remaining elevators, wab peers?
	for i, ack := range al {
		if ack.ID == a.ID {
			ack.sendNum++

			if ack.sendNum > config.MAX_RESENDS {
				return ack.elevID
			}
			return nil
		}
	}
	al = append(al, a)
	return nil
}

func (al ackList) AckRecieved(a *Acknowledgement) {

	for i, ack := range al {
		if ack.ID == a.ID {
			if !contains(ack.elevID, a.elevID[0])
			{
				ack.elevID = append(ack.elevID, a.elevID[0]) //only add if it is not already there
			}

			if len(ack.elevID) == len(peerIDs) {
				al.removeAtPos(i) //should check index i again bc how remove alters al
			} 
		}
	}
}

func (al ackList) remove(i int) {
	al[i] = al[len(al)-1]
	al = al[:len(al)-1]
}


bcast_alt(alertOrderHandler chan string, sendChans []interface{}, recvChans []interface{}) {

	peerIDs := make([]string, 0)
	waitingForAcks := make(ackList, 0)


	AckRecieved := make(chan Acknowledgement)
	NewAckReqMade := make(chan Acknowledgement)

	go bcast.Transmitter(bcastPort, AckRecieved, sendChans)
	go bcast.Reciever(bcastPort, AckRecieved, recvChans)


	peerChan := make(chan PeerUpdate)
	transmitEnable := make(chan bool)
	go peers.Transmitter(peerPort, myID, transmitEnable)
	go peers.Reciever(peerPort, peerChan)

	msgResendTicker := time.Ticker(config.RESEND_RATE)


	for {
		select {
		case p<-peerUpdateChannel:		
			fmt.Println("PEER UPDATE")
			fmt.Printf("* Peers: %v\n", p.Peers)
			fmt.Printf("* New: %v\n", p.New)
			fmt.Printf("* Lost: %v\n", p.Lost)

			peerIDs = append(peerIDs, p.New)


			if p.Lost != "" {		
				for _, lostPeer := range p.Lost {
					peerIDs = remove(peerIDs, lostPeer)
					alertOrderHandler <-lostPeer
				}
			}

		case <-msgResendTicker.C:
			for _, ack := range waitingForAcks {
				if time.Now().After(ack.SendTime + ack.sendNum * RESEND_RATE) {
					// use reflect with ack.msg to send on correct sendchan	
					for _, ch := range sendChans {
						if reflect.TypeOf(ch).Elem().String() == reflect.TypeOf(ack.msg).Elem().String()
						{
							ch <- ack.Msg
							break
						}
					}					
				}
			}
		
		case ack := <-ackRecv:
			waitingForAcks.AckRecieved(&ack)

		case ack := <-ackReqCh:
			err := waitingForAcks.AddAck(&ack)
			if err != nil { //error contains elevators that has acket the msg, need to find remaining using peers
				for _, p := range peerIDs {
					if !contains(err, p) {
						remove(peerIDs, p)
						alertOrderHandler <- p											
						fmt.Println("removed because of lacking ack: " + p)
					}
				}
			}
		}
	}
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func remove(s []string, e string) []string {
	for i, a := range s {
        if a == e {
			s[i] = s[len(s)-1]
			return s[:len(s)-1]
        }
    }
    return s
}

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func Transmitter(port int, ackReqCh chan<- Acknowledgment, chans ...interface{}) {
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


		//er v sendt når man kommer hit???
		ackReqCh <- Acknowledgement{	Msg: 		value,
										ID:			buf, 
										sendTime: 	time.Now(),
										sendNum: 	1,
										elevID: 	make([]string){thiselevatorsId}}
		
	}
}


// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, ackRecv chan<- Acknowledgment, chans ...interface{}) {
	checkArgs(chans...)

	var buf [1024]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])

		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
			//return
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
				

				//er v mottat når man kommer hit???
				ackRecv <- Acknowledgement{	Msg: 		v,
											ID:			buf, 
											sendTime: 	time.Now(),
											sendNum: 	1,
											elevID: 	make([]string){thiselevatorsId}}
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
