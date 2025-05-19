package services

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/utils/random"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Session struct {
	ID         datatypes.UUID `gorm:"type:uuid;primaryKey"`
	SessionID  string         `json:"session_id" gorm:"index"`
	Username   string         `json:"username"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	GameMember GameMember     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"game_member"`
}

func CreateSession(db *gorm.DB, cfg *config.Config, username string) (*Session, error) {
	sessionObj := &Session{
		ID:        datatypes.NewUUIDv4(),
		SessionID: random.RandomString(32),
		Username:  username,
	}

	err := db.Create(sessionObj).Error
	if err != nil {
		return nil, err
	}

	return sessionObj, nil
}

func GetSessionBySessionID(db *gorm.DB, sessionID string) (*Session, error) {
	var sessionObj Session
	err := db.First(&sessionObj, "session_id = ?", sessionID).Error
	if err != nil {
		return nil, err
	}

	return &sessionObj, nil
}

func (session *Session) UpdateUsername(db *gorm.DB, username string) error {
	session.Username = username

	err := db.Save(session).Error
	if err != nil {
		return err
	}

	return nil
}

func (session *Session) Delete(db *gorm.DB) error {
	err := db.Delete(session).Error
	if err != nil {
		return err
	}

	return nil
}

func GetSessionByID(db *gorm.DB, id datatypes.UUID) (*Session, error) {
	var sessionObj Session
	err := db.First(&sessionObj, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &sessionObj, nil
}
