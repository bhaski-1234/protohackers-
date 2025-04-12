package main

import (
	"flag"
	"github.com/bhaski-1234/protohackers/MeansToAnEnd/server"
	"log"
)
import "github.com/bhaski-1234/protohackers/MeansToAnEnd/config"

func getFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the application")
	flag.IntVar(&config.Port, "port", 9000, "Port for the application")
	flag.Parse()
}

func main() {
	getFlags()

	if err := server.RunServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
