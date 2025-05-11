package database

import (
	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func New() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		utils.Logger.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&services.Game{}, &services.GameMember{}, &services.Session{})
	if err != nil {
		utils.Logger.Fatalf("failed to auto-migrate database: %v", err)
	}
	return db
}
