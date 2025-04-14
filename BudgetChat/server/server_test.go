package server_test

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func mustReadLine(t *testing.T, reader *bufio.Reader, label string) string {
	t.Helper()
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("[%s] failed to read line: %v", label, err)
	}
	return line
}

func connectAndReadWelcome(t *testing.T, label string) (net.Conn, *bufio.Reader) {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Fatalf("[%s] Failed to connect: %v", label, err)
	}
	reader := bufio.NewReader(conn)
	welcome := mustReadLine(t, reader, label)
	if welcome != "Welcome to Budget Chat! What's your name?\n" {
		t.Errorf("[%s] Unexpected welcome message: %s", label, welcome)
	}
	return conn, reader
}

func TestChatServerWithInvalidUserName(t *testing.T) {
	conn, reader := connectAndReadWelcome(t, "invalid-username")
	defer conn.Close()

	conn.Write([]byte("invalid@user\n"))
	msg := mustReadLine(t, reader, "invalid response")

	if !strings.Contains(msg, "Invalid username") {
		t.Errorf("Expected invalid username error, got: %s", msg)
	}

	_, err := reader.ReadString('\n')
	if err == nil {
		t.Errorf("Expected connection to be closed after invalid username")
	}
}

func TestDuplicateUsername(t *testing.T) {
	conn1, r1 := connectAndReadWelcome(t, "conn1")
	conn2, r2 := connectAndReadWelcome(t, "conn2")
	defer conn1.Close()
	defer conn2.Close()

	conn1.Write([]byte("alice\n"))
	time.Sleep(10 * time.Millisecond)
	conn2.Write([]byte("alice\n"))

	_ = mustReadLine(t, r1, "room contains alice")
	msg := mustReadLine(t, r2, "duplicate error")
	if !strings.Contains(msg, "Username already taken") {
		t.Errorf("Expected duplicate username error, got: %s", msg)
	}

	_, err := r2.ReadString('\n')
	if err == nil {
		t.Errorf("Expected EOF after duplicate username")
	}
}

func TestChatRoomWithMultipleUsers(t *testing.T) {
	conn1, r1 := connectAndReadWelcome(t, "user1")
	conn2, r2 := connectAndReadWelcome(t, "user2")
	conn3, r3 := connectAndReadWelcome(t, "user3")
	defer conn1.Close()
	defer conn2.Close()
	defer conn3.Close()

	conn1.Write([]byte("user1\n"))
	time.Sleep(10 * time.Millisecond)
	conn2.Write([]byte("user2\n"))
	time.Sleep(10 * time.Millisecond)
	conn3.Write([]byte("user3\n"))
	time.Sleep(10 * time.Millisecond)

	// Room messages
	roomMsg2 := mustReadLine(t, r2, "user2 sees room")
	if !strings.HasPrefix(roomMsg2, "* The room contains: user1") {
		t.Errorf("Expected user1 in room for user2, got: %s", roomMsg2)
	}

	roomMsg3 := mustReadLine(t, r3, "user3 sees room")
	if !strings.Contains(roomMsg3, "user1") || !strings.Contains(roomMsg3, "user2") {
		t.Errorf("Expected user1, user2 in room for user3, got: %s", roomMsg3)
	}

	// User3 joins - need to read join notifications on other connections
	joinMsg2 := mustReadLine(t, r2, "user2 sees user3 joined")
	if !strings.Contains(joinMsg2, "user3 has entered") {
		t.Errorf("Expected join notification, got: %s", joinMsg2)
	}

	_ = mustReadLine(t, r1, "user1 sees user1 joined")
	_ = mustReadLine(t, r1, "user1 sees user2 joined")
	_ = mustReadLine(t, r1, "user2 sees user3 joined")

	// user1 sends message
	conn1.Write([]byte("Hello everyone!\n"))
	broadcast2 := mustReadLine(t, r2, "broadcast to user2")
	broadcast3 := mustReadLine(t, r3, "broadcast to user3")
	expected := "[user1] Hello everyone!\n"
	if broadcast2 != expected || broadcast3 != expected {
		t.Errorf("Expected %q, got %q and %q", expected, broadcast2, broadcast3)
	}

	// user2 leaves
	conn2.Close()

	leftMsg1 := mustReadLine(t, r1, "user2 left")
	leftMsg3 := mustReadLine(t, r3, "user2 left")
	expectedLeft := "* user2 has left the room\n"
	if leftMsg1 != expectedLeft || leftMsg3 != expectedLeft {
		t.Errorf("Expected leave msg %q, got %q and %q", expectedLeft, leftMsg1, leftMsg3)
	}
}

func TestJoinNotificationOnlyToJoinedUsers(t *testing.T) {
	conn1, r1 := connectAndReadWelcome(t, "userA")
	conn2, r2 := connectAndReadWelcome(t, "userB")
	defer conn1.Close()
	defer conn2.Close()

	conn1.Write([]byte("userA\n"))
	time.Sleep(10 * time.Millisecond)

	_ = mustReadLine(t, r1, "room contains userA")

	conn2.Write([]byte("userB\n"))
	time.Sleep(10 * time.Millisecond)

	roomMsg := mustReadLine(t, r2, "userB room list")
	if !strings.Contains(roomMsg, "userA") {
		t.Errorf("Expected userA in room list for userB, got: %s", roomMsg)
	}

	notify := mustReadLine(t, r1, "userA sees userB joined")
	if !strings.Contains(notify, "userB has entered the room") {
		t.Errorf("Expected join notification to userA, got: %s", notify)
	}
}
