package websocket

import (
	"github.com/kyaxcorp/go-helper/function"
	"github.com/kyaxcorp/go-helper/sync/_bool"
	"github.com/kyaxcorp/go-helper/sync/_map_string_interface"
	"github.com/kyaxcorp/go-helper/sync/_time"
)

func newHub(WSServer *Server) *Hub {

	// TODO: subscribe to server channels when unregister or close connection of a client happens!
	// In this case it auto removes from the hub!

	hub := &Hub{
		createdAt: _time.NewNow(),
		// The WebSocket server
		server: WSServer,
		// Create the c map!
		//c: make(map[*Client]bool),
		c: NewClientsInstance(),
		// Create the data channel!
		broadcast: make(chan []byte),
		// Create (to) the Sending data Channel
		broadcastTo: make(chan hubBroadcastTo),
		// Create the stop call channel
		//stopBroadcaster: make(chan bool),
		//stopGetter:      make(chan bool),
		//stopController:  make(chan bool),
		// Is it running!
		isRunning: _bool.NewVal(false),

		StopCalled: _bool.NewVal(false),

		ControlChannel: make(chan int),

		// Events
		onClientRegister:   _map_string_interface.New(),
		onClientUnRegister: _map_string_interface.New(),

		// Unregister channel!
		UnregisterClientChannel: make(chan *Client),
	}

	hub.SetContext(WSServer.ctx.Context())

	return hub
}

// NewHub -> It creates a special custom hub with specific functionality for handling c
func (s *Server) NewHub(getter ...HubGetter) *Hub {
	hub := newHub(s)
	if len(getter) > 0 {
		if function.IsCallable(getter[0]) {
			hub.SetGetter(getter[0])
		}
	}

	return hub
}
