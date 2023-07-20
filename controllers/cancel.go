package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewCancelGameHandler builds a handler for handling the cancellation of games
func NewCancelGameHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.CancelGame(c.Request.Context(), hostAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}
