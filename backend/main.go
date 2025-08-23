package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"context"
	"syscall"
	"time"

	"game-knowledge/backend/database"
	"game-knowledge/backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables
	if err := godotenv.Load("properties.env"); err != nil {
		log.Println("⚠️  Warning: properties.env file not found, using system environment variables")
	} else {
		log.Println("✅ Environment variables loaded from properties.env")
	}

	// Log important environment variables (don't log secrets!)
	log.Printf("🏈 League ID: %s", os.Getenv("LEAGUE_ID"))
	log.Printf("📅 Season: %s", os.Getenv("SEASON"))
	log.Printf("🌐 Server Mode: %s", getEnvWithDefault("GIN_MODE", "debug"))

	// Check if API key is set (don't log the actual key)
	if os.Getenv("API_FOOTBALL_KEY") != "" {
		log.Println("✅ Football API key is configured")
	} else {
		log.Println("⚠️  Warning: API_FOOTBALL_KEY not set - match fetching will fail")
	}
}

func main() {
	log.Println("🚀 Starting Ball Knowledge API Server...")

	// Set Gin mode
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	// Initialize database
	if err := database.ConnectDatabase(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.CloseDatabase(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Setup router
	router := setupRouter()

	// Setup static file serving for frontend
	setupStaticRoutes(router)

	// Setup API routes
	api := router.Group("/api")
	routes.SetupRoutes(api)

	// Get port from environment
	port := getEnvWithDefault("PORT", "8081")

	// Start server
	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("🔗 API available at: http://localhost:%s/api", port)
	log.Printf("📊 Health check: http://localhost:%s/api/health", port)

	if gin.Mode() == gin.DebugMode {
		log.Printf("📝 API Documentation:")
		log.Printf("   POST /api/register          - Register new user")
		log.Printf("   POST /api/login             - User login")
		log.Printf("   GET  /api/matches           - Get all matches")
		log.Printf("   POST /api/predictions       - Create prediction (auth)")
		log.Printf("   GET  /api/leaderboard       - View leaderboard")
		log.Printf("   GET  /api/profile           - Get user profile (auth)")
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}

// setupRouter configures the Gin router with middleware
func setupRouter() *gin.Engine {
	router := gin.New()

	// Custom logging middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Recovery middleware
	router.Use(gin.Recovery())

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}))

	// Security headers middleware
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})

	return router
}

// setupStaticRoutes configures static file serving for the frontend
func setupStaticRoutes(router *gin.Engine) {
	buildPath := filepath.Join(".", "frontend", "build")

	// Check if build directory exists
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		log.Printf("⚠️  Frontend build directory not found at %s", buildPath)
		return
	}

	// Serve static files
	router.Static("/static", filepath.Join(buildPath, "static"))

	// Serve React app for root route
	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(buildPath, "index.html"))
	})

	// Handle client-side routing (SPA)
	router.NoRoute(func(c *gin.Context) {
		// Don't handle API routes here
		if gin.IsDebugging() && c.Request.URL.Path != "/" {
			log.Printf("404 - Route not found: %s %s", c.Request.Method, c.Request.URL.Path)
		}

		// For non-API routes, serve the React app
		if !isAPIRoute(c.Request.URL.Path) {
			c.File(filepath.Join(buildPath, "index.html"))
		} else {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
		}
	})

	log.Println("✅ Frontend static files configured")
}

// isAPIRoute checks if a path is an API route
func isAPIRoute(path string) bool {
	return len(path) >= 4 && path[:4] == "/api"
}

// getEnvWithDefault returns environment variable or default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}