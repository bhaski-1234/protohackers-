package server

import (
	"fmt"
	"github.com/bhaski-1234/protohackers/smoketest/config"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from connection: %v", err)
			return
		}

		_, err = conn.Write(buffer[:n])
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
			return
		}
	}
}

func RunServer() {
	lsnr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))

	if err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
	log.Printf("Listening on %s:%d", config.Host, config.Port)

	for {
		conn, err := lsnr.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
		}

		go handleConnection(conn)
	}

}
