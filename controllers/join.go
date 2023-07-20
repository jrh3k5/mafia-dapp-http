package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewJoinHandler creates a handler used to join an initialized game
func NewJoinHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		playerNickname := c.Query("playerNickname")
		if playerNickname == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.JoinGame(c.Request.Context(), hostAddress, playerNickname); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}

		c.Status(http.StatusOK)
	}
}
