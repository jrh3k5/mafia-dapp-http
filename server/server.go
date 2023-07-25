package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/controllers"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

func NewServer() *gin.Engine {
	var gameEngine game.Engine = game.NewInMemoryGameEngine()

	r := gin.Default()
	r.POST("/game/:hostAddress", controllers.NewInitializeGameHandler(gameEngine))
	r.DELETE("/game/:hostAddress", controllers.NewCancelGameHandler(gameEngine))
	r.POST("/game/:hostAddress/join", controllers.NewJoinHandler(gameEngine))
	r.POST("/game/:hostAddress/phase/execute", controllers.NewPhaseExecutionHandler(gameEngine))
	r.GET("/game/:hostAddress/phase/wait", controllers.NewPhaseExecutionWaitHandler(gameEngine))
	r.GET("/game/:hostAddress/players", controllers.NewGetPlayersHandler(gameEngine))
	r.GET("/game/:hostAddress/players/:playerAddress", controllers.NewGetPlayerHandler(gameEngine))
	r.POST("/game/:hostAddress/players/:voterAddress/vote/:action", controllers.NewPlayerVoteHandler(gameEngine))
	r.POST("/game/:hostAddress/start", controllers.NewStartGameHandler(gameEngine))
	r.GET("/game/:hostAddress/start/wait", controllers.NewGameStartWaitHandler(gameEngine))

	return r
}