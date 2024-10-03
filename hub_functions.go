package websocket

import (
	"time"

	"github.com/kyaxcorp/go-helper/function"
)

func (h *Hub) NrOfClients() uint {
	return h.c.GetNrOfClients()
}

// RegisterClient -> Adds the client to the hub!
func (h *Hub) RegisterClient(client *Client) *Hub {
	// Register the Client to the Index
	h.c.registerClient(client)
	// Calling an event with Goroutine
	go h.onClientRegister.Scan(func(k string, v interface{}) {
		v.(HubOnClientRegister)(h, client)
	})
	return h
}

// UnRegisterClient -> Removes the client from the Hub
func (h *Hub) UnRegisterClient(client *Client) *Hub {
	// Unregister the client from the index
	// Well if the client exists, we should trigger the event...

	if h.c.IsClientExist(client) {
		h.c.unregisterClient(client)
		// Calling an event with Goroutine
		go h.onClientUnRegister.Scan(func(k string, v interface{}) {
			v.(HubOnClientUnRegister)(h, client)
		})
	}
	return h
}

func (h *Hub) GetClientsByFilter(filter FindClientsFilter) map[uint64]*Client {
	return h.c.getClientsByFilter(filter)
}

// SetGetter  -> This is the function which gets the info and sends to the broadcaster!
func (h *Hub) SetGetter(getter HubGetter) bool {
	if !function.IsCallable(getter) {
		return false
	}
	h.getter = getter
	return true
}

func (h *Hub) GetCreatedTime() time.Time {
	return h.createdAt.Get()
}

func (h *Hub) GetNrOfClients() uint {
	return h.c.GetNrOfClients()
}

func (h *Hub) GetClients() map[*Client]bool {
	return h.c.GetClients()
}
