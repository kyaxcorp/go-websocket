package websocket

import (
	"github.com/rs/zerolog"
)

type BroadcastHub struct {
	// Inbound messages from the c.
	broadcast chan []byte

	s *Server
}

func NewBroadcastHub(s *Server) *BroadcastHub {
	return &BroadcastHub{
		// Channels
		broadcast: make(chan []byte),
		s:         s, // Server
	}
}

func (h *BroadcastHub) run() {
	// Start an infinite loop!`
	info := func() *zerolog.Event {
		return h.s.LInfo().Str("sub_module", "broadcast_hub")
	}

	info().Msg("running...")
	defer info().Msg("leaving...")

	terminate := false
	for {

		select {
		case message := <-h.broadcast:
			// On Broadcast (Messages to all c)
			//log.Println("Sending broadcast Message")
			info().Msg("sending message")

			// TODO: make we should slice... because the data can change when reading, and there is no guarantee
			// that will be the same when looping!
			for client := range h.s.c.GetClients() {
				// TODO: we should check if the channel is still active!

				select {
				case client.send <- message:
				default:
					/*// Closing the channel!
					close(client.send)
					// Deletes the element from the map!
					delete(h.c, client)*/
				}
			}
		case <-h.s.ctx.Done():
			// This is the general Hub... we should simply terminate it!
			info().Msg("terminating...")
			terminate = true
		}
		if terminate {
			break
		}
	}
}
