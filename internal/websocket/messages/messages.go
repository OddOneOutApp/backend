package messages

import "gorm.io/datatypes"

type Message struct {
	Type    MessageType    `json:"type"`
	Content interface{}    `json:"content"`
	GameID  string         `json:"game_id"`
	UserID  datatypes.UUID `json:"user_id"`
}

type MessageType string

const (
	MessageTypeJoin       MessageType = "join"
	MessageTypeLeave      MessageType = "leave"
	MessageTypeUserStatus MessageType = "user_status"
	MessageTypeUpdateUser MessageType = "update_user"
	MessageTypeStart      MessageType = "start"
	MessageTypeQuestion   MessageType = "question"
	MessageTypeAnswers    MessageType = "answers"
)

func JoinMessage(gameID string, userID datatypes.UUID) Message {
	return Message{
		Type:   MessageTypeJoin,
		GameID: gameID,
		UserID: userID,
	}
}

func LeaveMessage(gameID string, userID datatypes.UUID) Message {
	return Message{
		Type:   MessageTypeLeave,
		GameID: gameID,
		UserID: userID,
	}
}

func StartMessage(gameID string) Message {
	return Message{
		Type:   MessageTypeStart,
		GameID: gameID,
	}
}

func QuestionMessage(gameID, question string) Message {
	return Message{
		Type:    MessageTypeQuestion,
		GameID:  gameID,
		Content: question,
	}
}

func AnswersMessage(gameID string, answers []string) Message {
	return Message{
		Type:    MessageTypeAnswers,
		GameID:  gameID,
		Content: answers,
	}
}

func UserStatusMessage(gameID string, userID datatypes.UUID, active bool) Message {
	return Message{
		Type:    MessageTypeUserStatus,
		GameID:  gameID,
		UserID:  userID,
		Content: active,
	}
}
