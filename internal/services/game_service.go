package services

import (
	"fmt"
	"time"

	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/utils/random"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Game struct {
	ID              string       `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	Category        string       `gorm:"index" json:"category"`
	RegularQuestion string       `json:"regular_question"`
	SneakyQuestion  string       `json:"sneaky_question"`
	AnswersEndTime  time.Time    `json:"answers_end_time"`
	VotingEndTime   time.Time    `json:"voting_end_time"`
	State           GameState    `gorm:"default:'lobby'" json:"state"`
	GameMembers     []GameMember `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"game_members"`
	Answers         []Answer     `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"answers"`
}

type GameMember struct {
	ID        datatypes.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	GameID    string         `gorm:"index" json:"game_id"`
	UserID    datatypes.UUID `gorm:"type:uuid;index" json:"user_id"`
	Host      bool           `json:"host"`
	Impostor  bool           `json:"impostor"`
	Vote      datatypes.UUID `gorm:"type:uuid;index" json:"vote"`
}

type Answer struct {
	ID        datatypes.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	GameID    string         `gorm:"index" json:"game_id"`
	UserID    datatypes.UUID `gorm:"type:uuid;index" json:"user_id"`
	Answer    string         `json:"answer"`
}

type GameState string

const (
	GameStateLobby     GameState = "lobby"
	GameStateAnswering GameState = "answering"
	GameStateVoting    GameState = "voting"
	GameStateFinished  GameState = "finished"
)

func CreateGame(db *gorm.DB, cfg *config.Config, hostID datatypes.UUID, category string) (*Game, error) {
	// Check if user is already in a game
	var existingMember GameMember
	result := db.Where("user_id = ?", hostID).First(&existingMember)
	if result.Error == nil {
		// User already in a game, return an error
		return nil, fmt.Errorf("user is already in a game")
	} else if result.Error != gorm.ErrRecordNotFound {
		// Some other error occurred
		return nil, result.Error
	}

	// User not in a game, create new game
	gameObj := &Game{
		ID:             random.RandomString(4),
		Category:       category,
		State:          GameStateLobby,
		AnswersEndTime: time.Unix(0, 0).UTC(),
		VotingEndTime:  time.Unix(0, 0).UTC(),
	}

	err := db.Create(gameObj).Error
	if err != nil {
		return nil, err
	}

	gameMemberObj := &GameMember{
		ID:     datatypes.NewUUIDv4(),
		GameID: gameObj.ID,
		UserID: hostID,
		Host:   true,
	}
	err = db.Create(gameMemberObj).Error
	if err != nil {
		return nil, err
	}

	return gameObj, nil
}

func GetGameByID(db *gorm.DB, gameID string) (*Game, error) {
	var gameObj Game
	err := db.First(&gameObj, "id = ?", gameID).Error
	if err != nil {
		return nil, err
	}

	return &gameObj, nil
}

func (game *Game) Join(db *gorm.DB, userID datatypes.UUID) (*GameMember, error) {
	// Check if user is already in the game
	var existingMember GameMember
	result := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&existingMember)
	if result.Error == nil {
		// User already in the game, return an error
		return nil, fmt.Errorf("user is already in the game")
	} else if result.Error != gorm.ErrRecordNotFound {
		// Some other error occurred
		return nil, result.Error
	}

	// User not in the game, create new member
	gameMemberObj := &GameMember{
		ID:     datatypes.NewUUIDv4(),
		GameID: game.ID,
		UserID: userID,
		Host:   false,
	}

	err := db.Create(gameMemberObj).Error
	if err != nil {
		return nil, err
	}

	return gameMemberObj, nil
}

func (game *Game) SetAnswersEndTimeAndGameState(db *gorm.DB, endTime time.Time) error {
	game.AnswersEndTime = endTime
	game.State = GameStateAnswering
	err := db.Save(game).Error
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) SetVotingEndTimeAndGameState(db *gorm.DB, endTime time.Time) error {
	game.VotingEndTime = endTime
	game.State = GameStateVoting
	err := db.Save(game).Error
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) Leave(db *gorm.DB, userID datatypes.UUID) error {
	err := db.Where("game_id = ? AND user_id = ?", game.ID, userID).Delete(&GameMember{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) Delete(db *gorm.DB) error {
	var gameMembers []GameMember
	err := db.Where("game_id = ?", game.ID).Find(&gameMembers).Error
	if err != nil {
		return err
	}

	for _, member := range gameMembers {
		err = db.Delete(&member).Error
		if err != nil {
			return err
		}
	}

	err = db.Delete(game).Error
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) GetMembers(db *gorm.DB) ([]GameMember, error) {
	var gameMembers []GameMember
	err := db.Where("game_id = ?", game.ID).Find(&gameMembers).Error
	if err != nil {
		return nil, err
	}

	return gameMembers, nil
}

func (game *Game) IsHost(db *gorm.DB, userID datatypes.UUID) (bool, error) {
	var gameMemberObj GameMember
	err := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&gameMemberObj).Error
	if err != nil {
		return false, err
	}

	return gameMemberObj.Host, nil
}

func (game *Game) IsUserInGame(db *gorm.DB, userID datatypes.UUID) (bool, error) {
	var gameMemberObj GameMember
	err := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&gameMemberObj).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // User not found in game
		}
		return false, err // Some other error occurred
	}

	return true, nil // User found in game
}

func (game *Game) AddAnswer(db *gorm.DB, userID datatypes.UUID, answer string) (*Answer, error) {
	// Check if user is already in the game
	var existingMember GameMember
	result := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&existingMember)
	if result.Error != nil {
		return nil, fmt.Errorf("user is not in the game")
	}

	// User is in the game, create new answer
	answerObj := &Answer{
		ID:     datatypes.NewUUIDv4(),
		GameID: game.ID,
		UserID: userID,
		Answer: answer,
	}

	err := db.Create(answerObj).Error
	if err != nil {
		return nil, err
	}

	return answerObj, nil
}

func (game *Game) GetAnswers(db *gorm.DB) ([]Answer, error) {
	var answers []Answer
	err := db.Where("game_id = ?", game.ID).Find(&answers).Error
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (game *Game) Vote(db *gorm.DB, userID datatypes.UUID, answerID datatypes.UUID) error {
	// Check if user is already in the game
	var existingMember GameMember
	result := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&existingMember)
	if result.Error != nil {
		return fmt.Errorf("user is not in the game")
	}

	// User is in the game, update vote count for the answer
	var answerObj Answer
	err := db.Where("user_id = ? AND game_id = ?", answerID, game.ID).First(&answerObj).Error
	if err != nil {
		return err
	}

	existingMember.Vote = answerID
	err = db.Save(&existingMember).Error
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) GetVoteResults(db *gorm.DB) (map[datatypes.UUID]datatypes.UUID, error) {
	var gameMembers []GameMember
	err := db.Where("game_id = ?", game.ID).Find(&gameMembers).Error
	if err != nil {
		return nil, err
	}

	votesMap := make(map[datatypes.UUID]datatypes.UUID)
	for _, member := range gameMembers {
		if member.Vote != (datatypes.UUID{}) {
			votesMap[member.UserID] = member.Vote
		}
	}

	return votesMap, nil
}

func (game *Game) GetImpostors(db *gorm.DB) ([]GameMember, error) {
	var impostors []GameMember
	err := db.Where("game_id = ? AND impostor = ?", game.ID, true).Find(&impostors).Error
	if err != nil {
		return nil, err
	}

	return impostors, nil
}

func (game *Game) SelectImpostors(db *gorm.DB, count int) ([]GameMember, error) {
	var gameMembers []GameMember
	err := db.Where("game_id = ?", game.ID).Find(&gameMembers).Error
	if err != nil {
		return nil, err
	}

	if len(gameMembers) < count {
		return nil, fmt.Errorf("not enough players to select impostors")
	}

	selectedImpostors := random.RandomSelect(gameMembers, count)

	for i := range selectedImpostors {
		selectedImpostors[i].Impostor = true
		err = db.Save(&selectedImpostors[i]).Error
		if err != nil {
			return nil, err
		}
	}

	return selectedImpostors, nil
}

func (game *Game) IsGameMemberImpostor(db *gorm.DB, userID datatypes.UUID) (bool, error) {
	var gameMember GameMember
	err := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&gameMember).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // User not found in game
		}
		return false, err // Some other error occurred
	}

	return gameMember.Impostor, nil
}

func (game *Game) GetQuestionForUser(db *gorm.DB, userID datatypes.UUID) (string, error) {
	isImpostor, err := game.IsGameMemberImpostor(db, userID)
	if err != nil {
		return "", err
	}

	if isImpostor {
		return game.SneakyQuestion, nil
	}
	return game.RegularQuestion, nil
}
