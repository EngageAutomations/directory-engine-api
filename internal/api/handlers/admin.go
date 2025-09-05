package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"marketplace-app/internal/services"
)

type AdminHandler struct {
	services *services.Services
}

func NewAdminHandler(services *services.Services) *AdminHandler {
	return &AdminHandler{
		services: services,
	}
}

// Token Management

// GetAllTokens returns status of all company tokens
func (h *AdminHandler) GetAllTokens(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get filter parameters
	status := c.Query("status") // "valid", "expired", "all"

	tokens, err := h.services.Token.GetAllTokenStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve token status",
			"details": err.Error(),
		})
		return
	}

	// Filter tokens based on status
	filteredTokens := make([]services.TokenExpiryInfo, 0)
	for _, token := range tokens {
		if status == "" || status == "all" {
			filteredTokens = append(filteredTokens, token)
		} else if status == "valid" && !token.IsExpired {
			filteredTokens = append(filteredTokens, token)
		} else if status == "expired" && token.IsExpired {
			filteredTokens = append(filteredTokens, token)
		}
	}

	// Apply pagination
	total := len(filteredTokens)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTokens := filteredTokens[start:end]

	c.JSON(http.StatusOK, gin.H{
		"tokens": paginatedTokens,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
		"filter": gin.H{
			"status": status,
		},
	})
}

// RefreshAllTokens triggers refresh for all expired tokens
func (h *AdminHandler) RefreshAllTokens(c *gin.Context) {
	// Get force parameter
	force := c.Query("force") == "true"

	var refreshed int
	var errors []string

	// Get all token statuses
	tokens, err := h.services.Token.GetAllTokenStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get token status",
			"details": err.Error(),
		})
		return
	}

	// Refresh tokens
	for _, token := range tokens {
		companyID := token.CompanyID
		isValid := !token.IsExpired

		// Refresh if expired or if force is true
		if !isValid || force {
			err := h.services.Token.RefreshTokenForCompany(companyID)
			if err != nil {
				errors = append(errors, companyID+": "+err.Error())
			} else {
				refreshed++
			}
		}
	}

	response := gin.H{
		"message": "Token refresh completed",
		"refreshed_count": refreshed,
		"total_tokens": len(tokens),
		"force_refresh": force,
		"timestamp": time.Now().Unix(),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["error_count"] = len(errors)
	}

	c.JSON(http.StatusOK, response)
}

// CleanupExpiredTokens removes old expired token records
func (h *AdminHandler) CleanupExpiredTokens(c *gin.Context) {
	// Get days parameter (default: 30 days)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 {
		days = 30
	}

	err := h.services.Token.CleanupExpiredTokens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup expired tokens",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expired tokens cleaned up successfully",
		"older_than_days": days,
		"timestamp": time.Now().Unix(),
	})
}

// Scheduler Management

// GetSchedulerStatus returns scheduler status and job information
func (h *AdminHandler) GetSchedulerStatus(c *gin.Context) {
	stats := h.services.Scheduler.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"scheduler": gin.H{
			"running": stats.IsRunning,
			"jobs_count": stats.JobCount,
			"last_run": stats.LastJobTime,
			"next_run": stats.NextJobTime,
		},
		"timestamp": time.Now().Unix(),
	})
}

// StartScheduler starts the scheduler service
func (h *AdminHandler) StartScheduler(c *gin.Context) {
	err := h.services.Scheduler.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start scheduler",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler started successfully",
		"timestamp": time.Now().Unix(),
	})
}

// StopScheduler stops the scheduler service
func (h *AdminHandler) StopScheduler(c *gin.Context) {
	h.services.Scheduler.Stop()

	c.JSON(http.StatusOK, gin.H{
		"message": "Scheduler stopped successfully",
		"timestamp": time.Now().Unix(),
	})
}

