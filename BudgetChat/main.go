package main

import (
	"flag"
	"github.com/bhaski-1234/protohackers/budgetChat/config"
	"log"
)

func getFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the application")
	flag.IntVar(&config.Port, "port", 8080, "Port for the application")
	flag.Parse()
}

func main() {
	getFlags()
	if err := server.RunServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	// server.RunServer()
	// server := server.NewServer()
	// defer server.Stop()
	// go server.Start()
	// select {} // Block forever
}
