package websocket

func (h *Hub) runController() {
	for {
		if h.StopCalled.Get() {
			break
		}
		select {
		//case <-h.stopController: // TODO: deprecated, remove it!
		//	break
		case <-h.ctx.Done():
			break
		case client := <-h.UnregisterClientChannel:
			h.UnRegisterClient(client)
		}
	}
}
