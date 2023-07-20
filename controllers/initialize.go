package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewInitializeGameHandler creates a new handler to initialize a game
func NewInitializeGameHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.InitializeGame(c.Request.Context(), hostAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}

		c.Status(http.StatusOK)
	}
}
