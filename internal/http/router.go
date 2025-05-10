package http

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/database/game"
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/OddOneOutApp/backend/internal/utils/random"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Initialize(db *gorm.DB, cfg *config.Config) {
	router := gin.Default()

	router.Use(ginzap.Ginzap(utils.RawLogger, time.RFC3339, true))

	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/api/games", func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.JSON(400, gin.H{
				"error": "Session ID not found",
			})
			return
		}

		gameID := random.RandomString(4)

		err = db.Create(&game.Game{
			ID:     gameID,
			HostID: sessionID,
		}).Error

		if err != nil {
			utils.Logger.Errorf("Error creating game: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}

		utils.Logger.Infof("Game created with ID: %s for session ID: %s", gameID, sessionID)

		c.JSON(200, gin.H{
			"message": "Game created",
			"game_id": gameID,
		})
	})

	router.POST("/api/session", func(c *gin.Context) {
		c.SetCookie("session_id", random.RandomString(32), 72*60*60, "/", cfg.Host, false, true)

		c.JSON(200, gin.H{
			"message": "Session created successfully",
		})
	})

	router.Run(":8080")
}
