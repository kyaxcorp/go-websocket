package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

/*
 *  %x0 denotes a continuation frame

 *  %x1 denotes a text frame

 *  %x2 denotes a binary frame

 *  %x3-7 are reserved for further non-control frames

 *  %x8 denotes a connection close

 *  %x9 denotes a ping

 *  %xA denotes a pong

 *  %xB-F are reserved for further control frames
 */

// It returns a string
func (r *ReceivedMessage) Text() string {
	// Convert bytes to String
	//Message := bytes.TrimSpace(bytes.Replace(r.Message, newline, space, -1))
	//return Message
	//return string(Message[:])
	return string(r.Message[:])
}

// It gives the Binary!
func (r *ReceivedMessage) Binary() []byte {
	return r.Message
}

// It decodes JSON into a Structure!
func (r *ReceivedMessage) JSONDecode() (interface{}, error) {
	var tmp interface{}
	err := json.Unmarshal(r.Message, &tmp)
	if err != nil {
		return nil, err
	}
	return tmp, nil
}

func (r *ReceivedMessage) JSONDecodeTo(to interface{}) error {
	return json.Unmarshal(r.Message, to)
}

func (r *ReceivedMessage) IsText() bool {
	if r.MessageType == websocket.TextMessage {
		return true
	}
	return false
}

func (r *ReceivedMessage) IsBinary() bool {
	if r.MessageType == websocket.BinaryMessage {
		return true
	}
	return false
}

func (r *ReceivedMessage) IsContinuation() bool {
	if r.MessageType == 0 {
		return true
	}
	return false
}

func (r *ReceivedMessage) IsClose() bool {
	if r.MessageType == websocket.CloseMessage {
		return true
	}
	return false
}

func (r *ReceivedMessage) IsPing() bool {
	if r.MessageType == websocket.PingMessage {
		return true
	}
	return false
}

func (r *ReceivedMessage) IsPong() bool {
	if r.MessageType == websocket.PongMessage {
		return true
	}
	return false
}
