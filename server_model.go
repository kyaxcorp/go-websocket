package websocket

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kyaxcorp/go-helper/_context"
	"github.com/kyaxcorp/go-helper/sync/_bool"
	"github.com/kyaxcorp/go-helper/sync/_map_string_interface"
	"github.com/kyaxcorp/go-helper/sync/_time"
	"github.com/kyaxcorp/go-helper/sync/_uint64"
	"github.com/kyaxcorp/go-helper/sync/duration"
	"github.com/kyaxcorp/go-logger/model"
)

type OnJsonError func(err error, message interface{})

type OnEvent func(...interface{})

type OnBeforeUpgrade func(s *Server)
type OnUpgrade func(c *Client, s *Server)

// On Connect it will be launched in a goroutine!
type OnConnect func(c *Client, s *Server)
type OnMessage func(message *ReceivedMessage, c *Client, s *Server)
type OnClose func(c *Client, s *Server)

// Stop
type OnStop func(s *Server)
type OnBeforeStop func(s *Server)
type OnStopped func(s *Server)

// Start
type OnStart func(s *Server)
type OnBeforeStart func(s *Server)
type OnStarted func(s *Server)

/*type Client struct {
	Message   string
	ginClient *gin.Context
}*/

type Server struct {
	Name        string
	Description string

	// This is the original config from the constructor
	config Config

	connectionID *_uint64.Uint64

	isStopCalled  *_bool.Bool
	isStopped     *_bool.Bool
	isStartCalled *_bool.Bool
	isStarted     *_bool.Bool

	// Starting time of the server
	startTime *_time.Time
	// Stop time of the server
	stopTime *_time.Time

	// This will be the main folder where we will store the logs
	LoggerDirPath string
	// This is the logger configuration!
	Logger *model.Logger

	// Enables Server Status through HTTP
	enableServerStatus *_bool.Bool
	// These are the server status credentials
	statusUsername string
	statusPassword string

	// enableUnsecure -> most of the time is readonly!
	enableUnsecure bool // Enable unsecure connections
	//
	enableSSL   bool
	sslCertPath string
	sslKeyPath  string

	// It also includes port
	ListeningAddresses    []string // This is for unencrypted
	ListeningAddressesSSL []string // This is for encrypted
	// Context
	parentCtx context.Context
	ctx       *_context.CancelCtx

	// This is GIN Server
	WSServer *gin.Engine
	// This is the registrationHub that registers the c...
	WSRegistrationHub *RegistrationHub
	// This is the registrationHub which sends broadcast messages!
	WSBroadcastHub *BroadcastHub
	// This is the upgrader which transfers from http to WebSocket
	WSUpgrader websocket.Upgrader

	// These settings are being set into the Upgrader
	readBufferSize    *_uint64.Uint64
	writeBufferSize   *_uint64.Uint64
	enableCompression *_bool.Bool

	// It enables automatic upgrade from http to ws
	// Usually the server will be behind a proxy server, so it's not necessary....
	EnableHttpToWSUpgrade *_bool.Bool

	// Events/Callbacks - they are common for all routes!
	onBeforeUpgrade *_map_string_interface.MapStringInterface
	onConnect       *_map_string_interface.MapStringInterface
	onMessage       *_map_string_interface.MapStringInterface
	onClose         *_map_string_interface.MapStringInterface

	// Stop
	onStop       *_map_string_interface.MapStringInterface
	onBeforeStop *_map_string_interface.MapStringInterface
	onStopped    *_map_string_interface.MapStringInterface

	// Start
	onStart       *_map_string_interface.MapStringInterface
	onBeforeStart *_map_string_interface.MapStringInterface
	onStarted     *_map_string_interface.MapStringInterface

	// ------Settings ---------\\
	// Time allowed to write a Message to the peser.
	//WriteWait time.Duration
	WriteWait *duration.Duration
	// Time allowed to read the next pong Message from the peer.
	PongWait *duration.Duration
	// send pings to peer with this period. Must be less than pongWait.
	PingPeriod *duration.Duration
	// Maximum Message size allowed from peer.
	MaxMessageSize *_uint64.Uint64

	// ------Settings ---------\\

	// Here we store the active/registered ClientsStatus (Connections)
	// c      map[*Client]bool
	// clientsIndex ClientsIndex
	c *clientsData

	// Here we store the created Hubs
	// The Hubs are being added here when they are started only!
	// There is no reference until they are started!
	Hubs map[*Hub]bool
}

func (s *Server) genConnID() uint64 {
	return s.connectionID.ResetMaxAndInc(1, 1000000, 0)
}

type ReceivedMessage struct {
	// There are few types...
	MessageType int8
	// Bytes?!!
	MessageLength uint
	Message       []byte
}
