package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/bhaski-1234/protohackers/MeansToAnEnd/config"
)

const (
	InsertOperation byte = 'I'
	QueryOperation  byte = 'Q'
	MessageSize     int  = 9
)

// PriceServer handles the means-to-an-end protocol
type PriceServer struct {
	listener net.Listener
	wg       sync.WaitGroup
}

// NewServer creates a new price server
func NewServer() (*PriceServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &PriceServer{
		listener: listener,
	}, nil
}

// Start begins accepting connections
func (s *PriceServer) Start() {
	log.Printf("Server listening on %s:%d", config.Host, config.Port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// Stop gracefully shuts down the server
func (s *PriceServer) Stop() error {
	err := s.listener.Close()
	s.wg.Wait() // Wait for all connections to finish
	return err
}

func (s *PriceServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	log.Printf("New connection from %s", conn.RemoteAddr())

	// Map of timestamp to price for this connection
	priceMap := make(map[uint32]uint32)
	buf := make([]byte, MessageSize)

	for {
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by client: %s", conn.RemoteAddr())
			} else {
				log.Printf("Error reading from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		if err := s.processMessage(conn, buf, priceMap); err != nil {
			log.Printf("Error processing message from %s: %v", conn.RemoteAddr(), err)
			return
		}
	}
}

func (s *PriceServer) processMessage(conn net.Conn, data []byte, priceMap map[uint32]uint32) error {
	switch data[0] {
	case InsertOperation:
		timestamp := binary.BigEndian.Uint32(data[1:5])
		price := binary.BigEndian.Uint32(data[5:9])
		priceMap[timestamp] = price
		return nil

	case QueryOperation:
		minTime := binary.BigEndian.Uint32(data[1:5])
		maxTime := binary.BigEndian.Uint32(data[5:9])

		if maxTime < minTime {
			// Swap for consistency with problem definition
			minTime, maxTime = maxTime, minTime
		}

		avgPrice := calculateAverage(priceMap, minTime, maxTime)

		responseData := make([]byte, 4)
		binary.BigEndian.PutUint32(responseData, avgPrice)

		if _, err := conn.Write(responseData); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown operation: %c", data[0])
	}
}

func calculateAverage(priceMap map[uint32]uint32, minTime, maxTime uint32) uint32 {
	var total uint64 // Use uint64 to avoid overflow
	var count uint32

	for timestamp, price := range priceMap {
		if timestamp >= minTime && timestamp <= maxTime {
			total += uint64(price)
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return uint32(total / uint64(count))
}

// RunServer starts the server and blocks until complete
func RunServer() error {
	server, err := NewServer()
	if err != nil {
		return err
	}

	server.Start()
	return nil
}
