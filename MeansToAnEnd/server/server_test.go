package server_test

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"
)

func sendMessage(t *testing.T, conn net.Conn, msg []byte) {
	t.Helper()
	_, err := conn.Write(msg)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
}

func readResponse(t *testing.T, conn net.Conn) int32 {
	t.Helper()
	buf := make([]byte, 4)
	_, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	var result int32
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, &result)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	return result
}

func buildInsert(timestamp, price int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte('I')
	binary.Write(buf, binary.BigEndian, timestamp)
	binary.Write(buf, binary.BigEndian, price)
	return buf.Bytes()
}

func buildQuery(minTime, maxTime int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte('Q')
	binary.Write(buf, binary.BigEndian, minTime)
	binary.Write(buf, binary.BigEndian, maxTime)
	return buf.Bytes()
}

func TestInsertAndQuerySingle(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Insert: T=1000, Price=500
	sendMessage(t, conn, buildInsert(1000, 500))

	// Query: T=[500, 1500]
	sendMessage(t, conn, buildQuery(500, 1500))

	// Expect: 500
	result := readResponse(t, conn)
	if result != 500 {
		t.Errorf("Expected 500, got %d", result)
	}
}

func TestMultipleInsertsAndQuery(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Insert multiple
	sendMessage(t, conn, buildInsert(1000, 100))
	sendMessage(t, conn, buildInsert(2000, 300))
	sendMessage(t, conn, buildInsert(3000, 500))

	// Query range: [1000, 3000] => mean of 100, 300, 500 = 300
	sendMessage(t, conn, buildQuery(1000, 3000))

	result := readResponse(t, conn)
	if result != 300 {
		t.Errorf("Expected mean 300, got %d", result)
	}
}

func TestQueryEmptyRangeReturnsZero(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	sendMessage(t, conn, buildQuery(5000, 6000))
	result := readResponse(t, conn)

	if result != 0 {
		t.Errorf("Expected 0 for empty query, got %d", result)
	}
}

func TestQueryWithMinTimeAfterMaxTime(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	sendMessage(t, conn, buildQuery(3000, 1000))
	result := readResponse(t, conn)

	if result != 0 {
		t.Errorf("Expected 0 for invalid time range, got %d", result)
	}
}

func TestSessionIsolation(t *testing.T) {
	conn1, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect conn1: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("Failed to connect conn2: %v", err)
	}
	defer conn2.Close()

	// Insert into conn1 only
	sendMessage(t, conn1, buildInsert(1111, 999))

	// conn2 queries that range
	sendMessage(t, conn2, buildQuery(1000, 1200))

	// conn2 should get 0 since it sees no data
	result := readResponse(t, conn2)
	if result != 0 {
		t.Errorf("Expected 0 for session isolation, got %d", result)
	}
}
