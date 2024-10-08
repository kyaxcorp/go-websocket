package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/kyaxcorp/go-websocket/msg"
)

func (h *Hub) BroadcastTextTo(message string, to FindClientsFilter) *Hub {
	go func() {
		clients := h.c.getClientsByFilter(to)
		if len(clients) > 0 {
			h.broadcastTo <- hubBroadcastTo{
				to:   clients,
				data: msg.TextToBytes(message),
			}
		}
	}()
	return h
}

func (h *Hub) BroadcastText(message string) map[*Client]SendStatus {
	h.broadcast <- msg.TextToBytes(message)
	return <-h.broadcastStatus
}

func (h *Hub) BroadcastByReceivedMessageType(message *ReceivedMessage) map[*Client]SendStatus {
	switch message.MessageType {
	case websocket.TextMessage:
		h.broadcast <- msg.TextBytesToBytes(message.Message)
	case websocket.BinaryMessage:
		h.broadcast <- msg.ToBinary(message.Message)
	}
	return <-h.broadcastStatus
}

func (h *Hub) BroadcastByReceivedMessageTypeTo(message *ReceivedMessage, to FindClientsFilter) *Hub {
	go func() {
		clients := h.c.getClientsByFilter(to)
		if len(clients) > 0 {
			var convMsg []byte
			switch message.MessageType {
			case websocket.TextMessage:
				convMsg = msg.TextBytesToBytes(message.Message)
			case websocket.BinaryMessage:
				convMsg = msg.ToBinary(message.Message)
			}
			h.broadcastTo <- hubBroadcastTo{
				to:   clients,
				data: convMsg,
			}
		}
	}()
	return h
}

func (h *Hub) BroadcastJSON(message interface{}, onJsonError OnJsonError) (map[*Client]SendStatus, error) {
	encoded, err := msg.JsonToBytes(message)
	if err != nil {
		return nil, err
	}
	h.broadcast <- encoded
	return <-h.broadcastStatus, nil
}

func (h *Hub) BroadcastBinary(message []byte) map[*Client]SendStatus {
	h.broadcast <- msg.ToBinary(message)
	return <-h.broadcastStatus
}
