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
	ID          string       `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	GameMembers []GameMember `gorm:"foreignKey:GameID;constraint:OnDelete:CASCADE" json:"game_members"`
}

type GameMember struct {
	ID        datatypes.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	GameID    string         `gorm:"index" json:"game_id"`
	UserID    datatypes.UUID `gorm:"type:uuid;index" json:"user_id"`
	Host      bool           `json:"host"`
}

func CreateGame(db *gorm.DB, cfg *config.Config, hostID datatypes.UUID) (*Game, error) {
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
		ID: random.RandomString(4),
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

func (game *Game) Leave(db *gorm.DB, userID datatypes.UUID) error {
	var gameMemberObj GameMember
	err := db.Where("game_id = ? AND user_id = ?", game.ID, userID).First(&gameMemberObj).Error
	if err != nil {
		return err
	}

	err = db.Delete(&gameMemberObj).Error
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
