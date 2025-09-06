package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"marketplace-app/internal/api"
	"marketplace-app/internal/config"
	"marketplace-app/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Skip database initialization for testing
	// Initialize services without database
	services := services.NewServicesWithoutDB(cfg)

	// Initialize router
	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router, services)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Test server starting on port %s (without database)", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}