package routes

import (
	"game-knowledge/backend/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.RouterGroup) {
	router.POST("/register", controllers.RegisterUser)
	router.POST("/login", controllers.LoginUser)
	router.POST("/predictions", controllers.CreatePrediction)
	router.GET("/matches", controllers.GetMatches)
	router.GET("/matches/:gameweek", controllers.GetMatchesForGameWeek)
	router.POST("/matches", controllers.CreateMatch)
	router.GET("/leaderboard", controllers.GetLeaderboard)
	router.GET("/profile", controllers.GetUserProfile)
	router.GET("/predictions/:matchId", controllers.GetPrediction)
}
