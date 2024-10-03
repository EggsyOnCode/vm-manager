package main

import (
	"fmt"

	"github.com/EggsyOnCode/vm-manager/api"
)

func startServer() {
	s := new(api.Server)
	fmt.Print("Starting server...")
	go s.Start()
}

func main() {
	startServer()

	select {}
}
