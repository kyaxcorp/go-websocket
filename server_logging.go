package websocket

import (
	"github.com/rs/zerolog"
)

// LDebug -> 0
func (s *Server) LDebug() *zerolog.Event {
	return s.Logger.Debug()
}

// LInfo -> 1
func (s *Server) LInfo() *zerolog.Event {
	return s.Logger.Info()
}

// LWarn -> 2
func (s *Server) LWarn() *zerolog.Event {
	return s.Logger.Warn()
}

// LError -> 3
func (s *Server) LError() *zerolog.Event {
	return s.Logger.Error()
}

// LFatal -> 4
func (s *Server) LFatal() *zerolog.Event {
	return s.Logger.Fatal()
}

// LPanic -> 5
func (s *Server) LPanic() *zerolog.Event {
	return s.Logger.Panic()
}

//

//------------------------------------\\

func (s *Server) LEvent(eventType string, eventName string, beforeMsg func(event *zerolog.Event)) {
	s.Logger.InfoEvent(eventType, eventName, beforeMsg)
}

//

//-------------------------------------\\

//

// LWarnF -> when you need specifically to indicate in what function the logging is happening
func (s *Server) LWarnF(functionName string) *zerolog.Event {
	return s.Logger.WarnF(functionName)
}

// LInfoF -> when you need specifically to indicate in what function the logging is happening
func (s *Server) LInfoF(functionName string) *zerolog.Event {
	return s.Logger.InfoF(functionName)
}

// LDebugF -> when you need specifically to indicate in what function the logging is happening
func (s *Server) LDebugF(functionName string) *zerolog.Event {
	return s.Logger.DebugF(functionName)
}

// LErrorF -> when you need specifically to indicate in what function the logging is happening
func (s *Server) LErrorF(functionName string) *zerolog.Event {
	return s.Logger.ErrorF(functionName)
}
