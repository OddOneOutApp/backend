package user

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID        datatypes.UUID `gorm:"type:uuid;primaryKey"`
	SessionID string         `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.ID = datatypes.NewUUIDv4()
	return
}
