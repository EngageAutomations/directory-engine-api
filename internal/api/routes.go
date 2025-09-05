package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"marketplace-app/internal/api/handlers"
	"marketplace-app/internal/api/middleware"
	"marketplace-app/internal/services"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, services *services.Services) {
	// Setup CORS
	setupCORS(router)

	// Setup middleware
	setupMiddleware(router, services)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(services)
	businessHandler := handlers.NewBusinessHandler(services)
	// TODO: Implement TokenHandler
	// tokenHandler := handlers.NewTokenHandler(services)
	adminHandler := handlers.NewAdminHandler(services)
	healthHandler := handlers.NewHealthHandler(services)
	webhookHandler := handlers.NewWebhookHandler(services)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health and status endpoints
		v1.GET("/health", healthHandler.BasicHealth)
		v1.GET("/status", healthHandler.DetailedHealth)

		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.GET("/oauth/callback", authHandler.HandleOAuthCallback)
			auth.POST("/oauth/exchange", authHandler.ExchangeToken)
			auth.GET("/oauth/url", authHandler.GetAuthURL)
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(services))
		{
			// Company routes
			companies := protected.Group("/companies")
			{
				companies.GET("/:companyId", businessHandler.GetCompany)
				companies.GET("/:companyId/locations", businessHandler.GetLocations)
				companies.POST("/:companyId/sync", businessHandler.SyncCompanyData)
			}

			// Location routes
			locations := protected.Group("/locations")
			{
				locations.GET("/:locationId", businessHandler.GetLocation)
				// TODO: Implement UpdateLocation method
				// locations.PUT("/:locationId", businessHandler.UpdateLocation)
				locations.GET("/:locationId/contacts", businessHandler.GetContacts)
				locations.POST("/:locationId/contacts", businessHandler.CreateContact)
				locations.GET("/:locationId/products", businessHandler.GetProducts)
				locations.POST("/:locationId/products", businessHandler.CreateProduct)
			}

			// TODO: Implement token management routes
			// tokens := protected.Group("/tokens")
			// {
			//	tokens.GET("/status/:companyId", tokenHandler.GetTokenStatus)
			//	tokens.POST("/refresh/:companyId", tokenHandler.RefreshToken)
			//	tokens.GET("/validate/:companyId", tokenHandler.ValidateToken)
			// }
		}

		// Admin routes (require admin authentication)
		admin := v1.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			// TODO: Implement token management endpoints
			// admin.GET("/tokens/all", tokenHandler.GetAllTokenStatuses)
			// admin.POST("/tokens/refresh-all", tokenHandler.RefreshAllTokens)
			// admin.POST("/tokens/cleanup", tokenHandler.CleanupExpiredTokens)
			admin.GET("/scheduler/stats", adminHandler.GetSchedulerStatus)
			admin.POST("/scheduler/run-refresh", adminHandler.RefreshAllTokens)
			admin.POST("/scheduler/run-cleanup", adminHandler.CleanupExpiredTokens)
			admin.GET("/cache/stats", adminHandler.GetCacheStats)
			admin.POST("/cache/flush", adminHandler.ClearCache)
			admin.GET("/system/health", adminHandler.GetSystemHealth)
		}
	}

	// Webhook routes (for external integrations)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/nango/token-refresh", webhookHandler.GenericWebhook)
		webhooks.POST("/nango/company-update", webhookHandler.GenericWebhook)
		webhooks.GET("/health", webhookHandler.WebhookHealth)
	}

	// Start background services
	if err := services.Start(); err != nil {
		panic("Failed to start background services: " + err.Error())
	}
}

// setupCORS configures Cross-Origin Resource Sharing
func setupCORS(router *gin.Engine) {
	config := cors.Config{
		AllowOrigins:     []string{"*"}, // Configure this properly for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(config))
}

// setupMiddleware configures global middleware
func setupMiddleware(router *gin.Engine, services *services.Services) {
	// Request logging
	router.Use(gin.Logger())

	// Recovery middleware
	router.Use(gin.Recovery())

	// Rate limiting middleware
	router.Use(middleware.RateLimitMiddleware(services))

	// Request ID middleware
	router.Use(middleware.RequestIDMiddleware())

	// Security headers middleware
	router.Use(middleware.SecurityHeadersMiddleware())

	// Error handling middleware
	router.Use(middleware.ErrorHandlingMiddleware())
}