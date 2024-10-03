package websocket

import (
	"github.com/gookit/color"
	"github.com/rs/zerolog"
)

type RegistrationHub struct {
	// Register requests from the c.
	register chan *Client
	// Unregister requests from c.
	unregister chan *Client

	s *Server
}

func NewRegistrationHub(s *Server) *RegistrationHub {
	return &RegistrationHub{
		// Channels
		register:   make(chan *Client),
		unregister: make(chan *Client),
		s:          s, // Server
	}
}

func (h *RegistrationHub) unregisterClient(client *Client) {
	// Unregister from all current hubs
	go func() {
		for hub := range h.s.Hubs {
			// We are sending to all hubs that are related to this websocket server! (children)
			hub.UnregisterClientChannel <- client
		}
	}()

	// Hub -> Server -> ClientsStatus
	h.s.c.unregisterClient(client)

	// Close the channel for output buffering
	// Should we wait until we unregister from everywhere?!
	close(client.send)
}

func (h *RegistrationHub) run() {
	// Start an infinite loop!`

	info := func() *zerolog.Event {
		return h.s.LInfoF("run").Str("sub_module", "registration_hub")
	}
	_error := func() *zerolog.Event {
		return h.s.LErrorF("run").Str("sub_module", "registration_hub")
	}

	info().Msg("running...")
	defer info().Msg("leaving...")
	terminate := false
	for {
		select {
		case client := <-h.register:
			h.s.c.registerClient(client)
			info().Msg(color.Style{color.LightGreen}.Render("registering new client"))
		case client := <-h.unregister:
			// On Unregister
			// Unregister from hubs!
			h.unregisterClient(client)
			info().Msg(color.Style{color.FgLightMagenta}.Render("unregistering client"))

		// TODO: should we destroy the object itself?! CLIENT -> i suppose yes... because multiple have references?!
		// TODO: maybe create a timer which will destroy after couple of seconds!

		case <-h.s.ctx.Done():
			terminate = true
			info().Msg("terminating and disconnecting all connected clients")

			// TODO: Break from this routine!
			// TODO: close all connections?!

			// Simply unregister all clients...
			// The clients routines will also receive the signal and they will auto close...!?

			// TODO: all clients should disconnect and get offline based on the provided context from the WebSocket Server!!!
			// TODO: we should remove this code from here!!!
			clients := h.s.GetClients()
			for cl, _ := range clients {
				_err := cl.Disconnect()
				_error().Err(_err).Msg(color.Style{color.LightRed}.Render("failed to disconnect client"))
				h.unregisterClient(cl)
			}
		}
		if terminate {
			break
		}
	}
}
