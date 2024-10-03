package main

import (
	"fmt"
	"log"

	"github.com/EggsyOnCode/vm-manager/api"
	"github.com/joho/godotenv"
)

func startServer() {
	s := api.NewServer()
	s.Start(":3000")
	fmt.Println("Starting server...")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	startServer()

	select {}
}
