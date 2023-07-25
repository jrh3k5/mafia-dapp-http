package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

func NewPhaseExecutionHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.ExecutePhase(c.Request.Context(), hostAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

func NewPhaseExecutionWaitHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		ctx, cancelFn := context.WithTimeout(c.Request.Context(), 10*time.Minute)
		defer cancelFn()

		phaseExecution, err := gameEngine.WaitForPhaseExecution(ctx, hostAddress)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, &phaseExecutionResponse{
			HostAddress:      phaseExecution.HostAddress,
			PhaseOutcome:     int(phaseExecution.PhaseOutcome),
			CurrentPhase:     int(phaseExecution.CurrentPhase),
			KilledPlayers:    phaseExecution.KilledPlayers,
			ConvictedPlayers: phaseExecution.ConvictedPlayers,
		})
	}
}

type phaseExecutionResponse struct {
	HostAddress      string   `json:"hostAddress"`
	PhaseOutcome     int      `json:"phaseOutcome"`
	CurrentPhase     int      `json:"currentPhase"`
	KilledPlayers    []string `json:"killedPlayers"`
	ConvictedPlayers []string `json:"convictedPlayers"`
}
