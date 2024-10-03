package websocket

// It stops the hub!
func (h *Hub) Stop() *Hub {
	// If it's not running then return
	if !h.isRunning.Get() {
		return h
	}
	h.StopCalled.True()

	// Remove itself from the server stack!
	if _, ok := h.server.Hubs[h]; ok {
		delete(h.server.Hubs, h)
	}

	// Call through Channel the Stop command!
	//go func() {
	h.ctx.Cancel()

	//h.stopBroadcaster <- true
	//h.stopController <- true
	//h.stopGetter <- true
	//}()

	return h
}
