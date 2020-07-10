package socket

import (
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
			// add new client in general storage of clients and
			// add in map by login for directed sending messages
			h.clients[newClient] = true
			h.users[newClient.login] = newClient
		case newUser := <-h.addUser:
			// send message about new added user for all clients
			clientInBytes := user.ConvertToBytes(newUser.Login)
			for client := range h.clients {
				// except to myself
				if client.login != newUser.Login {
					select {
					case client.send <- clientInBytes:
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
			messageInBytes := message.ConvertToBytes()
			// send to specific user
			if message.To != "" {
				// add message to storage for both users
				h.userStorage.AddMessage(message.From, message.To, message)
				h.userStorage.AddMessage(message.To, message.From, message)

				destination := h.users[message.To]
				destination.send <- messageInBytes
			} else {
				// send to all users expect ourselves
				for client := range h.clients {
					if message.From != client.login {
						select {
						case client.send <- messageInBytes:
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
