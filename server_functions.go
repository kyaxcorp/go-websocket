package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/kyaxcorp/go-helper/file"
	"github.com/kyaxcorp/go-helper/filesystem"
)

func (s *Server) GetNrOfClients() uint {
	return s.c.GetNrOfClients()
}

func (s *Server) GetWSServer() *gin.Engine {
	return s.WSServer
}

func (s *Server) GetClients() map[*Client]bool {
	return s.c.GetClients()
}

func (s *Server) GetClientsOrderedByConnectionID() map[int64]*Client {
	return s.c.GetClientsOrderedByConnectionID()
}

// GetClientsLogPath -> returns the path where the logs for clients are stored
func (s *Server) GetClientsLogPath() string {
	// Creating clients path
	return file.FilterPath(s.LoggerDirPath + filesystem.DirSeparator() + "clients" + filesystem.DirSeparator())
}
