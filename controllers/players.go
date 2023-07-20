package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

// NewGetPlayerHandler provides a means of getting an individual player
func NewGetPlayerHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		playerAddress := c.Param("playerAddress")
		if playerAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		player, err := gameEngine.GetPlayer(c.Request.Context(), hostAddress, playerAddress)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		if player == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		playerRoleInt := int(player.PlayerRole)
		c.JSON(http.StatusOK, &playerResponse{
			PlayerAddress:  player.PlayerAddress,
			PlayerNickname: player.PlayerNickname,
			PlayerRole:     &playerRoleInt,
		})
	}
}

// NewGetPlayersHandler builds a handler for returning all players in a particular game
func NewGetPlayersHandler(gameEngine game.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostAddress := c.Param("hostAddress")
		if hostAddress == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		players, err := gameEngine.GetPlayers(c.Request.Context(), hostAddress)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		returnedPlayers := make([]*playerResponse, len(players))
		for playerIndex, player := range players {
			returnedPlayers[playerIndex] = &playerResponse{
				PlayerAddress:  player.PlayerAddress,
				PlayerNickname: player.PlayerNickname,
				// deliberately leave out the player role to not leak information
			}
		}

		c.JSON(http.StatusOK, returnedPlayers)
	}
}

type playerResponse struct {
	PlayerAddress  string `json:""`
	PlayerNickname string `json:""`
	PlayerRole     *int   `json:",omitempty"`
}
