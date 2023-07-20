package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/controllers"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

func main() {
	var gameEngine game.Engine = nil

	r := gin.Default()
	r.POST("/game/:hostAddress", controllers.NewInitializeGameHandler(gameEngine))
	r.DELETE("/game/:hostAddress", controllers.NewCancelGameHandler(gameEngine))
	r.POST("/game/:hostAddress/join", controllers.NewJoinHandler(gameEngine))
	r.POST("/game/:hostAddress/phase/execute", controllers.NewPhaseExecutionHandler(gameEngine))
	r.GET("/game/:hostAddress/phase/wait", controllers.NewPhaseExecutionWaitHandler())
	r.GET("/game/:hostAddress/players", controllers.NewGetPlayersHandler(gameEngine))
	r.GET("/game/:hostAddress/players/:playerAddress", controllers.NewGetPlayerHandler(gameEngine))
	r.POST("/game/:hostAddress/players/:accuserAddress/vote/accuse", controllers.NewMafiaAccusationHandler(gameEngine))
	r.POST("/game/:hostAddress/players/:killerAddress/vote/kill", controllers.NewKillVoteHandler(gameEngine))
	r.POST("/game/:hostAddress/start", controllers.NewStartGameHandler(gameEngine))
	r.GET("/game/:hostAddress/start/wait", controllers.NewGameStartWaitHandler(gameEngine))
	r.Run("0.0.0.0:3000")
}
