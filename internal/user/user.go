package user

import (
	"encoding/json"
	"github.com/rdnply/wschat/internal/message"
)

type User struct {
	Login    string                        `json:"login"`
	Messages map[string][]*message.Message `json:"messages"`
}

func New(login string) *User {
	return &User{
		Login:    login,
		Messages: make(map[string][]*message.Message),
	}
}

func ToSend(login string) []byte {
	u := New(login)
	b, err := json.Marshal(u)
	if err != nil {
		return nil
	}

	return b
}

type Storage interface {
	Find(login string) bool
	Add(login string)
	FindMessages(login string, companion string) []*message.Message
	AddMessage(login string, companion string, message *message.Message)
	GetLogins() []string
}
