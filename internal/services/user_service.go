package services

import (
	"github.com/OddOneOutApp/backend/internal/database/user"
	"github.com/OddOneOutApp/backend/internal/utils/random"
	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, username string) (*user.User, error) {
	sessionID := random.RandomString(32)

	userObj := &user.User{
		SessionID: sessionID,
		Username:  username,
	}

	err := db.Create(userObj).Error
	if err != nil {
		return nil, err
	}

	return userObj, nil
}

func GetUserBySessionID(db *gorm.DB, sessionID string) (*user.User, error) {
	var userObj user.User
	err := db.First(&userObj, "session_id = ?", sessionID).Error
	if err != nil {
		return nil, err
	}

	return &userObj, nil
}

func UpdateUser(db *gorm.DB, userObj *user.User) error {
	err := db.Save(userObj).Error
	if err != nil {
		return err
	}

	return nil
}
