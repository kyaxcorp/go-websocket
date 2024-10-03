package websocket

func (s *Server) SetReadBufferSize(bufferSize uint64) {
	s.readBufferSize.Set(bufferSize)
	// s.createWSUpgrader()
}

func (s *Server) SetWriteBufferSize(bufferSize uint64) {
	s.writeBufferSize.Set(bufferSize)
	// s.createWSUpgrader()
}
