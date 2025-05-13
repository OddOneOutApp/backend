package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn *websocket.Conn
	Send chan []byte
}

func NewConnection(w http.ResponseWriter, r *http.Request) (*Connection, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Adjust for production
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &Connection{
		Conn: conn,
		Send: make(chan []byte, 256),
	}, nil
}

func (c *Connection) ReadPump(hub *Hub, gameID string) {
	defer func() {
		hub.RemoveConnection(gameID, c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// Handle incoming messages (e.g., broadcast to the game)
		hub.Broadcast(gameID, message)
	}
}

func (c *Connection) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write error:", err)
			break
		}
	}
}

func (c *Connection) SendMessage(message []byte) {
	select {
	case c.Send <- message:
	default:
		log.Println("send buffer full, dropping message")
	}
}
