package websocket

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kyaxcorp/go-http/middlewares/authentication"
	"github.com/kyaxcorp/go-websocket/msg"
	"github.com/rs/zerolog"
)

func (c *Client) GetConnectTime() time.Time {
	return c.connectTime
}

func (c *Client) GetConnectedTimeSeconds() int64 {
	// TODO : calculate
	return 1
}

func (c *Client) setAsClosed() {
	// Setting as connection closed!
	c.isClosed.True()
}

// DisconnectGracefully -> set 0 and "" for default values!
func (c *Client) DisconnectGracefully(code uint16, message string) {
	// We send to the client that we want to close the connection!
	// And we should receive response back, and after that the disconnect will be called!

	c.setAsClosed()

	if code > 0 {
		c.closeCode = code
	}
	if message != "" && len(message) > 0 {
		c.closeMessage = message
	}
	// send through channel!
	c.send <- []byte(strconv.Itoa(msg.Close))
}

// Disconnect the client forcefully!!
func (c *Client) Disconnect() error {
	info := func() *zerolog.Event {
		return c.LInfoF("Disconnect")
	}
	warn := func() *zerolog.Event {
		return c.LWarnF("Disconnect")
	}
	c.ctx.Cancel()

	/*_error := func() *zerolog.Event {
		return c.LErrorF("Disconnect")
	}*/
	info().Msg("calling...")
	defer info().Msg("leaving...")

	if c.IsDisconnecting() {
		warn().Msg("already disconnecting...")
		return nil
	}
	c.setAsClosed()

	// On Close callback
	c.server.onClose.Scan(func(k string, v interface{}) {
		v.(OnClose)(c, c.server)
	})

	info().Msg("closing the client connection...")

	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
	// TODO: check if channel closed
	c.closeWritePump <- true // Close the write pump!
	if c.isDisconnecting != nil {
		c.isDisconnecting.True()
	}
	c.server.WSRegistrationHub.unregister <- c
	// This is a force Close!
	return c.conn.Close()
}

func (c *Client) IsDisconnecting() bool {
	return c.isDisconnecting.Get()
}

func (c *Client) GetConnectionID() uint64 {
	return c.connectionID
}

func (c *Client) GetHttpContext() *gin.Context {
	return c.httpContext
}

func (c *Client) GetSafeHttpContext() *gin.Context {
	return c.safeHttpContext
}

func (c *Client) GetDeviceID() string {
	return c.authDetails.DeviceDetails.DeviceID
}

func (c *Client) GetDeviceUUID() string {
	return c.authDetails.DeviceDetails.DeviceUUID
}

func (c *Client) GetUserID() string {
	return c.authDetails.UserDetails.UserID
}

func (c *Client) GetAuthToken() string {
	return c.authDetails.AuthTokenDetails.Token
}

func (c *Client) GetAuthTokenID() string {
	return c.authDetails.AuthTokenDetails.TokenID
}

func (c *Client) GetIPAddress() string {
	return c.connDetails.ClientIPAddress
}

func (c *Client) GetRemoteIP() string {
	return c.connDetails.RemoteIP
}

func (c *Client) GetRequestPath() string {
	return c.connDetails.RequestPath
}

func (c *Client) GetTokenExpirationTime() time.Time {
	return c.authDetails.AuthTokenDetails.ExpireDate
}

func (c *Client) GetAuthDetails() *authentication.AuthDetails {
	return c.authDetails
}

// Set custom Data to client connection!
func (c *Client) Set(key string, value interface{}) *Client {
	//c.customData[key] = value
	c.customData.Set(key, value)
	return c
}

// Get custom data from the client connection!
func (c *Client) Get(key string) interface{} {
	return c.customData.Get(key)
	/*if val, ok := c.customData[key]; ok {
		return val
	}
	return nil*/
}
