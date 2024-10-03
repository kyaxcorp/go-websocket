package websocket

import (
	"time"

	"github.com/kyaxcorp/go-websocket/msg"
	//"GoCore/core/listeners/websocket/server/Message"
)

/*
1. send large text messages?!
2. send large files?!
3. How large text messages will be sent?!
4.
*/

type BinaryPayload struct {
	// This is the unique ID of the payload
	PayloadID string
	// This is chunk size based on what the data will be split!
	ChunkSize  int64
	OnResponse TextPayloadOnResponse
	// That's when we receive a response from the client with a specific chunk id!
	OnChunkResponse func()
	// This is the data
	Data []byte
}

type TextPayloadOnResponse func(response ReceivedMessage)

type TextPayloadStr struct {
	// TODO Generate time?
	// Fragment ID?! or Part ID
	// Size?!

	PayloadID string
	SentTime  time.Time
	Data      interface{}
}

type TextPayload struct {
	// This is the data which should be sent!
	Message interface{}
	// This when the payload has being sent successfully
	OnFinish TextPayloadOnResponse
	// On response of a specific part
	OnPartResponse func()
	// On json error
	OnJsonError OnJsonError
	// This is the size of the packet... the message will be split into parts
	ChunkSize int64

	// TODO: where do we save how much or where has being stopped sending?! we should save in a file
	// or we should save simply in the memory... someone should always save where the process has being stopped...
	// usually some of the parts should save that... the client or the server...
	// but because the server is usually always online, it's better that the server will save that info!
	// But even on reconnect the client needs to know where it has stopped... so, it also needs to know that?!
}

type FilePayload struct {
	// Complete file path for reading!
	FilePath string
	// Auto continue if connection interrupted
	AutoContinue bool

	// The size of the part should be sent!
	ChunkSize int64

	FileName string
}

//-------------CLIENT FUNCTIONS------------------\\

func (c *Client) WriteFile(filePayload FilePayload) *Client {
	/*
		1. choose what file is this... and read in parts...
		2. it should have a common db on local client where it saves the file that was being sending?! and also
		where it stopped in case of connection loss
		3.
	*/
	return c
}

// WriteBinaryPayload - This can be large files, audio data, anything... this is multiparted!
func (c *Client) WriteBinaryPayload(binaryPayload BinaryPayload) *Client {
	/*
		1. it should auto continue if connection failed!?...
		2. it should be able to send anything large....
		3. The client should be notified when the entire payload has being received?!
		4. But if we are doing live streaming?! it's important to
	*/

	return c
}

func (c *Client) LiveStreaming() *Client {
	/*
		1. It generates 1 payload id as identifier in the channel, the payload id never changes!
		2. It sends another field part_id, as what part has being sent!
		3. The idea is to know what is this and identifying when the client receives the data!
	*/

	return c
}

// WriteTextPayload - We will write any type of Message which will be formatted into a JSON and into a specific structure!
// This also will receive a response from the client!
// It's also limited to a specific packet length!
// It's destined to receive back a response from the Client!
func (c *Client) WriteTextPayload(textPayload TextPayload) *Client {
	/*
		1. Generate a Message ID
		2. Save the callback on receive!
		3. Execute callback on receive as a goroutine!
	*/

	// Converting to bytes

	if textPayload.Message == "" {
		return c
	}

	// Generate payload id
	payloadId := c.genPayloadID()

	// Create the payload
	payload := TextPayloadStr{
		PayloadID: payloadId,
		SentTime:  time.Now(),
		Data:      textPayload.Message,
	}

	encoded, err := msg.JsonToBytes(payload)
	if err != nil {
		if textPayload.OnJsonError != nil {
			textPayload.OnJsonError(err, payload)
		}
		return c
	}

	c.payloadMessageCallbackLock.Lock()
	// Saving the callback!
	c.payloadMessageCallbacks[payloadId] = textPayload.OnFinish
	c.payloadMessageCallbackLock.Unlock()

	// Sending the data!
	c.send <- encoded

	return c
}

func (c *Client) SendTextPayload(textPayload TextPayload) *Client {
	return c.SendTextPayload(textPayload)
}

// WriteText - It sends clear Text to the client! without any encoding!
func (c *Client) WriteText(message string) *Client {
	if message == "" {
		return c
	}
	go func() {
		c.send <- msg.TextToBytes(message)
	}()
	return c
}

func (c *Client) SendText(message string) *Client {
	return c.WriteText(message)
}

// WriteJSON - It sends Any structure to the client encoded as JSON!
func (c *Client) WriteJSON(message interface{}, onJsonError OnJsonError) *Client {
	go func() {
		encoded, err := msg.JsonToBytes(message)
		if err != nil {
			if onJsonError != nil {
				onJsonError(err, message)
			}
			return
		}
		c.send <- encoded
	}()
	return c
}
func (c *Client) SendJSON(message interface{}, onJsonError OnJsonError) *Client {
	return c.WriteJSON(message, onJsonError)
}

// WriteBinary - It sends clear bytes to the client
func (c *Client) WriteBinary(message []byte) *Client {
	if len(message) == 0 {
		return c
	}
	go func() {
		c.send <- msg.ToBinary(message)
	}()
	return c
}
func (c *Client) SendBinary(message []byte) *Client {
	return c.WriteBinary(message)
}

// BroadcastText - It sends clear Text to the c! without any encoding!
func (c *Client) BroadcastText(message string) *Client {
	if message == "" {
		return c
	}
	go func() {
		c.broadcastHub.broadcast <- msg.TextToBytes(message)
	}()
	return c
}

// BroadcastJSON - It sends Any structure to the c encoded as JSON!
func (c *Client) BroadcastJSON(message interface{}, onJsonError OnJsonError) *Client {
	go func() {
		encoded, err := msg.JsonToBytes(message)
		if err != nil {
			if onJsonError != nil {
				onJsonError(err, message)
			}
			return
		}
		c.broadcastHub.broadcast <- encoded
	}()
	return c
}

// BroadcastBinary -> It sends clear bytes to the c
func (c *Client) BroadcastBinary(message []byte) *Client {
	if len(message) == 0 {
		return c
	}
	go func() {
		c.broadcastHub.broadcast <- msg.ToBinary(message)
	}()
	return c
}
