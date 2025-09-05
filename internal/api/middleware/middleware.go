package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"marketplace-app/internal/services"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("your-secret-key"), nil // Use config.JWTSecret in production
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Set user context
			c.Set("company_id", claims["company_id"])
			c.Set("user_id", claims["user_id"])
			c.Set("token_exp", claims["exp"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware validates admin access
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get admin token from header
		adminToken := c.GetHeader("X-Admin-Token")
		if adminToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin token required"})
			c.Abort()
			return
		}

		// Validate admin token (in production, use proper admin authentication)
		if adminToken != "admin-secret-token" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid admin token"})
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting using Redis
func RateLimitMiddleware(services *services.Services) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		clientID := getClientID(c)
		
		// Rate limit key
		rateKey := fmt.Sprintf("rate_limit:%s", clientID)
		
		// Get current request count
		currentCount, err := services.Cache.Increment(rateKey, 1)
		if err != nil {
			// If cache fails, allow the request but log the error
			fmt.Printf("Rate limit cache error: %v\n", err)
			c.Next()
			return
		}

		// Set expiration for the key (1 minute window)
		if currentCount == 1 {
			services.Cache.SetExpiration(rateKey, time.Minute)
		}

		// Check rate limit (100 requests per minute by default)
		rateLimit := int64(100) // This should come from config
		if currentCount > rateLimit {
			c.Header("X-RateLimit-Limit", strconv.FormatInt(rateLimit, 10))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(rateLimit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(rateLimit-currentCount, 10))

		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new UUID for request ID
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// ErrorHandlingMiddleware handles panics and errors gracefully
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		
		// Log the error with request ID
		fmt.Printf("PANIC [%v]: %v\n", requestID, recovered)

		// Return generic error response
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"request_id": requestID,
		})
	})
}

// CacheMiddleware adds caching for GET requests
func CacheMiddleware(services *services.Services, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Create cache key from URL and query parameters
		cacheKey := fmt.Sprintf("cache:%s:%s", c.Request.URL.Path, c.Request.URL.RawQuery)

		// Try to get cached response
		if cached := services.Cache.Get(cacheKey); cached != nil {
			if cachedResponse, ok := cached.(CachedResponse); ok {
				// Set cached headers
				for key, value := range cachedResponse.Headers {
					c.Header(key, value)
				}
				c.Header("X-Cache", "HIT")
				c.Data(cachedResponse.StatusCode, cachedResponse.ContentType, cachedResponse.Body)
				c.Abort()
				return
			}
		}

		// Create response writer wrapper to capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
			headers:        make(map[string]string),
		}
		c.Writer = writer

		c.Next()

		// Cache successful responses
		if writer.statusCode >= 200 && writer.statusCode < 300 {
			cachedResp := CachedResponse{
				StatusCode:  writer.statusCode,
				ContentType: writer.Header().Get("Content-Type"),
				Headers:     writer.headers,
				Body:        writer.body,
			}
			services.Cache.Set(cacheKey, cachedResp, duration)
		}

		c.Header("X-Cache", "MISS")
	}
}

// Helper types and functions

type CachedResponse struct {
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
}

type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	headers    map[string]string
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	// Capture headers
	for key, values := range w.Header() {
		if len(values) > 0 {
			w.headers[key] = values[0]
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func getClientID(c *gin.Context) string {
	// Try to get user ID from context first
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Fallback to IP address
	clientIP := c.ClientIP()
	return fmt.Sprintf("ip:%s", clientIP)
}