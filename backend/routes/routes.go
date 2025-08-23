package routes

import (
	"ball-knowledge/backend/controllers"
	"ball-knowledge/backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := router.Group("/")
	{
		// Authentication routes
		public.POST("/register", controllers.RegisterUser)
		public.POST("/login", controllers.LoginUser)

		// Public match data (optional: make these require auth)
		public.GET("/matches", controllers.GetMatches)
		public.GET("/matches/:gameweek", controllers.GetMatchesForGameWeek)
		public.GET("/matches/details/:id", controllers.GetMatchDetails)
		public.GET("/leaderboard", controllers.GetLeaderboard)
	}

	// Protected routes (authentication required)
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User profile
		protected.GET("/profile", controllers.GetUserProfile)
		protected.POST("/refresh-token", controllers.RefreshToken)

		// Predictions
		protected.POST("/predictions", controllers.CreatePrediction)
		protected.GET("/predictions/:matchId", controllers.GetPrediction)
		protected.GET("/my-predictions", controllers.GetUserPredictions)
		protected.PUT("/predictions/:id", controllers.UpdatePrediction)

		// Admin-only routes (you can add admin middleware later)
		protected.POST("/matches", controllers.CreateMatch)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Ball Knowledge API is running",
			"version": "1.0.0",
		})
	})
}