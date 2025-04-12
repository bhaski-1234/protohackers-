package server_test

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// Define structure for expected response
type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

// Helper function to connect and exchange a single request-response
func sendRequest(t *testing.T, req string) (string, error) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(req + "\n"))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	respLine, err := reader.ReadString('\n')
	return strings.TrimSpace(respLine), err
}

func TestValidPrimeRequest(t *testing.T) {
	req := `{"method":"isPrime","number":7}`
	resp, err := sendRequest(t, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var result response
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		t.Fatalf("Invalid JSON in response: %v", err)
	}

	if result.Method != "isPrime" || result.Prime != true {
		t.Errorf("Unexpected response: %s", resp)
	}
}

func TestValidNonPrimeRequest(t *testing.T) {
	req := `{"method":"isPrime","number":10}`
	resp, err := sendRequest(t, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var result response
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		t.Fatalf("Invalid JSON in response: %v", err)
	}

	if result.Method != "isPrime" || result.Prime != false {
		t.Errorf("Unexpected response: %s", resp)
	}
}

func TestNonIntegerRequest(t *testing.T) {
	req := `{"method":"isPrime","number":7.5}`
	resp, err := sendRequest(t, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var result response
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		t.Fatalf("Invalid JSON in response: %v", err)
	}

	if result.Method != "isPrime" || result.Prime != false {
		t.Errorf("Unexpected response: %s", resp)
	}
}

func TestMalformedJSON(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Connection error: %v", err)
	}
	defer conn.Close()

	req := `{bad json}`
	_, err = conn.Write([]byte(req + "\n"))
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	reader := bufio.NewReader(conn)
	_, err = reader.ReadString('\n')
	if err == nil {
		t.Errorf("Expected connection to be closed after malformed request")
	}
}

func TestMissingFields(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Connection error: %v", err)
	}
	defer conn.Close()
	req := `{"method":"isPrime"}` // missing "number"

	_, err = conn.Write([]byte(req + "\n")) // missing "number"
	if err != nil {
		t.Fatalf("Failed to write to connection: %v", err)
	}

	reader := bufio.NewReader(conn)
	_, err = reader.ReadString('\n')
	if err == nil {
		t.Errorf("Expected connection to be closed after malformed request")
	}
}

func TestConcurrentClients(t *testing.T) {
	var wg sync.WaitGroup
	numClients := 5
	errorsChan := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := `{"method":"isPrime","number":17}`
			resp, err := sendRequest(t, req)
			if err != nil {
				errorsChan <- err
				return
			}
			var result response
			err = json.Unmarshal([]byte(resp), &result)
			if err != nil || result.Method != "isPrime" || result.Prime != true {
				errorsChan <- err
				return
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case err := <-errorsChan:
		t.Errorf("Client error: %v", err)
	case <-time.After(5 * time.Second):
		t.Errorf("Concurrent test timed out")
	}
}
