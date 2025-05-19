package http

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/OddOneOutApp/backend/internal/config"
	"github.com/OddOneOutApp/backend/internal/services"
	"github.com/OddOneOutApp/backend/internal/utils"
	"github.com/OddOneOutApp/backend/internal/websocket"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Initialize(db *gorm.DB, cfg *config.Config) {
	router := gin.Default()

	router.Use(ginzap.Ginzap(utils.RawLogger, time.RFC3339, true))

	router.Use(func(c *gin.Context) {
		regex := regexp.MustCompile(`^/api/games(/[a-zA-Z0-9]+/join)?$`)
		path := c.Request.URL.Path

		if path == "/api/categories" {
			c.Next()
			return
		}

		if (regex.MatchString(path)) && c.Request.Method == "POST" {
			sessionID, err := c.Cookie("session_id")
			if err == nil {
				session, err := services.GetSessionBySessionID(db, sessionID)
				if err == nil {
					c.Set("session", session)
					c.Next()
					return
				}
			}
			requestBody := struct {
				Username string `json:"username"`
			}{}
			err = json.NewDecoder(c.Request.Body).Decode(&requestBody)
			if err != nil {
				utils.Logger.Errorf("Error decoding request body: %v", err)
				c.JSON(400, gin.H{
					"error": "Invalid request body",
				})
				c.Abort()
				return
			}

			session, err := services.CreateSession(db, cfg, requestBody.Username)
			if err != nil {
				utils.Logger.Errorf("Error creating session: %v", err)
				c.JSON(500, gin.H{
					"error": "Internal server error",
				})
				c.Abort()
				return
			}
			c.SetCookie("session_id", session.SessionID, 72*60*60, "/", cfg.Host, cfg.Secure, true)
			utils.Logger.Debugf("New session created with ID: %s", session.SessionID)
			c.Set("session", session)
			c.Next()
			return
		}
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.SetCookie("session_id", "", -1, "/", cfg.Host, cfg.Secure, true)
			utils.Logger.Debugf("Session ID cookie not found, cleared")
			c.JSON(401, gin.H{
				"error": "Session ID cookie not found",
			})
			c.Redirect(302, "/")
			c.Abort()
			return
		}
		session, err := services.GetSessionBySessionID(db, sessionID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.SetCookie("session_id", "", -1, "/", cfg.Host, cfg.Secure, true)
				utils.Logger.Debugf("Session not found, cleared cookie")
				c.JSON(401, gin.H{
					"error": "Session not found",
				})
				c.Abort()
				c.Redirect(302, "/")
				return
			}
			utils.Logger.Errorf("Error fetching session: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			c.Redirect(302, "/")
			return
		}
		c.Set("session", session)
		c.Next()
	})

	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/api/games", func(c *gin.Context) {
		session, ok := getSessionFromContext(c)
		if !ok {
			return
		}

		type createGameRequest struct {
			Category string `json:"category"`
			Username string `json:"username"`
		}

		var requestBody createGameRequest
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Make sure category is not empty
		if requestBody.Category == "" {
			c.JSON(400, gin.H{
				"error": "Category is required",
			})
			return
		}

		game, err := services.CreateGame(db, cfg, session.ID, requestBody.Category)
		if err != nil {
			if err.Error() == "user is already in a game" {
				c.JSON(400, gin.H{
					"error": "You are already in the game",
				})
				return
			}
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			utils.Logger.Errorf("Error creating game: %v", err)
			return
		}
		utils.Logger.Infof("Game created with ID: %s for session ID: %s", game.ID, session.ID)
		c.JSON(200, gin.H{
			"message": "Game created successfully",
			"data": gin.H{
				"game_id": game.ID,
			},
		})
	})

	router.POST("/api/games/:game_id/join", func(c *gin.Context) {
		session, ok := getSessionFromContext(c)
		if !ok {
			return
		}

		gameID := c.Param("game_id")
		game, err := services.GetGameByID(db, gameID)
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

		_, err = game.Join(db, session.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "Game not found",
				})
				return
			}
			if err.Error() == "user is already in the game" {
				c.JSON(400, gin.H{
					"error": "You are already in the game",
				})
				return
			}
			utils.Logger.Errorf("Error joining game: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		utils.Logger.Infof("User with session ID: %s joined game with ID: %s", session.SessionID, gameID)
		c.JSON(200, gin.H{
			"message": "Joined game successfully",
			"data": gin.H{
				"game_id": gameID,
			},
		})
	})

	router.POST("/api/games/:game_id/leave", func(c *gin.Context) {
		session, ok := getSessionFromContext(c)
		if !ok {
			return
		}
		gameID := c.Param("game_id")
		game, err := services.GetGameByID(db, gameID)
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
		isHost, err := game.IsHost(db, session.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(404, gin.H{
					"error": "Game not found",
				})
				return
			}
			if err.Error() == "user is not in the game" {
				c.JSON(400, gin.H{
					"error": "You are not in the game",
				})
				return
			}
			utils.Logger.Errorf("Error checking if user is host: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		if isHost {
			err = game.Delete(db)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(404, gin.H{
						"error": "Game not found",
					})
					return
				}
				if err.Error() == "user is not in the game" {
					c.JSON(400, gin.H{
						"error": "You are not in the game",
					})
					return
				}
				utils.Logger.Errorf("Error deleting game: %v", err)
				c.JSON(500, gin.H{
					"error": "Internal server error",
				})
				return
			}
			websocket.SendGameDeleteMessage(gameID)
		} else {
			err = game.Leave(db, session.ID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(404, gin.H{
						"error": "Game not found",
					})
					return
				}
				if err.Error() == "user is not in the game" {
					c.JSON(400, gin.H{
						"error": "You are not in the game",
					})
					return
				}
				utils.Logger.Errorf("Error leaving game: %v", err)
				c.JSON(500, gin.H{
					"error": "Internal server error",
				})
				return
			}
		}

		utils.Logger.Infof("User with session ID: %s left game with ID: %s", session.SessionID, gameID)
		c.JSON(200, gin.H{
			"message": "Left game successfully",
			"data": gin.H{
				"game_id": gameID,
			},
		})
	})

	router.GET("/api/categories", func(c *gin.Context) {
		categories, err := services.GetAvailableCategories()
		if err != nil {
			utils.Logger.Errorf("Error fetching categories: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		c.JSON(200, gin.H{
			"data": gin.H{
				"categories": categories,
			},
		})
	})

	router.GET("/api/games/:game_id", func(c *gin.Context) {
		session, ok := getSessionFromContext(c)
		if !ok {
			return
		}
		gameID := c.Param("game_id")
		game, err := services.GetGameByID(db, gameID)
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
		// check if user is in game
		_, err = game.IsUserInGame(db, session.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(403, gin.H{
					"error": "You are not in this game",
				})
				return
			}
			utils.Logger.Errorf("Error checking user in game: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}

		// connect to websocket
		connection, err := websocket.NewConnection(c.Writer, c.Request)
		if err != nil {
			utils.Logger.Errorf("Error creating websocket connection: %v", err)
			c.JSON(500, gin.H{
				"error": "Internal server error",
			})
			return
		}
		websocket.HubInstance.AddConnection(gameID, connection, session.ID)

		websocket.SendInitMessage(gameID, session.ID, db)
		websocket.SendJoinMessage(gameID, session.ID, session.Username)
		websocket.SendUserStatusMessage(gameID, session.ID, true)

		go connection.ReadPump(websocket.HubInstance, gameID)
		go connection.WritePump()
		utils.Logger.Infof("WebSocket connection established for game ID: %s", gameID)

	})

	router.Run(":8080")
}

func getSessionFromContext(c *gin.Context) (*services.Session, bool) {
	sessionValue, exists := c.Get("session")
	if !exists {
		utils.Logger.Errorf("Session not found in context")
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
		return nil, false
	}

	session, ok := sessionValue.(*services.Session)
	if !ok {
		utils.Logger.Errorf("Invalid session type in context")
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
		return nil, false
	}

	return session, true
}
