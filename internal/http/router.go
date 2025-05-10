package http

import (
	"time"

	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/database/game"
	"github.com/OddOneOutApp/backend/internal/database/user"
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
			sessionID = random.RandomString(32)
			var requestBody struct {
				Username string `json:"username"`
			}

			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(400, gin.H{
					"error": "Invalid request body",
				})
				return
			}

			username := requestBody.Username

			userObj := &user.User{
				SessionID: sessionID,
				Username:  username,
			}
			err := db.Create(userObj).Error
			if err != nil {
				utils.Logger.Errorf("Error creating user: %v", err)
				c.JSON(500, gin.H{
					"error": "Internal server error",
				})
				return
			}
			c.SetCookie("session_id", sessionID, 72*60*60, "/", cfg.Host, cfg.Secure, true)
			utils.Logger.Debugf("User created with session ID: %s", sessionID)
		}
		var userObj user.User
		err = db.First(&userObj, "session_id = ?", sessionID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "User not found",
				})
				return
			}
			utils.Logger.Errorf("Error fetching user: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}

		gameID := random.RandomString(4)

		err = db.Create(&game.Game{
			ID:     gameID,
			HostID: userObj.ID,
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
			"message": "Game created successfully",
			"game_id": gameID,
		})
	})

	router.POST("/api/games/:game_id/join", func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			sessionID = random.RandomString(32)

			var requestBody struct {
				Username string `json:"username"`
			}

			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(400, gin.H{
					"error": "Invalid request body",
				})
				return
			}

			username := requestBody.Username

			err = db.Create(&user.User{
				SessionID: sessionID,
				Username:  username,
			}).Error
			if err != nil {
				utils.Logger.Errorf("Error creating user: %v", err)
				c.JSON(500, gin.H{
					"error": "Internal server error",
				})
				return
			}

			c.SetCookie("session_id", sessionID, 72*60*60, "/", cfg.Host, cfg.Secure, true)
			utils.Logger.Debugf("New session ID created: %s", sessionID)
		}

		var userObj user.User
		err = db.First(&userObj, "session_id = ?", sessionID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "User not found",
				})
				return
			}
			utils.Logger.Errorf("Error fetching user: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}

		gameID := c.Param("game_id")
		var gameData game.Game
		err = db.First(&gameData, "id = ?", gameID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "Game not found",
				})
				return
			}
			utils.Logger.Errorf("Error fetching game: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		if gameData.HostID == userObj.ID {
			c.JSON(400, gin.H{
				"error": "You cannot join your own game",
			})
			return
		}
		var userData user.User
		err = db.First(&userData, "session_id = ?", sessionID).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "User not found",
				})
				return
			}
			utils.Logger.Errorf("Error fetching user: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		if userData.ID == gameData.HostID {
			c.JSON(400, gin.H{
				"error": "You cannot join your own game",
			})
			return
		}
		err = db.Model(&gameData).Update("members", &[]string{"test"}).Error
		if err != nil {
			utils.Logger.Errorf("Error updating game members: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		utils.Logger.Infof("User with session ID: %s joined game with ID: %s", sessionID, gameID)
		c.JSON(200, gin.H{
			"message": "Joined game successfully",
			"game_id": gameID,
		})
	})

	router.Run(":8080")
}
