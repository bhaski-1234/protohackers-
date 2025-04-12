package main

import (
	"flag"
	"github.com/bhaski-1234/protohackers/PrimeTime/config"
	"github.com/bhaski-1234/protohackers/PrimeTime/server"
)

func getFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the application")
	flag.IntVar(&config.Port, "port", 9000, "Port for the application")
	flag.Parse()
}

func main() {
	getFlags()
	server.RunServer()
}
