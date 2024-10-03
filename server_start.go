package websocket

import (
	"crypto/tls"
	"flag"
	"net/http"
	"strings"

	"github.com/gookit/color"
	"github.com/kyaxcorp/go-helper/errors2/define"
	"github.com/kyaxcorp/go-helper/network/port"
	"github.com/rs/zerolog"
)

// Start THe user can start it in a goroutine
func (s *Server) Start() error {
	s.LInfo().Msg("entering start function")

	info := func() *zerolog.Event {
		return s.LInfoF("Start")
	}
	warn := func() *zerolog.Event {
		return s.LWarnF("Start")
	}
	_error := func() *zerolog.Event {
		return s.LErrorF("Start")
	}

	defer info().Msg("leaving start function")

	s.LEvent("start", "OnBeforeStart", nil)
	s.onBeforeStart.Scan(func(k string, v interface{}) {
		v.(OnBeforeStart)(s)
	})
	s.LEvent("finish", "OnBeforeStart", nil)

	// Check if stop command wasn't called right now...
	if s.isStopCalled.Get() {
		warn().Msg("stop already called")
		return define.Err(0, "stop already called")
	}

	// Check if the server is stopped
	if !s.isStopped.Get() {
		// The server is not stopped... so it cannot start
		warn().Msg("server already stopped")
		return define.Err(0, "server already stopped")
	}
	// Check if start command was called
	if s.isStartCalled.IfFalseSetTrue() {
		warn().Msg("start already called")
		return define.Err(0, "start already called")
	}

	s.LEvent("start", "OnStart", nil)
	s.onStart.Scan(func(k string, v interface{}) {
		v.(OnStart)(s)
	})
	s.LEvent("finish", "OnStart", nil)

	s.LInfo().Msg("creating withCancel context")
	// We create each time when we start the server!
	s.NewCancelContext()

	// Create WS Upgrader
	s.createWSUpgrader()

	// Launching registration goroutine
	info().Msg("launching registration hub")
	go s.WSRegistrationHub.run()
	// Launching broadcast goroutine
	info().Msg("launching broadcast hub")
	go s.WSBroadcastHub.run()

	info().Msg("creating http instances for secure and unsecure servers")
	// Standard instances
	instances := make(map[string]*http.Server)
	// SSL instances
	instancesSSL := make(map[string]*http.Server)

	// Creating non-secure http server
	if s.enableUnsecure {
		info().Msg("unsecure listening is enabled")
		for _, listeningAddress := range s.ListeningAddresses {
			var addr *string

			if listeningAddress == "" {
				continue
			}

			searchFreePort := false
			if strings.Contains(listeningAddress, "+") {
				searchFreePort = true
			}
			listeningAddress = port.FilterAddress(listeningAddress)

			if !searchFreePort {
				if busy, _err := port.IsTCPBusy(listeningAddress); busy {
					_error().
						Err(_err).
						Str("listening_address", listeningAddress).
						Msg("listening address already busy")
					continue
				}
			} else {
				newListeningAddress, _err := port.SearchAndLockFreeTCPAddress(listeningAddress)
				if _err != nil {
					_error().
						Err(_err).
						Str("new_listening_address", newListeningAddress).
						Str("listening_address", listeningAddress).
						Msg("listening address already busy")
					continue
				}
				if listeningAddress != newListeningAddress {
					warn().
						Str("new_listening_address", newListeningAddress).
						Str("listening_address", listeningAddress).
						Msg("auto binding is enabled, listening address has been changed")
				}

				listeningAddress = newListeningAddress
			}

			// Creating the address flag
			addr = flag.String("webSocketServer", listeningAddress, "Web Socket Server")
			// Saving the instance
			if addr == nil {
				_error().Str("listening_address", listeningAddress).Msg("addr is nil")
				continue
			}
			info().
				Str("listening_on", listeningAddress).
				Str("listening_addr", *addr).
				Msg("creating http instance")
			instances[listeningAddress] = &http.Server{
				Addr:    *addr,
				Handler: s.WSServer,
			}
		}
	}

	// Creating secure http server
	if s.enableSSL {
		info().Msg("secure listening is enabled")
		for _, listeningAddress := range s.ListeningAddressesSSL {
			var addrSSL *string

			if listeningAddress == "" {
				continue
			}

			searchFreePort := false
			if strings.Contains(listeningAddress, "+") {
				searchFreePort = true
			}
			listeningAddress = port.FilterAddress(listeningAddress)

			// //
			if !searchFreePort {
				if busy, _err := port.IsTCPBusy(listeningAddress); busy {
					_error().
						Err(_err).
						Str("listening_address", listeningAddress).
						Msg("listening address already busy")
					continue
				}
			} else {
				newListeningAddress, _err := port.SearchAndLockFreeTCPAddress(listeningAddress)
				if _err != nil {
					_error().
						Err(_err).
						Str("new_listening_address", newListeningAddress).
						Str("listening_address", listeningAddress).
						Msg("listening address already busy")
					continue
				}
				if listeningAddress != newListeningAddress {
					warn().
						Str("new_listening_address", newListeningAddress).
						Str("listening_address", listeningAddress).
						Msg("auto binding is enabled, listening address has been changed")
				}

				listeningAddress = newListeningAddress
			}

			addrSSL = flag.String("webSocketServerSSL", listeningAddress, "Web Socket Server SSL")
			if addrSSL == nil {
				_error().Str("listening_address", listeningAddress).Msg("addrSSL is nil")
				continue
			}
			info().
				Str("listening_ssl_on", listeningAddress).
				Str("listening_ssl_addr", *addrSSL).
				Msg("creating http(s) instance")

			cert, _err := tls.LoadX509KeyPair(s.sslCertPath, s.sslKeyPath)
			if _err != nil {
				_error().Str("listening_address", listeningAddress).Msg("failed to load certificates")
				continue
			}
			/*tlsConfig, _err := connhelpers.TlsConfigWithHttp2Enabled(&tls.Config{
				Certificates: []tls.Certificate{cert},
			})
			if _err != nil {
				_error().Str("listening_address", listeningAddress).Msg("failed to prepare config with http2 enabled")
				continue
			}*/
			// By leaving the auto-configuration, the app already will have http/2 enabled!

			instancesSSL[listeningAddress] = &http.Server{
				Addr:    *addrSSL,
				Handler: s.WSServer,
				TLSConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
				},
			}
		}
	}

	// This routine will handle termination of the server!
	go func() {
		terminate := false
		for {
			select {
			case <-s.ctx.Done():
				// TODO: Stop Listeners!
				//s.WSServer
				terminate = true
				info().Msg("terminating...")

				// Shutdown standard instances
				if s.enableUnsecure {
					for listeningAddress, instance := range instances {
						info().Str("shutting_down", listeningAddress).Msg("shutting down server")
						_err := instance.Shutdown(s.ctx.Context())
						if _err != nil {
							_error().Err(_err).Msg("failed shutting down websocket server")
						}
					}
				}
				// Shutdown SSL instances
				if s.enableSSL {
					for listeningAddress, instance := range instancesSSL {
						info().Str("shutting_down", listeningAddress).Msg("shutting down server")
						_err := instance.Shutdown(s.ctx.Context())
						if _err != nil {
							_error().Err(_err).Msg("failed shutting down websocket server")
						}
					}
				}
			}
			if terminate {
				break
			}
		}
	}()

	// Listen for unencrypted/plain connections
	if s.enableUnsecure {
		for listeningAddress, instance := range instances {
			info().Str("running_on", listeningAddress).Msg("running http server")
			go func(instance *http.Server) {
				// TODO: make a callback for fail listening!
				//err := s.WSServer.Run(*addr)
				_err := instance.ListenAndServe()

				if _err != nil {
					_error().Err(_err).Msg(color.Style{color.LightRed}.Render("failed to listen http server"))
				}
			}(instance)
		}
	}

	// Listen for SSL Connections
	if s.enableSSL {
		//TODO: SSL SERVER IS CPU CONSUMING!!!! even with no connections -> ONLY ON WINDOWS!!!!!
		for listeningAddress, instance := range instancesSSL {
			info().Str("running_on", listeningAddress).Msg("running http(s) server")
			go func(instance *http.Server) {
				// TODO: make a callback for fail listening!
				//log.Println("enabling SSL Connections")
				//err := s.WSServer.RunTLS(*addrSSL, s.sslCertPath, s.sslKeyPath)
				//_err := instance.ListenAndServeTLS(s.sslCertPath, s.sslKeyPath)
				_err := instance.ListenAndServeTLS("", "")
				//_err := instance.ListenAndServe()
				if _err != nil {
					_error().Err(_err).Msg(color.Style{color.LightRed}.Render("failed to listen http SSL server"))
				}
			}(instance)
		}
	}

	s.startTime.SetNow()

	// TODO: we should check every goroutine if it hasn't received any error!
	// TODO: but we will know that's listening?! after 1 second? or instantly!
	// Set server as Started!
	s.isStarted.True()
	// Set that start command has finished
	s.isStartCalled.False()

	s.LEvent("start", "OnStarted", nil)
	s.onStarted.Scan(func(k string, v interface{}) {
		v.(OnStarted)(s)
	})
	s.LEvent("finish", "OnStarted", nil)

	return nil
}

func (s *Server) IsStarting() bool {
	return s.isStartCalled.Get()
}

func (s *Server) IsStarted() bool {
	return s.isStarted.Get()
}
