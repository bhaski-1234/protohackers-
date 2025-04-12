package main

import (
	"flag"
	"github.com/bhaski-1234/protohackers/smoketest/server"
)
import "github.com/bhaski-1234/protohackers/smoketest/config"

func setFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the application")
	flag.IntVar(&config.Port, "port", 9000, "Port for the application")
	// Parse the command line flags
	flag.Parse()
}

func main() {
	setFlags()
	server.RunServer()
}
