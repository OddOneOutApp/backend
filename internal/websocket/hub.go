package websocket

import (
	"sync"

	"github.com/OddOneOutApp/backend/internal/utils"
	"gorm.io/datatypes"
)

type Hub struct {
	// Map of game IDs to a list of connections
	Games map[string]map[datatypes.UUID]*Connection
	mu    sync.RWMutex
}

var HubInstance *Hub

func Init() {
	HubInstance = NewHub()
	utils.Logger.Infoln("WebSocket Hub initialized")
}

func NewHub() *Hub {
	return &Hub{
		Games: make(map[string]map[datatypes.UUID]*Connection),
	}
}

// Add a connection to a game
func (h *Hub) AddConnection(gameID string, conn *Connection, userID datatypes.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Games[gameID]; !ok {
		h.Games[gameID] = make(map[datatypes.UUID]*Connection)
	}
	h.Games[gameID][userID] = conn
	conn.UserID = userID
}

// Remove a connection from a game
func (h *Hub) RemoveConnection(gameID string, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.Games[gameID]; ok {
		delete(connections, conn.UserID)
		if len(connections) == 0 {
			delete(h.Games, gameID)
		}
	}
}

// Broadcast a message to all connections in a game
func (h *Hub) Broadcast(gameID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if connections, ok := h.Games[gameID]; ok {
		for _, conn := range connections {
			conn.SendMessage(message)
			utils.Logger.Debugf("Broadcasting message to connection: %s", string(message))
		}
	}
}
