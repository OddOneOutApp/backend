package game

import "time"

type Game struct {
	ID        string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	HostID    string `gorm:"uniqueIndex"`
	Members   []string
}
