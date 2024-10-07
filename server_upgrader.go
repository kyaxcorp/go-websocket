package websocket

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kyaxcorp/go-helper/_context"
	"github.com/kyaxcorp/go-helper/file"
	"github.com/kyaxcorp/go-helper/filesystem"
	"github.com/kyaxcorp/go-helper/function"
	"github.com/kyaxcorp/go-helper/sync/_bool"
	"github.com/kyaxcorp/go-helper/sync/_map_string_interface"
	"github.com/kyaxcorp/go-helper/sync/_time"
	"github.com/kyaxcorp/go-helper/sync/_uint64"
	"github.com/kyaxcorp/go-http/middlewares/authentication"
	"github.com/kyaxcorp/go-http/middlewares/connection"
	"github.com/kyaxcorp/go-logger"
)

// UpgradeToWS -> You can call it from Http Server before the connection has being initialized
func (s *Server) UpgradeToWS(
	c *gin.Context,
	onMessage OnMessage,
	onUpgrade OnUpgrade,
) bool {
	w := c.Writer
	r := c.Request

	// when we will do the upgrade, the http context will not remain, so we need to take
	// all we need from it

	s.LInfo().Msg("upgrading to websocket")

	// Before Upgrade
	s.onBeforeUpgrade.Scan(func(k string, v interface{}) {
		v.(OnBeforeUpgrade)(s)
	})

	// we should gather info about the client before the upgrade...even the context
	// should be copied after the upgrade!
	var _authDetails *authentication.AuthDetails
	var _connDetails *connection.ConnDetails

	// Generate an ID for this connection
	connectionID := s.genConnID()

	// GET and GET Authentication Details!
	_authDetails = authentication.GetAuthDetailsFromCtx(c)

	// Set the connection Details
	_connDetails = connection.GetConnectionDetailsFromCtx(c)

	safeHttpContext := c.Copy()
	clientIP := c.ClientIP()

	// Upgrade the http connection to websocket!
	conn, errUpgrade := s.WSUpgrader.Upgrade(w, r, nil)
	if errUpgrade != nil {
		//log.Println(err)
		s.LError().Err(errUpgrade).Msg("failed to upgrade client to websocket")
		return false
	}
	// log.Println("Creating client")

	// TODO: extract http auth token!

	// TODO: extract Device UUID
	// TODO: or simply get the middleware data!

	/*
		Try logging clients as per connection ID? or by other authentication details?!
		Save the logs in separate files!
		Enable/Disable saving logs in te same file as websocket!

	*/

	// Based on the identifier, we will create the logs, and we will identify in much easier way the client
	var identifiedBy string
	clientIdentifier := ""
	clientIPFiltered := clientIP
	isIpv6 := false
	switch clientIPFiltered {
	// TODO: maybe we should remove this...
	case "::1":
		isIpv6 = true
		clientIPFiltered = "localhost"
	case "127.0.0.1":
		clientIPFiltered = "localhost"
	}

	// Check if it's ipv6 by checking :
	// If there are any :, replace them with dots
	if strings.Contains(clientIPFiltered, ":") || isIpv6 {
		identifiedBy = "ipv6"
		clientIPFiltered = "ipv6." + strings.ReplaceAll(clientIPFiltered, ":", ".")
	} else {
		identifiedBy = "ipv4"
		clientIPFiltered = "ipv4." + clientIPFiltered
	}

	if _authDetails.DeviceDetails.DeviceID != "" {
		// Get the device id
		identifiedBy = "device_id"
		//clientIdentifier = "device_id_" + conv.UInt64ToStr(_authDetails.DeviceDetails.DeviceID)
		clientIdentifier = "device_id_" + _authDetails.DeviceDetails.DeviceID
	} else if _authDetails.DeviceDetails.DeviceUUID != "" {
		// Get by the device uuid
		identifiedBy = "device_uuid"
		clientIdentifier = "device_uuid_" + _authDetails.DeviceDetails.DeviceUUID
	} else if _authDetails.UserDetails.UserID != "" {
		identifiedBy = "user_id"
		//clientIdentifier = "user_id_" + conv.UInt64ToStr(_authDetails.UserDetails.UserID)
		clientIdentifier = "user_id_" + _authDetails.UserDetails.UserID
	} else if clientIPFiltered != "" {
		// Get the IP Address only...
		clientIdentifier = clientIPFiltered
	}

	// create a context for the client!
	clientCtx := _context.WithCancel(s.ctx.Context())

	// Can be IPv4 & IPv6, so we should take care of that very wisely
	// The best way to identify the connection is by device id, or something unique that identifies it!
	// We can check sum all the metadata, but after that we will not understand the folder path name
	loggerConfig := s.Logger.Config

	// Set as reference the parent logger!
	loggerConfig.ParentWriter = s.Logger.MainWriter

	// Get the clients log path
	clientsLogPath := s.GetClientsLogPath()
	// Creating clients path
	loggerConfig.DirLogPath = file.FilterPath(clientsLogPath + filesystem.DirSeparator() +
		identifiedBy + filesystem.DirSeparator() + clientIdentifier + filesystem.DirSeparator())
	// Set the name
	loggerConfig.Name = clientIdentifier

	clientLogger := logger.New(loggerConfig)
	// Generate a sub context logger
	subLogger := clientLogger.Logger.With().
		Uint64("connection_id", connectionID).
		Str("ip_address", clientIPFiltered).
		Logger()
	// set back to logger
	clientLogger.Logger = &subLogger

	client := &Client{
		parentCtx: s.ctx.Context(),
		ctx:       clientCtx,

		// Logger
		Logger: clientLogger,

		// Connect Time
		connectTime: time.Now(),

		// Generate connection ID
		connectionID: connectionID,

		// Ping info
		// Ping Send

		lastSendPingTry:     _time.NewNow(),
		lastSentPingTime:    _time.NewNow(),
		nrOfSentPings:       _uint64.NewVal(0),
		nrOfFailedSendPings: _uint64.NewVal(0),

		// Ping Receive
		lastSentPongTime:     _time.NewNow(),
		lastReceivedPongTime: _time.NewNow(),
		nrOfReceivedPongs:    _uint64.NewVal(0),
		nrOfSentPongs:        _uint64.NewVal(0),
		nrOfFailedSendPongs:  _uint64.NewVal(0),

		isClosed: _bool.New(),

		// Gin Context
		httpContext:     c,
		safeHttpContext: safeHttpContext,

		// registrationHub:  s.WSRegistrationHub,
		onMessage:       onMessage,
		server:          s,
		conn:            conn,
		registrationHub: s.WSRegistrationHub,
		broadcastHub:    s.WSBroadcastHub,
		isDisconnecting: _bool.NewVal(false),

		nrOfSentMessages:        _uint64.New(),
		nrOfSentFailedMessages:  _uint64.New(),
		nrOfSentSuccessMessages: _uint64.New(),

		// Creating the channel for sending messages
		//send: make(chan []byte, 256),

		closeWritePump: make(chan bool),

		closeCode:    DefaultCloseCode,
		closeMessage: DefaultCloseReason,

		//send: make(chan []byte, maxMessageSize),
		send:       make(chan []byte),
		sendStatus: make(chan SendStatus),
		// Custom Data Array create!
		//customData: make(map[string]interface{}),
		customData: _map_string_interface.New(),

		connDetails: _connDetails,
		authDetails: _authDetails,
	}

	// log.Println("Before on connect")

	// log.Println("Registering client")

	// Register the client!
	client.server.WSRegistrationHub.register <- client

	// log.Println("Creating buffers")

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.

	// This is the writer...!
	// If we want to send something, it's better for us to create a local
	// buffer which will be read automatically from this process!
	// We should have one process handling the messaging process!
	go client.writePump()

	// Same thing should be for the reader! This process should
	// Mainly handle the receiving data, and if needed create a separate process
	// to work on the data!
	go client.readPump() // This is the reader!

	// log.Println("connection started!")

	// On Connect Callback!
	s.onConnect.Scan(func(k string, v interface{}) {
		v.(OnConnect)(client, s)
	})

	s.LInfo().
		Uint64("connection_id", client.connectionID).
		Msg("new connection created")

	// On Upgrade callback!
	if function.IsCallable(onUpgrade) {
		s.LEvent("start", "onUpgrade", nil)
		go onUpgrade(client, s)
		s.LEvent("finish", "onUpgrade", nil)
	}

	return true
}
