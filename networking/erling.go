

// recvChans[0] must be sharestate
func network_init(	id string,
					errorchan chan<- error,
					enableIpBroadcast chan<- bool,
					sendChans []interface{},
					recvChans []interface{}) {

	ID := id
	var masterIP string

	bcastRx := make(chan string)
	bcastTx := make(chan string)
	
	go bcast.Receiver(config.BROADCAST_PORT, bcastRx)
	go bcast.Transmitter(config.BROADCAST_PORT, bcastTx)


	// go tcp.Transmitter(ip, port, errorChan, sendChans)
	// go tcp.Receiver(port, id, errorChan, recvChans)

	sharestateToLocalOh := recvChans[0]
	recvChans[0] = make(chan elevatormanager.Elevator)

	stopMasterCh := make(chan bool)
	stopSlaveCh := make(chan bool)
	
	// If you are master:
	//go tcp.Master(config.TCP_TX_PORT_MASTER, config.TCP_RX_PORT_MASTER, errorChan, sendChans, recvChans)
	// If you are slave/each time there is a new master:
	// go tcp.Slave(ip, config.TCP_TX_PORT_MASTER, config.TCP_RX_PORT_MASTER, errorChan, stopSlaveCh, sendChans, recvChans)





	ticker := time.NewTicker(config.BCAST_RATE)

	for {
		select {

		case b := <-enableIpBroadcast
			if b {
				stopSlaveCh<- true
				go tcp.Master(config.TCP_TX_PORT_MASTER, config.TCP_RX_PORT_MASTER, errorChan, stopMasterCh, sendChans, recvChans)
				ticker.Stop()
				ticker.Reset() //obs obs, bør tømme kanalen
			} else {
				stopMasterCh<- true
				go tcp.Slave(ip, config.TCP_TX_PORT_MASTER, config.TCP_RX_PORT_MASTER, errorChan, stopSlaveCh, sendChans, recvChans)
				ticker.Stop()
			}
			
		case <-ticker.C:
			bcastTx<-IP

		case state := <-sharestateToLocalOh:
			
		

		case msg := <-bcastRx:
			//hvis ikke lokal ip
			//koble til denne ip'en
			if (masterIP != msg) {
				masterIP = msg
				go tcp.Slave(masterIP, config.TCP_TX_PORT_MASTER, config.TCP_RX_PORT_MASTER, errorChan, stopSlaveCh, sendChans, recvChans)
			}
			
			//send videre til orderHandler

		case err := <- errorChan:
			//error si ifra til orderhandler

		case sharestate := <-recvChans[0]:
			for key, element := range ipToIdMap {
				if sharestate.ID == element {
					break
				} else {
					ipToIdMap[] = sharestate
				}
			}  
			sharestateToLocalOh <- sharestate
		}
	}
}