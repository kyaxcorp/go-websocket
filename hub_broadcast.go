package websocket

/*const SplitAfter = 10
const MaxSplitting = 50
const SplitForPercentageLoad = 25 // 25 percentage*/

// ------------- ALGO nr.1-----------------\\
// For better splitting, we should create ranges?
// If more than
// Check the Total
// we should define the Coeficient of proportion
// We should define in percentage how much want to split
// For example we have 100 connections
// We want to split using 25 %
//  100 Connections ....... 100 %
//  x Connections.......... 25 %
/// 100 Connections * 25 % / 100 % = 25 Connections
// 100 Connections / 25 Connections = 4 Go Routines
// 1750 Connections ........ 100 %
// x Connections ............ 25 %
// 1750 Conn * 25% / 100% = 437.5 Connections
// 1750 / 437.5 = 4 GoRoutines
//
// ---------------ALGO nr. 2------------------\\
// Total nr of ClientsStatus
// If Nr of ClientsStatus
// x < 10 -> 2 routine
// 10 < x < 50 ->  5 routines
// 50 < x < 100 -> 10 routines
// 100 < x < 500 ->  20 routines
// 500 < x < 1000 -> 40 routines
// 1000 < x < 5000 -> 80 routines
// 5000 < x < 10000 -> 160 routines
// 10000 < x < 50000 -> 320 routines
// 50000 < x -> 400

func getNrOfRoutines(nrOfConnections uint64) uint16 {
	routines := 0
	switch conn := nrOfConnections; {
	case conn <= 5:
		routines = 1
	case 6 <= conn && conn <= 10:
		routines = 2
	case 11 <= conn && conn <= 50:
		routines = 5
	case 51 <= conn && conn <= 100:
		routines = 10
	case 101 <= conn && conn <= 500:
		routines = 20
	case 501 <= conn && conn <= 1000:
		routines = 40
	case 1001 <= conn && conn <= 5000:
		routines = 80
	case 5001 <= conn && conn <= 10000:
		routines = 320
	case 10001 <= conn && conn <= 50000:
		routines = 400
	case 50001 <= conn:
		routines = 500
	default:
		routines = 1
	}
	return uint16(routines)
}

// This function should be called by the programmer!
func (h *Hub) run() {
	// If it's running then return
	if h.isRunning.IfFalseSetTrue() {
		return
	}

	defer func() {
		h.isRunning.False()
	}()

	// On Start callback
	if h.onStartBroadCast != nil {
		h.onStartBroadCast(h)
	}

	for {
		if h.StopCalled.Get() {
			break
		}
		select {
		//case <-h.stopBroadcaster:
		//	break
		case <-h.ctx.Done():
			break
		case message := <-h.broadcast:
			// TODO: if we have multiple c, we should split the sending by creating additional
			// Goroutines for faster sending!
			// Each goroutine will handle a specific nr of c
			// TODO: we should define a param, which is the maximum nr of c a goroutine can handle...
			// and if it's higher, then we should create a formula of generating a specific nr. of goroutines and
			// split the c to them!

			// On Broadcast (Messages to all c)

			go func() {
				nrOfClients := h.c.GetNrOfClients()
				if nrOfClients == 0 {
					return
				}

				nrOfRoutines := getNrOfRoutines(uint64(nrOfClients))
				clients := h.c.GetClientsInChunks(nrOfRoutines)

				// Split in multiple routines if there are many ClientsStatus

				if clients == nil {
					return
				}

				for _, clientsChunk := range clients {
					go func(c map[*Client]bool) {
						for client := range c {
							if h.StopCalled.Get() {
								break
							}

							// TODO: before sending we should check if this client is not disconnected
							// Or unregistered somehow .... because when starting the loop.. it can take some time to send the information
							// to the c!

							if client != nil && !client.isClosed.Get() {
								client.send <- message
							}
						}
					}(clientsChunk)
				}
			}()
		case broadcastTo := <-h.broadcastTo:

			// For faster broadcasting maybe we should goroutine here... because looping through ClientsStatus takes some time...!
			// And if starting more goroutines that will also start looping and transmit messages will not be a problem!
			// This will improve speed, but can consume resources!

			// TODO: do we need locks on the map that's coming through channel!?
			// here usually we will not need it because the map it's being created else where and it's not used
			// by multiple goroutines!

			go func() {
				nrOfClients := len(broadcastTo.to)
				if nrOfClients == 0 {
					return
				}
				nrOfRoutines := getNrOfRoutines(uint64(nrOfClients))
				clients := GetClientsInChunksWithConn(broadcastTo.to, nrOfRoutines)
				if clients == nil {
					return
				}

				for _, clientsChunk := range clients {
					go func(c map[uint64]*Client) {
						for _, client := range c {
							if h.StopCalled.Get() {
								break
							}

							// TODO: before sending we should check if this client is not disconnected
							// Or unregistered somehow .... because when starting the loop.. it can take some time to send the information
							// to the c!

							if client != nil && !client.isClosed.Get() {
								client.send <- broadcastTo.data
							}
						}
					}(clientsChunk)
				}
			}()
		}
	}
}
