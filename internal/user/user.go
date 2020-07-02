package user

import (
	"encoding/json"
)

type User struct {
	Login string `json:"login"`
}

func New(login string) *User {
	return &User{
		Login: login,
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
}
