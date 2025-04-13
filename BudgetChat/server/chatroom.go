package server

import (
	"strings"
	"sync"
)

type ChatRoom struct {
	users map[string]*User
	mutex sync.RWMutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		users: make(map[string]*User),
	}
}

//Add user if does not exisis
func (cr *ChatRoom) AddUserIfNotExists(user *User) bool {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if _, exists := cr.users[user.name]; exists {
		return false
	}

	cr.users[user.name] = user
	return true
}

func (cr *ChatRoom) RemoveUser(user *User) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	delete(cr.users, user.name)
}

func (cr *ChatRoom) GetCurrentUsersList(excludeUser string) string {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	users := make([]string, 0, len(cr.users))
	for username := range cr.users {
		if username == excludeUser {
			users = append(users, username)
		}
	}
	if len(users) == 0 {
		return ""
	}
	return strings.Join(users, ",")
}
