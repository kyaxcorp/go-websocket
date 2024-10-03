package msg

import "github.com/gorilla/websocket"

const (
	Binary = websocket.BinaryMessage
	Text   = websocket.TextMessage
	Ping   = websocket.PingMessage
	Pong   = websocket.PongMessage
	Close  = websocket.CloseMessage
)
