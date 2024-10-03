package websocket

// THe user can start it in a goroutine
func (s *Server) EnableSSL(keyPath string, certPath string) error {

	// TODO: check if params are not empty!
	// TODO: check if the files exist!

	s.sslKeyPath = keyPath
	s.sslCertPath = certPath
	s.enableSSL = true
	return nil
}

func (s *Server) EnableUnsecure() *Server {
	s.enableUnsecure = true
	return s
}

func (s *Server) DisableUnsecure() *Server {
	s.enableUnsecure = false
	return s
}
