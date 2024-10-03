package api

import (
	"github.com/EggsyOnCode/vm-manager/api/handlers"
	"github.com/labstack/echo/v4"
)

type Server struct {
	echo *echo.Echo
}

func NewServer() *Server {
	e := echo.New()
	server := &Server{
		echo: e,
	}

	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	s.echo.POST("/create", handlers.HandleVMCreateReq)
}
func (s *Server) Start(addr string) {
	s.echo.Logger.Fatal(s.echo.Start(addr))
}
