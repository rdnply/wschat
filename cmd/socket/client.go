package socket

import (
	"encoding/json"
	"fmt"
	"github.com/rdnply/wschat/internal/message"
	"github.com/rdnply/wschat/internal/user"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	readBufSize    = 1024
	writeBufSize   = 1024
)

type Client struct {
	hub *Hub
	//user *user.User
	login string
	conn  *websocket.Conn
	send  chan []byte
	//send chan *message.Message
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
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

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var u user.User
		if err := json.Unmarshal(msg, &u); err == nil && u.Login != "" {
			c.hub.addUser <- &u
			continue
		}

		var m message.Message
		if err := json.Unmarshal(msg, &m); err != nil {
			log.Printf("error: %v", err)
		}

		c.hub.broadcast <- &m

		//_, msg, err := c.conn.ReadMessage()
		//if err != nil {
		//	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		//		log.Printf("error: %v", err)
		//	}
		//	break
		//}
		//fmt.Println("msg:", msg)
		//msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		//os.Stdout.Write(msg)
		//fmt.Println("read")
		//c.hub.broadcast <- msg
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
	client := &Client{hub: hub, login: login, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	client.hub.addUser <- user.New(login)

	go client.readPump()
	go client.writePump()
}
