package websocket

import "github.com/kyaxcorp/go-helper/function"

func (h *Hub) OnStart(onStart HubOnStart) bool {
	if !function.IsCallable(onStart) {
		return false
	}
	h.onStart = onStart
	return true
}

func (h *Hub) OnStartGetter(onStartGetter HubOnStartGetter) bool {
	if !function.IsCallable(onStartGetter) {
		return false
	}
	h.onStartGetter = onStartGetter
	return true
}

func (h *Hub) OnStartBroadcast(onStartBroadCast HubOnStartBroadCast) bool {
	if !function.IsCallable(onStartBroadCast) {
		return false
	}
	h.onStartBroadCast = onStartBroadCast
	return true
}

//--------------ON CLIENT REGISTER-------------\\

func (h *Hub) OnClientRegisterRemove(name string) {
	h.onClientRegister.Del(name)
}

func (h *Hub) OnClientRegister(name string, callback HubOnClientRegister) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	h.onClientRegister.Set(name, callback)
	return true
}

func (h *Hub) HasOnClientRegister(name string) bool {
	return h.onClientRegister.Has(name)
}

//--------------ON CLIENT REGISTER-------------\\

//

//--------------ON CLIENT UNREGISTER-------------\\

func (h *Hub) OnClientUnRegisterRemove(name string) {
	h.onClientUnRegister.Del(name)
}

func (h *Hub) OnClientUnRegister(name string, callback HubOnClientUnRegister) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	h.onClientUnRegister.Set(name, callback)
	return true
}

func (h *Hub) HasOnClientUnRegister(name string) bool {
	return h.onClientUnRegister.Has(name)
}

//--------------ON CLIENT UNREGISTER-------------\\
