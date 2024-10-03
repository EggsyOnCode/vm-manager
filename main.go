package main

import (
	"fmt"

	"github.com/EggsyOnCode/vm-manager/api"
)

func startServer() {
	s := api.NewServer()
	s.Start(":3000")
	fmt.Println("Starting server...")
}

func main() {
	startServer()

	select {}
}
