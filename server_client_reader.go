package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/kyaxcorp/go-helper/function"
	"github.com/kyaxcorp/go-websocket/msg"
	"github.com/rs/zerolog"
)

// readPump pumps messages from the websocket connection to the registrationHub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	info := func() *zerolog.Event {
		return c.LInfoF("readPump")
	}
	debug := func() *zerolog.Event {
		return c.LDebugF("readPump")
	}
	_error := func() *zerolog.Event {
		return c.LErrorF("readPump")
	}

	// Defer
	defer func() {
		_err := c.Disconnect()
		_error().Err(_err).Msg("disconnect error")
	}()

	info().Msg("running...")
	defer info().Msg("closing...")

	var _err error
	var messageType int
	var message []byte

	// Set options

	// Set the maximum message size for reading...
	c.conn.SetReadLimit(int64(c.server.MaxMessageSize.Get()))

	// Set the specific time when the read is not possible anymore and the connection
	// should be considered dead/corrupt
	_err = c.conn.SetReadDeadline(time.Now().Add(c.server.PongWait.Get()))
	// The error from read deadline can be ignored...
	if _err != nil {
		_error().
			Err(_err).
			Msg("set read deadline")
	}

	// Set the pong handler... this handler will be called on some reader will try to
	// call the Read/ReadMessage/NextReader...
	// if we receive a Pong Frame, this handler will be the one reacting on it!
	c.conn.SetPongHandler(func(string) error {
		// if we received a pong, it means, that the client reacts to pings

		// Set again the read deadline after receiving this frame...
		_err = c.conn.SetReadDeadline(time.Now().Add(c.server.PongWait.Get()))

		// The error from read deadline can be ignored...
		if _err != nil {
			_error().
				Err(_err).
				Msg("set read deadline")
		}

		c.lastReceivedPongTime.SetNow()
		c.nrOfReceivedPongs.Inc(1)
		debug().Uint64("pong_nr", c.nrOfReceivedPongs.Get()).Msg("pong received...")
		return nil
	})

	// This is when we receive a ping handler from the client...
	// in this case the server should also respond to these messages...
	c.conn.SetPingHandler(func(string) error {
		debug().Msg("we have received a ping frame from the client, responding with pong frame")
		_err = c.conn.SetWriteDeadline(time.Now().Add(c.server.WriteWait.Get()))
		if _err != nil {
			_error().Err(_err).Msg("SetPingHandler -> set write deadline")
		}

		// Sending the frame
		if _err = c.conn.WriteMessage(msg.Pong, nil); _err != nil {
			// Failed
			c.nrOfFailedSendPongs.Inc(1)
			_error().Err(_err).Msg("SetPingHandler -> failed to send pong")
			return nil
		}
		c.lastSentPongTime.SetNow()
		c.nrOfSentPongs.Inc(1)
		debug().Uint64("pong_nr", c.nrOfSentPongs.Get()).Msg("pong sent success")
		return nil
	})

	// Loop/Wait
	for {
		messageType, message, _err = c.conn.ReadMessage()
		// log.Println("Read a Message...")

		if _err != nil {
			_error().
				Err(_err).
				Int("message_type", messageType).
				Str("received_message", string(message[:])).
				Bytes("received_bytes", message).
				Msg("some error?! closing the connection!?")
			c.setAsClosed()
			if websocket.IsUnexpectedCloseError(_err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				_error().
					Msg("unexpected close error,expected other codes like 1001, 1006")
			}
			break
		}

		// Without a pointer
		receivedMessage := &ReceivedMessage{
			MessageType:   int8(messageType),
			MessageLength: uint(len(message)),
			Message:       message,
		}

		// TODO: should we process here the pong messages?!!

		// Message = bytes.TrimSpace(bytes.Replace(Message, newline, space, -1))

		// On Message Callback!
		// We decide in the callback how fast we want to process the incoming messages!

		// Declared on the server
		c.server.LEvent("start", "OnMessage", nil)
		c.server.onMessage.Scan(func(k string, v interface{}) {
			v.(OnMessage)(receivedMessage, c, c.server)
		})
		c.server.LEvent("finish", "OnMessage", nil)

		// Declared in the upgrader
		if function.IsCallable(c.onMessage) {
			c.onMessage(receivedMessage, c, c.server)
		}

		// We are sending the Message through Channel to broadcast! The registrationHub is another goroutine which sends to other
		// c the Message!

		// c.server.WSRegistrationHub.broadcast <- Message
	}
}
