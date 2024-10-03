package websocket

import (
	"time"
)

func (s *Server) Stop() error {
	s.onBeforeStop.Scan(func(k string, v interface{}) {
		v.(OnBeforeStop)(s)
	})

	// Check if start is not running right now!
	if s.isStartCalled.Get() {
		// TODO: return error that start has being called right now!
		return nil
	}

	// Check if started
	if !s.isStarted.Get() {
		// Server it's not started to shutdown...
		// TODO: return an error that is not started!
		return nil
	}
	// Check if stop is called... if not then stop it!
	if s.isStopCalled.IfFalseSetTrue() {
		// Stop already has being called!
		// TODO: return error
		return nil
	}

	// Calling the existing callbacks!
	s.onStop.Scan(func(k string, v interface{}) {
		v.(OnStop)(s)
	})

	// Calling Cancel Function! it will send a signal!
	s.ctx.Cancel()

	s.stopTime.Set(time.Now())
	// Set that the server is stopped!
	s.isStopped.Set(true)
	// Set that stop command has being finished
	s.isStopCalled.Set(false)

	s.onStopped.Scan(func(k string, v interface{}) {
		v.(OnStopped)(s)
	})
	return nil
}

func (s *Server) IsStopping() bool {
	return s.isStopCalled.Get()
}

func (s *Server) IsStopped() bool {
	return s.isStopped.Get()
}
