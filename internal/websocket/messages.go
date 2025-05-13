package websocket

type Message struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
	GameID  string      `json:"game_id"`
	UserID  string      `json:"user_id"`
}

type MessageType string

const (
	MessageTypeJoin     MessageType = "join"
	MessageTypeLeave    MessageType = "leave"
	MessageTypeUpdate   MessageType = "update"
	MessageTypeStart    MessageType = "start"
	MessageTypeEnd      MessageType = "end"
	MessageTypeQuestion MessageType = "question"
	MessageTypeAnswers  MessageType = "answers"
)
