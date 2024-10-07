package websocket

import (
	"context"

	"github.com/kyaxcorp/go-helper/_context"
	"github.com/kyaxcorp/go-helper/sync/_bool"
	"github.com/kyaxcorp/go-helper/sync/_map_string_interface"
	"github.com/kyaxcorp/go-helper/sync/_time"
)

type hubBroadcastTo struct {
	to   map[uint64]*Client
	data []byte
}

type Hub struct {
	server *Server

	// The registered c in this hub!
	//c map[*Client]bool
	c *clientsData

	// Created time (when the hub) has being created
	createdAt *_time.Time

	// This is the channel where we receive the data for broadcast!
	broadcast       chan []byte
	broadcastStatus chan map[*Client]SendStatus
	// This is the channel which handles specific c!
	broadcastTo       chan hubBroadcastTo
	broadcastToStatus chan map[*Client]SendStatus
	// This is the command to stop the broadcaster which sends the data!
	//stopBroadcaster chan bool
	// This is the command to stop the getter
	//stopGetter chan bool
	// This is the command to stop the controller
	//stopController chan bool

	// If the stop has being called!
	StopCalled *_bool.Bool

	// Has being started!?
	isRunning *_bool.Bool

	// TODO: see how to use the context here!
	//stopContext context.CancelFunc

	parentCtx context.Context
	ctx       *_context.CancelCtx

	// Data HubGetter
	getter HubGetter
	// On Start
	onStart HubOnStart
	// On Start HubGetter
	onStartGetter HubOnStartGetter
	// On Start Broadcast
	onStartBroadCast HubOnStartBroadCast

	// When client has being registered
	onClientRegister *_map_string_interface.MapStringInterface
	// When client has being unregistered
	onClientUnRegister *_map_string_interface.MapStringInterface

	// ControlMessages
	ControlChannel chan int

	// Unregistered ClientsStatus
	UnregisterClientChannel chan *Client
}

type HubGetter func(h *Hub)
type HubOnStart func(h *Hub)
type HubOnStartGetter func(h *Hub)
type HubOnStartBroadCast func(h *Hub)
type HubOnClientRegister func(h *Hub, client *Client)
type HubOnClientUnRegister func(h *Hub, client *Client)
