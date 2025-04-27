package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/the-digital-watchdog-initiative/routes"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

func main() {
	// Initialize logger first
	utils.InitLogger()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// Continue running even if .env file is not found
		// as environment variables might be set in the system
		// especially in production environments
		if os.Getenv("APP_ENV") != "production" {
			// Only warn in non-production environments
			// In production, environment variables are typically
			// set differently (e.g. in container settings)
			utils.Logger.Warn("No .env file found. Using system environment variables.")
		}
	}

	// Initialize database connection
	utils.InitDB()

	// Seed admin user
	if err := utils.SeedAdmin(utils.DB); err != nil {
		utils.Logger.Errorf("Failed to seed admin user: %v", err)
		// Continue running the application even if seeding fails
	}

	// Set Gin mode
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize router
	router := gin.New() // Use New() instead of Default() to avoid using the default logger

	// Use our custom logger middleware
	router.Use(utils.LoggerMiddleware())

	// Use gin recovery middleware to recover from panics
	router.Use(gin.Recovery())

	// Request timing middleware
	router.Use(func(c *gin.Context) {
		// Add request timestamp to context
		c.Set("requestTime", time.Now())
		c.Next()
	})

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	router.Use(cors.New(config))

	// Initialize service provider
	serviceProvider := services.NewServiceProvider()
	if err := serviceProvider.Initialize(); err != nil {
		utils.Logger.Fatalf("Failed to initialize service provider: %v", err)
	}

	// API routes
	api := router.Group("/api")
	{
		// Register all routes from routes package
		routes.RegisterRoutes(api, serviceProvider)
	}

	// Create an HTTP server with the configured router
	server := &http.Server{
		Addr:    ":" + getPort(),
		Handler: router,
	}

	// Start the server in a separate goroutine
	go func() {
		utils.Logger.Infof("Server running on port %s", getPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Accept SIGINT (Ctrl+C) and SIGTERM (kill)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Logger.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Logger.Fatalf("Server forced to shutdown: %v", err)
	}

	utils.Logger.Info("Server exited gracefully")
}

// getPort returns the port from environment or defaults to 8080
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
