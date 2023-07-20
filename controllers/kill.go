package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewKillVoteHandler builds a handler to handle a vote to kill a player
func NewKillVoteHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		killerAddress := c.Param("killerAddress")
		if killerAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		victimAddress := c.Query("player")
		if victimAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.VoteToKill(c.Request.Context(), hostAddress, killerAddress, victimAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}
