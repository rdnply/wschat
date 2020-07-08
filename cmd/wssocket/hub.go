package wssocket

import (
	"fmt"
	"github.com/rdnply/wschat/internal/message"
	"github.com/rdnply/wschat/internal/user"
)

type Hub struct {
	clients     map[*Client]bool
	users       map[string]*Client
	register    chan *Client
	unregister  chan *Client
	broadcast   chan *message.Message
	addUser     chan *user.User
	userStorage user.Storage
}

func NewHub(us user.Storage) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		users:       make(map[string]*Client),
		broadcast:   make(chan *message.Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		addUser:     make(chan *user.User),
		userStorage: us,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case newClient := <-h.register:
			//toSend := user.ToSend(newClient.login)
			//for client := range h.clients {
			//	select {
			//	case client.send <- toSend:
			//	default:
			//		delete(h.clients, client)
			//		close(client.send)
			//	}
			//}
			h.clients[newClient] = true
			h.users[newClient.login] = newClient
		case newUser := <-h.addUser:
			toSend := user.ToSend(newUser.Login)
			for client := range h.clients {
				if client.login != newUser.Login {
					select {
					case client.send <- toSend:
					default:
						delete(h.clients, client)
						close(client.send)
					}
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.users, client.login)
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			fmt.Println(message)
			toSend := message.ToSend()
			if message.To != "" {
				h.userStorage.AddMessage(message.From, message.To, message)
				h.userStorage.AddMessage(message.To, message.From, message)

				destination := h.users[message.To]
				destination.send <- toSend
			} else {
				for client := range h.clients {
					if message.From != client.login {
						select {
						case client.send <- toSend:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			}

		}
	}
}

//func (h *Hub) Broadcast(u *user.User) {
//	done := make(chan bool)
//
//	defer close(done)
//
//	go func() {
//		h.broadcast <- u
//		done <- true
//	}()
//
//	<-done
//}
