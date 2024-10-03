package api

import (
	"github.com/labstack/echo/v4"
)

type Server struct{}

func (s *Server) Start() {
	e := echo.New()

	// register the routes

	// Routes
	e.GET("/", hello)

	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}

func hello(c echo.Context) error {
	return c.String(200, "Hello, World!")
}
