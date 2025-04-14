package server

import (
	"bufio"
	"fmt"
	"github.com/bhaski-1234/protohackers/budgetChat/config"
	"log"
	"net"
	"strings"
	"sync"
)

type ChatServer struct {
	listener net.Listener
	wg       sync.WaitGroup
	chatRoom *ChatRoom
	mutex    sync.RWMutex
}

// NewChatServer creates a new chat server
func NewChatServer() (*ChatServer, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &ChatServer{
		listener: listener,
		chatRoom: NewChatRoom(),
		mutex:    sync.RWMutex{},
	}, nil
}

func (s *ChatServer) writeToConnection(conn net.Conn, data string) {
	_, err := conn.Write([]byte(data))
	if err != nil {
		log.Printf("Error writing to connection: %v", err)
	}
}

func (s *ChatServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	reader := bufio.NewReader(conn)
	//Ask for usernmae
	s.writeToConnection(conn, "Welcome to Budget Chat! What's your name?\n")

	// Read username
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading from connection: %v", err)
		return
	}
	username = strings.TrimSpace(username)

	if !isValidUserName(username) {
		s.writeToConnection(conn, "Invalid username. Please try again.\n")
		return
	}

	user := NewUser(username, conn)
	if !s.chatRoom.AddUserIfNotExists(user) {
		s.writeToConnection(conn, "Username already taken. Please try again.\n")
		return
	}

	//Get current users in the room
	userList := s.chatRoom.GetCurrentUsersList(username)
	s.writeToConnection(conn, fmt.Sprintf("* The room contains: %s\n", userList))
	s.broadCastMessageExceptClient(s.chatRoom, fmt.Sprintf("* %s has entered the room\n", username), username)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		message = strings.TrimSpace(message)
		s.broadCastMessageExceptClient(s.chatRoom, fmt.Sprintf("[%s] %s\n", username, message), username)
	}

	//user disconnected
	s.broadCastMessageExceptClient(s.chatRoom, fmt.Sprintf("* %s has left the room\n", username), username)
	s.chatRoom.RemoveUser(user)
}

func (s *ChatServer) broadCastMessageExceptClient(cr *ChatRoom, message string, excludeUser string) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	for username, user := range cr.users {
		if username != excludeUser {
			s.writeToConnection(user.conn, message)
		}
	}
}

func (s *ChatServer) startServer() {
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

func RunServer() error {
	chatServer, err := NewChatServer()
	if err != nil {
		return fmt.Errorf("failed to create chat server: %w", err)
	}

	chatServer.startServer()
	return nil
}
