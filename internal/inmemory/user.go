package inmemory

import (
	"fmt"
	"github.com/rdnply/wschat/internal/message"
	"github.com/rdnply/wschat/internal/user"
	"sync"
)

var _ user.Storage = &UserStorage{}

type UserStorage struct {
	sync.Mutex
	st map[string]*user.User
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		st: make(map[string]*user.User),
	}
}

func (us *UserStorage) Find(login string) bool {
	us.Lock()
	defer us.Unlock()
	_, ok := us.st[login]
	if !ok {
		return false
	}

	return true
}

func (us *UserStorage) Add(login string) {
	us.Lock()
	us.st[login] = user.New(login)
	us.Unlock()
}

func (us *UserStorage) FindMessages(login string, companion string) []*message.Message {
	us.Lock()
	defer us.Unlock()
	u, ok := us.st[login]
	if !ok {
		return nil
	}

	return u.Messages[companion]
}

func (us *UserStorage) AddMessage(login string, companion string, message *message.Message) {
	us.Lock()
	messages := us.st[login].Messages[companion]
	messages = append(messages, message)
	us.st[login].Messages[companion] = messages
	us.Unlock()
	fmt.Println("messages", messages)
	fmt.Println("map", us.st[login].Messages[companion])
}


func (us *UserStorage) GetLogins() []string {
	keys := make([]string, 0, len(us.st))

	us.Lock()
	for k, _ := range us.st {
		keys = append(keys, k)
	}
	us.Unlock()

	return keys
}