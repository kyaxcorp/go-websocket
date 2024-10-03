package websocket

import (
	"io"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kyaxcorp/go-websocket/msg"
	"github.com/rs/zerolog"
)

// writePump pumps messages from the registrationHub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	info := func() *zerolog.Event {
		return c.LInfoF("writePump")
	}
	debug := func() *zerolog.Event {
		return c.LDebugF("writePump")
	}
	warn := func() *zerolog.Event {
		return c.LWarnF("writePump")
	}
	_error := func() *zerolog.Event {
		return c.LErrorF("writePump")
	}
	// Defer
	defer func() {
		err := c.Disconnect()
		_error().Err(err).Msg("disconnect error")
	}()

	info().Msg("running...")
	defer info().Msg("closing...")

	// launch a ticker which will ping...
	// we should also check if pong received... if not... then close the connection
	c.pingTicker = time.NewTicker(c.server.PingPeriod.Get())
	calledClose := false

	var err error
	var messageType int
	var w io.WriteCloser

	for {
		select {
		// We called the close pump!
		case calledClose = <-c.closeWritePump:
			if calledClose {
				break
			}
		case message, ok := <-c.send:
			// log.Println("Message received to channel!")

			// Well, if we send an empty... message it can give false! so in this case the connection will be closed...
			// it should always be called!
			err = c.conn.SetWriteDeadline(time.Now().Add(c.server.WriteWait.Get()))

			if err != nil {
				_error().Err(err).Msg("set write deadline")
			}

			if !ok {
				// The registrationHub closed the channel.
				// TODO: we should see how we handle this type of error!!! this is an internal problem!
				// We should somehow close the connection!
				// In this case, if we are sending close message, the connection will be closed gracefully!
				// And even if not closed gracefully, the ping/pong mechanism will handle the link and it will close it after all!

				closeCode := int(c.closeCode)
				closeMsg := c.closeMessage
				if c.closeCode == 0 {
					closeCode = 1011 // Internal server error
				}
				if closeMsg == "" {
					closeMsg = "ws server has an internal error"
				}

				err = c.conn.WriteMessage(
					msg.Close,
					websocket.FormatCloseMessage(closeCode, closeMsg),
				)
				if err != nil {
					_error().Err(err).Msg("failed to send close code")
				}
				return
			}

			// TODO: check if the the type is right!

			messageType, err = strconv.Atoi(string(message[0]))

			switch messageType {
			case msg.Close:
				c.setAsClosed()
				info().Msg("sending close message for graceful disconnect")
				// https://developer.mozilla.org/en-US/docs/Web/API/CloseEvent

				closeCode := int(c.closeCode)
				closeMsg := c.closeMessage
				if c.closeCode == 0 {
					closeCode = 1000 // Normal closure
				}
				if closeMsg == "" {
					closeMsg = "you have being disconnected gracefully"
				}

				if err = c.conn.WriteMessage(
					msg.Close,
					websocket.FormatCloseMessage(closeCode, closeMsg),
				); err != nil {
					return
				}
				info().Msg("continuing...")
				continue
			}

			if err != nil {
				warn().Msg("incorrect message type")
				return
			}

			//log.Println("Getting the writer!", Message[0], messageType)
			// log.Println("send Message ", Message)

			// Choose a Writer by type! (we have defined that first byte is the Message Type!)
			w, err = c.conn.NextWriter(messageType)
			if err != nil {
				return
			}
			// log.Println("Writing the Message")

			// Count
			c.nrOfSentMessages.Inc(1)

			// Writing the message to the client
			// We exclude the first byte because it includes the messageType
			_, err = w.Write(message[1:])
			if err != nil {
				_error().Err(err).Msg("failed write bytes")
				// Count
				c.nrOfSentFailedMessages.Inc(1)
				// Even if we failed to write... we should close the writer!
				//return
			} else {
				// Count
				c.nrOfSentSuccessMessages.Inc(1)
			}

			// Closing the writer
			if err = w.Close(); err != nil {
				_error().Err(err).Msg("failed to close the writer")
				return
			}
		case <-c.pingTicker.C:
			// Check when was the last time we have received a pong message...
			// If we haven't received for a specific amount of time the pong message
			// we should close automatically the connection!
			debug().Msg("preparing to send ping")
			if c.lastReceivedPongTime.Get().Unix()-180 > c.lastSentPingTime.Get().Unix() {
				// It's dead...or the client simply doesn't send any pongs...
				// close the connection...
				_error().Err(err).Msg("ping ticker -> haven't receive for 3 minutes any pong messages from the client... closing now...")
				err = c.Disconnect()
				warn().Msg(err.Error())
				return
			}

			// it should always be called!
			err = c.conn.SetWriteDeadline(time.Now().Add(c.server.WriteWait.Get()))
			if err != nil {
				_error().Err(err).Msg("ping ticker -> set write deadline")
			}

			c.lastSendPingTry.SetNow()
			if err = c.conn.WriteMessage(msg.Ping, nil); err != nil {
				// Failed
				_error().Err(err).Msg("ping ticker -> failed to send ping")
				c.nrOfFailedSendPings.Inc(1)
				return
			}

			// Success
			c.lastSentPingTime.SetNow()
			c.nrOfSentPings.Inc(1)
			debug().Uint64("ping_nr", c.nrOfSentPings.Get()).Msg("ping sent success")
		}

		if calledClose {
			//log.Println("WE WANT TO CLOSE!!!")
			break
		}
	}
}
