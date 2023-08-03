package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jrh3k5/mafia-dapp-http/controllers"
	"github.com/jrh3k5/mafia-dapp-http/game"
)

func NewServer() *gin.Engine {
	var gameEngine game.Engine = game.NewInMemoryGameEngine()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		// allow everything
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")

		c.Next()
	})

	r.POST("/game/:hostAddress", controllers.NewInitializeGameHandler(gameEngine))
	r.DELETE("/game/:hostAddress", controllers.NewCancelGameHandler(gameEngine))
	r.OPTIONS("/game/:hostAddress", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", http.MethodDelete)
		c.Status(http.StatusOK)
	})
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
