package controllers

import (
	"net/http"

	"errors"

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
		c.AbortWithError(http.StatusBadRequest, errors.New("host address must be supplidd"))
		return
	}

	voterAddress := c.Param("voterAddress")
	if voterAddress == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("voter address must be supplied"))
		return
	}

	accuseeAddress := c.Query("playerAddress")
	if accuseeAddress == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("vote receipient must be supplied"))
		return
	}

	if err := gameEngine.AccuseAsMafia(c.Request.Context(), hostAddress, voterAddress, accuseeAddress); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

func handleKillVote(c *gin.Context, gameEngine game.Engine) {
	hostAddress := c.Param("hostAddress")
	if hostAddress == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hostAddress must be supplied"))
		return
	}

	killerAddress := c.Param("voterAddress")
	if killerAddress == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("voterAddress must be supplied"))
		return
	}

	victimAddress := c.Query("playerAddress")
	if victimAddress == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("playerAddress must be supplied"))
		return
	}

	if err := gameEngine.VoteToKill(c.Request.Context(), hostAddress, killerAddress, victimAddress); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
