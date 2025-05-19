package websocket

import (
	"encoding/json"
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
func (h *Hub) AddConnection(gameID string, connection *Connection, userID datatypes.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Games[gameID]; !ok {
		h.Games[gameID] = make(map[datatypes.UUID]*Connection)
	}
	h.Games[gameID][userID] = connection
	connection.UserID = userID
}

// Remove a connection from a game
func (h *Hub) removeConnection(gameID string, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connections, ok := h.Games[gameID]; ok {
		delete(connections, conn.UserID)
		if len(connections) == 0 {
			delete(h.Games, gameID)
		}
	}
}

func (hub *Hub) broadcast(gameID string, message interface{}, exceptUserIDs ...datatypes.UUID) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		utils.Logger.Errorf("Failed to marshal message: %v", err)
		return
	}

	var exceptList []datatypes.UUID
	if len(exceptUserIDs) > 0 {
		exceptList = exceptUserIDs
	}

	if connections, ok := hub.Games[gameID]; ok {
		for userID, connection := range connections {
			if len(exceptList) == 0 || !contains(exceptList, userID) {
				connection.SendMessage(data)
				utils.Logger.Debugf("Broadcasting message to connection: %s", string(data))
			}
		}
	}
}

func (hub *Hub) sendToUser(gameID string, userID datatypes.UUID, message interface{}) {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		utils.Logger.Errorf("Failed to marshal message: %v", err)
		return
	}

	if connections, ok := hub.Games[gameID]; ok {
		if conn, ok := connections[userID]; ok {
			conn.SendMessage(data)
			utils.Logger.Debugf("Sending message to user %s in game %s: %s", userID, gameID, string(data))
		}
	}
}

// contains checks if a UUID is present in a slice of UUIDs
func contains(slice []datatypes.UUID, item datatypes.UUID) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
