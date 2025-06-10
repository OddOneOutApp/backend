package websocket

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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
	MessageTypeGameDelete MessageType = "game_delete"
	MessageTypeUserStatus MessageType = "user_status"
	MessageTypeInit       MessageType = "init"
	MessageTypeUpdateUser MessageType = "update_user"
	MessageTypeStart      MessageType = "start" // sent by client
	MessageTypeQuestion   MessageType = "question"
	MessageTypeAnswer     MessageType = "answer" // sent by client
	MessageTypeAnswers    MessageType = "answers"
	MessageTypeVote       MessageType = "vote" // sent by client
	MessageTypeVoteResult MessageType = "vote_result"
)

func SendJoinMessage(gameID string, userID datatypes.UUID, username string) {
	HubInstance.broadcast(gameID, Message{
		Type:    MessageTypeJoin,
		GameID:  gameID,
		UserID:  userID,
		Content: username,
	}, userID)
}

func SendUserLeaveMessage(gameID string, userID datatypes.UUID) {
	HubInstance.broadcast(gameID, Message{
		Type:   MessageTypeLeave,
		GameID: gameID,
		UserID: userID,
	})
}

func SendGameDeleteMessage(gameID string) {
	HubInstance.broadcast(gameID, Message{
		Type:   MessageTypeGameDelete,
		GameID: gameID,
	})
}

func SendUserStatusMessage(gameID string, userID datatypes.UUID, active bool) {
	HubInstance.broadcast(gameID, Message{
		Type:    MessageTypeUserStatus,
		GameID:  gameID,
		UserID:  userID,
		Content: active,
	}, userID)
}

func SendInitMessage(gameID string, userID datatypes.UUID, db *gorm.DB) {
	game, err := services.GetGameByID(db, gameID)
	if err != nil {
		utils.Logger.Errorf("Error fetching game by ID: %v", err)
		return
	}

	var users []UserInfo
	members, err := game.GetMembers(db)
	if err != nil {
		utils.Logger.Errorf("Error fetching game members: %v", err)
		return
	}
	for _, member := range members {
		userSession, err := services.GetSessionByID(db, member.UserID)
		if err != nil {
			utils.Logger.Errorf("Error fetching session by ID: %v", err)
			continue
		}

		connection := HubInstance.Games[gameID][member.UserID]
		if connection == nil {
			utils.Logger.Debugf("Connection not found for user ID: %s", member.UserID)
		}
		users = append(users, UserInfo{
			ID:     member.UserID,
			Name:   userSession.Username,
			Active: connection != nil,
		})
		utils.Logger.Debugf("User %s is in game %s", member.UserID, gameID)
	}

	question, err := game.GetQuestionForUser(db, userID)
	if err != nil {
		utils.Logger.Errorf("Error fetching question for user %s in game %s: %v", userID, gameID, err)
		question = ""
	}

	actualQuestion := game.RegularQuestion

	answers, err := game.GetAnswers(db)
	if err != nil {
		utils.Logger.Errorf("Error fetching answers for game %s: %v", gameID, err)
		answers = []services.Answer{}
	}

	if !(game.State == services.GameStateVoting || game.State == services.GameStateFinished) {
		actualQuestion = ""
		answers = []services.Answer{}
	}

	HubInstance.sendToUser(gameID, userID, Message{
		Type:   MessageTypeInit,
		GameID: gameID,
		UserID: userID,
		Content: map[string]interface{}{
			"users":            users,
			"game_state":       game.State,
			"answers_end_time": game.AnswersEndTime.Unix(),
			"voting_end_time":  game.VotingEndTime.Unix(),
			"question":         question,
			"actual_question":  actualQuestion,
			"answers":          answers,
		},
	})
}

func SendUpdateUserMessage(gameID string, userID datatypes.UUID, username string) {
	HubInstance.broadcast(gameID, Message{
		Type:    MessageTypeUpdateUser,
		GameID:  gameID,
		UserID:  userID,
		Content: username,
	}, userID)
}

func SendQuestionMessage(gameID string, impostorUUIDs []datatypes.UUID, question string, impostorQuestion string, gameEnd time.Time) {
	HubInstance.broadcast(gameID, Message{
		Type:   MessageTypeQuestion,
		GameID: gameID,
		Content: map[string]interface{}{
			"question":         question,
			"answers_end_time": gameEnd.Unix(),
		},
	}, impostorUUIDs...)

	for _, impostorUUID := range impostorUUIDs {
		utils.Logger.Debugf("Sending impostor question to user %s in game %s", impostorUUID, gameID)
		HubInstance.sendToUser(gameID, impostorUUID, Message{
			Type:   MessageTypeQuestion,
			GameID: gameID,
			Content: map[string]interface{}{
				"question":         impostorQuestion,
				"answers_end_time": gameEnd.Unix(),
			},
		})
	}

}

func SendAnswersMessage(gameID string, answers []services.Answer, actualQuestion string, votingEnd time.Time) {
	HubInstance.broadcast(gameID, Message{
		Type:   MessageTypeAnswers,
		GameID: gameID,
		Content: map[string]interface{}{
			"answers":         answers,
			"actual_question": actualQuestion,
			"voting_end_time": votingEnd.Unix(),
		},
	})
}

func SendVoteResultMessage(gameID string, votes map[datatypes.UUID]datatypes.UUID) {
	stringVotes := make(map[string]string, len(votes))
	for k, v := range votes {
		stringVotes[k.String()] = v.String()
	}
	HubInstance.broadcast(gameID, Message{
		Type:    MessageTypeVoteResult,
		GameID:  gameID,
		Content: stringVotes,
	})
}

type UserInfo struct {
	ID     datatypes.UUID `json:"id"`
	Name   string         `json:"name"`
	Active bool           `json:"active"`
}
