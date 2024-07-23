package main

import (
	"fmt"
	"game-knowledge/backend/database"
	"game-knowledge/backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"path/filepath"
)

func main() {
	// Set up the router with CORS enabled
	router := setupRouter()

	// Set up the database connection
	err := database.ConnectDatabase()
	if err != nil {
		fmt.Printf("Failed to connect to database: %v", err)
		return
	}

	// Absolute path to the build directory
	buildPath := filepath.Join(".", "frontend", "build")

	// Serve static files
	router.Static("/static", filepath.Join(buildPath, "static"))

	// Serve the index.html file for the root route
	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(buildPath, "index.html"))
	})

	// Serve the index.html file for all other routes (client-side routing)
	router.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(buildPath, "index.html"))
	})

	// Setup API routes
	api := router.Group("/api")
	routes.SetupRoutes(api)

	// Run the server
	err = router.Run(":8081") // Change the port number here if needed
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
}

// setupRouter sets up the Gin router with CORS enabled
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	return router
}
