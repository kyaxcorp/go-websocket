package websocket

import (
	"github.com/kyaxcorp/go-helper/function"
)

func (s *Server) OnStart(name string, callback OnStart) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onStart.Set(name, callback)
	return true
}

func (s *Server) OnBeforeStart(name string, callback OnStart) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onBeforeStart.Set(name, callback)
	return true
}

func (s *Server) OnStarted(name string, callback OnStart) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onStarted.Set(name, callback)
	return true
}

func (s *Server) OnStartRemove(name string) {
	s.onStart.Del(name)
}

func (s *Server) OnBeforeStartRemove(name string) {
	s.onBeforeStart.Del(name)
}

func (s *Server) OnStartedRemove(name string) {
	s.onStarted.Del(name)
}

func (s *Server) OnBeforeStop(name string, callback OnStop) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onBeforeStop.Set(name, callback)
	return true
}

func (s *Server) OnStop(name string, callback OnStop) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onStop.Set(name, callback)
	return true
}

func (s *Server) OnStopped(name string, callback OnStop) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onStopped.Set(name, callback)
	return true
}

func (s *Server) OnBeforeStopRemove(name string) {
	s.onBeforeStop.Del(name)
}

func (s *Server) OnStopRemove(name string) {
	s.onStop.Del(name)
}

func (s *Server) OnStoppedRemove(name string) {
	s.onStopped.Del(name)
}

func (s *Server) OnMessage(name string, callback OnMessage) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onMessage.Set(name, callback)
	return true
}

func (s *Server) OnMessageRemove(name string) {
	s.onMessage.Del(name)
}

func (s *Server) OnBeforeUpgrade(name string, callback OnBeforeUpgrade) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onBeforeUpgrade.Set(name, callback)
	return true
}

func (s *Server) OnBeforeUpgradeRemove(name string) {
	s.onBeforeUpgrade.Del(name)
}

func (s *Server) OnConnect(name string, callback OnConnect) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onConnect.Set(name, callback)
	return true
}

func (s *Server) OnConnectRemove(name string) {
	s.onConnect.Del(name)
}

func (s *Server) OnClose(name string, callback OnClose) bool {
	if !function.IsCallable(callback) || name == "" {
		return false
	}
	s.onClose.Set(name, callback)
	return true
}

func (s *Server) OnCloseRemove(name string) {
	s.onClose.Del(name)
}
