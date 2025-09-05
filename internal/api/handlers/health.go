package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"marketplace-app/internal/services"
)

type HealthHandler struct {
	services  *services.Services
	startTime time.Time
}

func NewHealthHandler(services *services.Services) *HealthHandler {
	return &HealthHandler{
		services:  services,
		startTime: time.Now(),
	}
}

// BasicHealth returns basic health status
func (h *HealthHandler) BasicHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"uptime": time.Since(h.startTime).Seconds(),
		"version": "1.0.0",
	})
}

// DetailedHealth returns comprehensive health information
func (h *HealthHandler) DetailedHealth(c *gin.Context) {
	// Check database health (TODO: implement database health check)
	dbHealthy := true

	// Check cache health (TODO: implement cache health check)
	cacheHealth := true

	// Check scheduler health
	schedulerStats := h.services.Scheduler.GetStats()
	schedulerHealthy := schedulerStats.IsRunning

	// Get system metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate overall health
	allHealthy := dbHealthy && cacheHealth && schedulerHealthy

	// Determine HTTP status code
	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	// Build response
	response := gin.H{
		"status": func() string {
			if allHealthy {
				return "healthy"
			}
			return "unhealthy"
		}(),
		"timestamp": time.Now().Unix(),
		"uptime": time.Since(h.startTime).Seconds(),
		"version": "1.0.0",
		"components": gin.H{
			"database": gin.H{
				"status": func() string {
					if dbHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"healthy": dbHealthy,
			},
			"cache": gin.H{
				"status": func() string {
					if cacheHealth {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"healthy": cacheHealth,
			},
			"scheduler": gin.H{
				"status": func() string {
					if schedulerHealthy {
						return "healthy"
					}
					return "unhealthy"
				}(),
				"healthy": schedulerHealthy,
				"details": schedulerStats,
			},
		},
		"system": gin.H{
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"memory": gin.H{
				"alloc_mb": bToMb(memStats.Alloc),
				"total_alloc_mb": bToMb(memStats.TotalAlloc),
				"sys_mb": bToMb(memStats.Sys),
				"gc_cycles": memStats.NumGC,
			},
			"cpu_count": runtime.NumCPU(),
		},
	}

	c.JSON(statusCode, response)
}

// ReadinessCheck checks if the application is ready to serve traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check database health (TODO: implement database health check)
	dbHealthy := true

	// Check cache health (TODO: implement cache health check)
	cacheHealthy := true

	// Application is ready if database and cache are available
	ready := dbHealthy && cacheHealthy

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"ready": ready,
		"timestamp": time.Now().Unix(),
		"checks": gin.H{
			"database": dbHealthy,
			"cache": cacheHealthy,
		},
	})
}

// LivenessCheck checks if the application is alive
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"alive": true,
		"timestamp": time.Now().Unix(),
		"uptime": time.Since(h.startTime).Seconds(),
	})
}

// MetricsEndpoint returns application metrics
func (h *HealthHandler) MetricsEndpoint(c *gin.Context) {
	// Get business metrics (TODO: implement count methods)
	companyCount := 0
	locationCount := 0
	contactCount := 0
	productCount := 0

	// Get token metrics
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

	// Get cache metrics
	cacheStats := h.services.Cache.GetStats()

	// Get scheduler metrics
	schedulerStats := h.services.Scheduler.GetStats()

	// Get system metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"metrics": gin.H{
			"business": gin.H{
				"companies_total": companyCount,
				"locations_total": locationCount,
				"contacts_total": contactCount,
				"products_total": productCount,
			},
			"tokens": gin.H{
				"total": len(tokens),
				"valid": validTokens,
				"expired": expiredTokens,
				"valid_percentage": func() float64 {
					if len(tokens) == 0 {
						return 0
					}
					return float64(validTokens) / float64(len(tokens)) * 100
				}(),
			},
			"cache": cacheStats,
			"scheduler": schedulerStats,
			"system": gin.H{
				"uptime_seconds": time.Since(h.startTime).Seconds(),
				"goroutines": runtime.NumGoroutine(),
				"memory_alloc_mb": bToMb(memStats.Alloc),
				"memory_sys_mb": bToMb(memStats.Sys),
				"gc_cycles": memStats.NumGC,
				"cpu_count": runtime.NumCPU(),
			},
		},
		"timestamp": time.Now().Unix(),
	})
}

// DatabaseHealth checks database connectivity
func (h *HealthHandler) DatabaseHealth(c *gin.Context) {
	// TODO: Implement database health check
	healthy := true

	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"database": gin.H{
			"healthy": healthy,
			"status": func() string {
				if healthy {
					return "connected"
				}
				return "disconnected"
			}(),
		},
		"timestamp": time.Now().Unix(),
	})
}

// CacheHealth checks cache connectivity
func (h *HealthHandler) CacheHealth(c *gin.Context) {
	// TODO: Implement cache health check
	health := true

	statusCode := http.StatusOK
	if !health {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"cache": gin.H{"healthy": health},
		"timestamp": time.Now().Unix(),
	})
}

// SchedulerHealth checks scheduler status
func (h *HealthHandler) SchedulerHealth(c *gin.Context) {
	stats := h.services.Scheduler.GetStats()
	healthy := stats.IsRunning

	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"scheduler": gin.H{
			"healthy": healthy,
			"status": func() string {
				if healthy {
					return "running"
				}
				return "stopped"
			}(),
			"details": stats,
		},
		"timestamp": time.Now().Unix(),
	})
}

// Helper function to convert bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}