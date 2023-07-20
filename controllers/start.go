package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewStartGameHandler builds a handler to start a game
func NewStartGameHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := gameEngine.StartGame(c.Request.Context(), hostAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}

// NewGameStartWaitHandler builds a handler to handle the waiting for a game start
func NewGameStartWaitHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		ctx, cancelFn := context.WithTimeout(c.Request.Context(), 10*time.Minute)
		defer cancelFn()

		if err := gameEngine.WaitForGameStart(ctx, hostAddress); err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusOK)
	}
}
