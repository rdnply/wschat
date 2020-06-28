package socket

import (
	"fmt"
	"github.com/rdnply/wschat/internal/user"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingPeriod   = (pongWait * 9) / 10
	readBufSize  = 1024
	writeBufSize = 1024
)

type Client struct {
	hub  *Hub
	//user *user.User
	login string
	conn *websocket.Conn
	send chan *user.User
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufSize,
		WriteBufferSize: writeBufSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "can't open websocket connection", http.StatusBadRequest)
	}

	login := r.URL.Query().Get("login")
	if login == "" {
		http.Error(w, "can't find login in url params", http.StatusBadRequest)
	}

	fmt.Println(login)
	client := &Client{hub: hub, login: login, conn: conn, send: make(chan *user.User)}
	client.hub.register <- client

	go client.writePump()
}
