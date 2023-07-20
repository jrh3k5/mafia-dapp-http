package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewMafiaAccusationHandler builds a handler to handle an accusation of a player being Mafia
func NewMafiaAccusationHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		playerAddress := c.Param("accuserAddress")
		if playerAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		accuseeAddress := c.Query("player")
		if accuseeAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.AccuseAsMafia(c.Request.Context(), hostAddress, playerAddress, accuseeAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}
