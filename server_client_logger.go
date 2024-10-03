package websocket

import (
	"github.com/rs/zerolog"
)

// LDebug -> 0
func (c *Client) LDebug() *zerolog.Event {
	return c.Logger.Debug()
}

// LInfo -> 1
func (c *Client) LInfo() *zerolog.Event {
	return c.Logger.Info()
}

// LInfoF -> when you need specifically to indicate in what function the logging is happening
func (c *Client) LInfoF(functionName string) *zerolog.Event {
	return c.Logger.InfoF(functionName)
}

func (c *Client) LDebugF(functionName string) *zerolog.Event {
	return c.Logger.DebugF(functionName)
}

// LWarn -> 2
func (c *Client) LWarn() *zerolog.Event {
	return c.Logger.Warn()
}

// LWarnF -> when you need specifically to indicate in what function the logging is happening
func (c *Client) LWarnF(functionName string) *zerolog.Event {
	return c.Logger.WarnF(functionName)
}

// LError -> 3
func (c *Client) LError() *zerolog.Event {
	return c.Logger.Error()
}

// LErrorF -> when you need specifically to indicate in what function the logging is happening
func (c *Client) LErrorF(functionName string) *zerolog.Event {
	return c.Logger.ErrorF(functionName)
}

// LFatal -> 4
func (c *Client) LFatal() *zerolog.Event {
	return c.Logger.Fatal()
}

// LPanic -> 5
func (c *Client) LPanic() *zerolog.Event {
	return c.Logger.Panic()
}

func (c *Client) LEvent(eventType string, eventName string, beforeMsg func(event *zerolog.Event)) {
	c.Logger.InfoEvent(eventType, eventName, beforeMsg)
}
