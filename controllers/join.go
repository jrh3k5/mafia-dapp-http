package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"

	"errors"
)

// NewJoinHandler creates a handler used to join an initialized game
func NewJoinHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("hostAddress must be supplied"))
			return
		}

		playerAddress := c.Query("playerAddress")
		if playerAddress == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("playerAddress must be supplied"))
			return
		}

		playerNickname := c.Query("playerNickname")
		if playerNickname == "" {
			c.AbortWithError(http.StatusBadRequest, errors.New("playerNickname must be supplied"))
			return
		}

		if err := gameEngine.JoinGame(c.Request.Context(), hostAddress, playerAddress, playerNickname); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}

		c.Status(http.StatusOK)
	}
}
