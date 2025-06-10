package websocket

import (
	"net/http"
	"strconv"
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

			var seconds int
			switch v := msg.Content.(type) {
			case float64:
				seconds = int(v)
			case int:
				seconds = v
			case string:
				parsed, err := strconv.Atoi(v)
				if err != nil {
					utils.Logger.Errorf("failed to parse seconds from string: %s", err)
					continue
				}
				seconds = parsed
			default:
				utils.Logger.Errorf("msg.Content is not a valid number type")
				continue
			}

			impostors, err := game.SelectImpostors(db, 1)
			if err != nil {
				utils.Logger.Errorf("failed to get impostors: %s", err)
				continue
			}

			impostorUUIDs := make([]datatypes.UUID, 0, len(impostors))
			for _, imp := range impostors {
				impostorUUIDs = append(impostorUUIDs, imp.UserID)
			}

			regularQuestion, sneakyQuestion, err := services.SelectQuestionFromCategory(game.Category)
			if err != nil {
				utils.Logger.Errorf("failed to select question: %s", err)
				continue
			}
			game.RegularQuestion = regularQuestion
			game.SneakyQuestion = sneakyQuestion
			if err := db.Save(&game).Error; err != nil {
				utils.Logger.Errorf("failed to save game: %s", err)
				continue
			}

			gameEnd := time.Now().Add(time.Duration(seconds) * time.Second)
			game.SetAnswersEndTimeAndGameState(db, gameEnd)
			SendQuestionMessage(gameID, impostorUUIDs, game.RegularQuestion, game.SneakyQuestion, gameEnd)

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

			// Expecting msg.Content to be a []interface{} representing the UUID bytes
			contentSlice, ok := msg.Content.([]interface{})
			if !ok || len(contentSlice) != 16 {
				utils.Logger.Errorf("msg.Content is not a valid uuid slice")
				continue
			}
			var uuidBytes [16]byte
			for i, v := range contentSlice {
				f, ok := v.(float64)
				if !ok {
					utils.Logger.Errorf("msg.Content element is not a float64")
					continue
				}
				uuidBytes[i] = byte(f)
			}
			vote := datatypes.UUID(uuidBytes)
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
