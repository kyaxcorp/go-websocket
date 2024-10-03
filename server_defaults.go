package websocket

import (
	"time"
)

const (
	// Time allowed to write a Message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong Message from the peer.
	PongWait = 60 * time.Second

	// send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum Message size allowed from peer. (By default its 512)
	MaxMessageSize = 128 * 1024 * 1024 // This is approx 134 MB?

	// Enable Automatic HTTP Upgrade to WS
	EnableHttpToWebSocketUpgrade = true
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const DefaultCloseCode = 1000
const DefaultCloseReason = "No specific reason!"
const DefaultListeningAddress = "0.0.0.0:8080"
const DefaultSSLListeningAddress = "0.0.0.0:8443"
const DefaultReadBufferSize = 1024
const DefaultWriteBufferSize = 1024
const DefaultEnableCompression = true