// AddSchedulerJob adds a new scheduled job
func (h *AdminHandler) AddSchedulerJob(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Schedule string `json:"schedule" binding:"required"`
		JobType  string `json:"job_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create job function based on type
	var jobFunc func()
	switch req.JobType {
	case "token_refresh":
		jobFunc = func() {
			h.services.Token.RefreshExpiredTokens()
		}
	case "token_cleanup":
		jobFunc = func() {
			h.services.Token.CleanupExpiredTokens()
		}
	case "health_check":
		jobFunc = func() {
			// Perform health checks
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job type"})
		return
	}

	_, err := h.services.Scheduler.AddJob(req.Schedule, jobFunc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add scheduled job",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Scheduled job added successfully",
		"job": gin.H{
			"name": req.Name,
			"schedule": req.Schedule,
			"job_type": req.JobType,
		},
		"timestamp": time.Now().Unix(),
	})
}

// RemoveSchedulerJob removes a scheduled job
func (h *AdminHandler) RemoveSchedulerJob(c *gin.Context) {
	// Note: Job removal by name is not currently supported
	// This would require maintaining a job name to EntryID mapping
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Job removal by name is not currently implemented",
		"message": "Use job entry ID for removal or restart scheduler to clear all jobs",
	})
}

// Cache Management

// GetCacheStats returns cache statistics
func (h *AdminHandler) GetCacheStats(c *gin.Context) {
	stats := h.services.Cache.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"cache": stats,
		"timestamp": time.Now().Unix(),
	})
}

// ClearCache clears all cache entries
func (h *AdminHandler) ClearCache(c *gin.Context) {
	// Get pattern parameter for selective clearing
	pattern := c.Query("pattern")

	var err error

	if pattern != "" {
		// Clear cache entries matching pattern
		err = h.services.Cache.DeletePattern(pattern)
	} else {
		// Clear all cache
		err = h.services.Cache.FlushAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cache",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"message": "Cache cleared successfully",
		"timestamp": time.Now().Unix(),
	}

	if pattern != "" {
		response["pattern"] = pattern
	}

	c.JSON(http.StatusOK, response)
}

// GetCacheHealth checks cache health
func (h *AdminHandler) GetCacheHealth(c *gin.Context) {
	health := h.services.Cache.Health()

	statusCode := http.StatusOK
	if !health["redis"] && !health["memory_cache"] {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"cache_health": health,
		"timestamp": time.Now().Unix(),
	})
}

// System Management

// GetSystemHealth returns overall system health
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	// Check cache health
	cacheHealth := h.services.Cache.Health()

	// Check scheduler health
	schedulerStats := h.services.Scheduler.GetStats()
	schedulerHealth := schedulerStats.IsRunning

	// Overall health status
	allHealthy := (cacheHealth["redis"] || cacheHealth["memory_cache"]) &&
		schedulerHealth

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"system_health": gin.H{
			"overall_healthy": allHealthy,
			"cache_healthy": cacheHealth,
			"scheduler_healthy": schedulerHealth,
		},
		"timestamp": time.Now().Unix(),
		"version": "1.0.0",
	})
}

// GetSystemStats returns system statistics
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	// Get business statistics
	companyCount := 0 // TODO: Implement count methods
	locationCount := 0 // TODO: Implement count methods
	contactCount := 0 // TODO: Implement count methods
	productCount := 0 // TODO: Implement count methods

	// Get token statistics
	tokens, _ := h.services.Token.GetAllTokenStatuses()
	validTokens := 0
	expiredTokens := 0
	for _, token := range tokens {
		if !token.IsExpired {
			validTokens++
		} else {
			expiredTokens++
		}
	}

	// Get cache statistics
	cacheStats := h.services.Cache.GetStats()

	// Get scheduler statistics
	schedulerStats := h.services.Scheduler.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"system_stats": gin.H{
			"business": gin.H{
				"companies": companyCount,
				"locations": locationCount,
				"contacts": contactCount,
				"products": productCount,
			},
			"tokens": gin.H{
				"total": len(tokens),
				"valid": validTokens,
				"expired": expiredTokens,
			},
			"cache": cacheStats,
			"scheduler": schedulerStats,
		},
		"timestamp": time.Now().Unix(),
	})
}

// RestartServices restarts all background services
func (h *AdminHandler) RestartServices(c *gin.Context) {
	// Stop services
	h.services.Stop()

	// Start services
	err := h.services.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restart services",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Services restarted successfully",
		"timestamp": time.Now().Unix(),
	})
}