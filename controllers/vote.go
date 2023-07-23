package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

func NewPlayerVoteHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		action := c.Param("action")
		switch action {
		case "accuse":
			handleAccusation(c, gameEngine)
		case "kill":
			handleKillVote(c, gameEngine)
		default:
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
}

func handleAccusation(c *gin.Context, gameEngine game.Engine) {
	hostAddress := c.Param("hostAddress")
	if hostAddress == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	playerAddress := c.Param("voterAddress")
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

func handleKillVote(c *gin.Context, gameEngine game.Engine) {
	hostAddress := c.Param("hostAddress")
	if hostAddress == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	killerAddress := c.Param("voterAddress")
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
