package websocket

import (
	"net/http"
	"time"

	"encoding/json"

	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/gorilla/websocket"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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
			utils.Logger.Errorf("read error: %s", err)
			break
		}

		utils.Logger.Debugf("Received message: %s", message)
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			utils.Logger.Errorf("failed to unmarshal message: %s", err)
			continue
		}

		switch msg.Type {
		case MessageTypeStart:
			utils.Logger.Debugf("Game %s started", gameID)
			game, err := services.GetGameByID(db, gameID)
			if err != nil {
				utils.Logger.Errorf("failed to get game: %s", err)
				continue
			}

			seconds, ok := msg.Content.(int)
			if !ok {
				utils.Logger.Errorf("msg.Content is not a number")
				continue
			}
			gameEnd := time.Now().Add(time.Duration(seconds) * time.Second)
			game.SetAnswersEndTime(db, gameEnd)
			SendQuestionMessage(gameID, msg.UserID, game.RegularQuestion, game.SneakyQuestion, gameEnd)

		case MessageTypeAnswer:
			utils.Logger.Debugf("Received answer: %s", msg.Content)

			answer, ok := msg.Content.(string)
			if !ok {
				utils.Logger.Errorf("msg.Content is not a string")
				continue
			}

			game, err := services.GetGameByID(db, gameID)
			if err != nil {
				utils.Logger.Errorf("failed to get game: %s", err)
				continue
			}

			game.AddAnswer(db, c.UserID, answer)

		case MessageTypeVote:
			utils.Logger.Debugf("Received vote: %s", msg.Content)

			vote, ok := msg.Content.(datatypes.UUID)
			if !ok {
				utils.Logger.Errorf("msg.Content is not a uuid")
				continue
			}
			game, err := services.GetGameByID(db, gameID)
			if err != nil {
				utils.Logger.Errorf("failed to get game: %s", err)
				continue
			}
			game.Vote(db, c.UserID, vote)

		default:
			utils.Logger.Errorf("unknown message type: %s", msg.Type)
		}

	}
}

func (c *Connection) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			utils.Logger.Errorf("write error: %s", err)
			break
		}
	}
}

func (c *Connection) SendMessage(message []byte) {
	select {
	case c.Send <- message:
	default:
		utils.Logger.Errorf("send buffer full, dropping message")
	}
}
