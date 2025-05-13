package websocket

import (
	"sync"

	"gorm.io/datatypes"
)

type Hub struct {
	// Map of game IDs to a list of connections
	Games map[string]map[*Connection]datatypes.UUID
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Games: make(map[string]map[*Connection]datatypes.UUID),
	}
}

// Add a connection to a game
func (h *Hub) AddConnection(gameID string, conn *Connection, userID datatypes.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Games[gameID]; !ok {
		h.Games[gameID] = make(map[*Connection]datatypes.UUID)
	}
	h.Games[gameID][conn] = userID
}

// Remove a connection from a game
func (h *Hub) RemoveConnection(gameID string, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.Games[gameID]; ok {
		delete(connections, conn)
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
		for conn := range connections {
			conn.SendMessage(message)
		}
	}
}
