package server

import "net"

type User struct {
	name string
	conn net.Conn
}

func NewUser(name string, conn net.Conn) *User {
	return &User{
		name: name,
		conn: conn,
	}
}

// Helper function
func isValidUserName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9')) {
			return false
		}
	}
	return true
}
