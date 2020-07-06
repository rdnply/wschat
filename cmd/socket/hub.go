package socket

import (
	"github.com/rdnply/wschat/internal/message"
	"github.com/rdnply/wschat/internal/user"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *message.Message
	addUser    chan *user.User
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *message.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		addUser:    make(chan *user.User),
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
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			toSend := message.ToSend()
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
