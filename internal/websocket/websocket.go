package websocket

import (
	"log"
	"net/http"

	"encoding/json"

	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/OddOneOutApp/backend/internal/websocket/messages"
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

func (c *Connection) ReadPump(hub *Hub, gameID string) {
	defer func() {
		hub.RemoveConnection(gameID, c)
		c.Conn.Close()

		userStatusMsg := messages.UserStatusMessage(gameID, c.UserID, false)
		userStatusMsgBytes, err := json.Marshal(userStatusMsg)
		if err != nil {
			log.Println("failed to marshal user status message:", err)
		} else {
			hub.Broadcast(gameID, userStatusMsgBytes)
		}
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		utils.Logger.Debugf("Received message: %s", message)
		var msg messages.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("failed to unmarshal message:", err)
			continue
		}

		switch msg.Type {
		case messages.MessageTypeJoin:
			utils.Logger.Debugf("User %s joined game %s", msg.UserID, gameID)
		case messages.MessageTypeStart:
			utils.Logger.Debugf("Game %s started", gameID)

			hub.Broadcast(gameID, message)
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
