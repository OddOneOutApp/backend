package game

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Game struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	HostID    datatypes.UUID `gorm:"type:uuid;uniqueIndex"`
}

type GameMember struct {
	ID        datatypes.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	GameID    string         `gorm:"index"`
	UserID    datatypes.UUID `gorm:"type:uuid;index"`
}

func (gameMember *GameMember) BeforeCreate(tx *gorm.DB) (err error) {
	gameMember.ID = datatypes.NewUUIDv4()
	return
}
