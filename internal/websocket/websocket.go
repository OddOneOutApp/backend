package websocket

import (
	"log"
	"net/http"

	"encoding/json"

	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/gorilla/websocket"
	"gorm.io/datatypes"
)

type Connection struct {
	Conn   *websocket.Conn
	Send   chan []byte
	UserID datatypes.UUID
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

func (c *Connection) ReadPump(db *gorm.DB, hub *Hub, gameID string) {
	defer func() {
		hub.removeConnection(gameID, c)
		c.Conn.Close()

		SendUserStatusMessage(gameID, c.UserID, false)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		utils.Logger.Debugf("Received message: %s", message)
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("failed to unmarshal message:", err)
			continue
		}

		switch msg.Type {
		case MessageTypeJoin:
			utils.Logger.Debugf("User %s joined game %s", msg.UserID, gameID)
		case MessageTypeStart:
			utils.Logger.Debugf("Game %s started", gameID)
			// TODO: relay start message to other users
		default:
			log.Println("unknown message type:", msg.Type)
		}

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
