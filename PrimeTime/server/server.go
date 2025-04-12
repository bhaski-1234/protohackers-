package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bhaski-1234/protohackers/PrimeTime/config"
	"log"
	"net"
)

type request struct {
	Method string      `json:"method"`
	Number json.Number `json:"number"` // Using json.Number to handle both integers and floats
}

type response struct {
	Method  string `json:"method"`
	IsPrime bool   `json:"prime"`
}

func writeToConnection(conn net.Conn, data string) {
	_, err := conn.Write([]byte(data))
	if err != nil {
		fmt.Println("Error writing to connection:", err)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			break
		}

		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			fmt.Println("Error unmarshalling request:", err)
			break
		}

		resp, err := handlePrimeRequest(req)
		if err != nil {
			fmt.Println("Error handling request:", err)
			break
		}

		respData, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("Error marshalling response:", err)
			break
		}

		writeToConnection(conn, string(respData)+"\n")
	}
}

func handlePrimeRequest(req request) (response, error) {
	if req.Method != "isPrime" {
		return response{}, errors.New("unknown method")
	}
	if req.Number == "" {
		return response{}, errors.New("missing number field")
	}

	return response{
		Method:  "isPrime",
		IsPrime: isPrime(req.Number),
	}, nil
}

func isPrime(num json.Number) bool {
	n, err := num.Float64()
	if err != nil {
		return false
	}

	if n != float64(int(n)) {
		return false
	}

	// Convert to int for prime check
	intN := int(n)
	if intN < 2 {
		return false
	}

	for i := 2; i*i <= intN; i++ {
		if intN%i == 0 {
			return false
		}
	}
	return true
}

func RunServer() {
	lsnr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer lsnr.Close()

	log.Printf("Listening on %s:%d", config.Host, config.Port)

	for {
		conn, err := lsnr.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
