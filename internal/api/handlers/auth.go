package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"marketplace-app/internal/services"
)

type AuthHandler struct {
	services *services.Services
}

func NewAuthHandler(services *services.Services) *AuthHandler {
	return &AuthHandler{
		services: services,
	}
}

// GetAuthURL generates Nango OAuth URL for company authorization
func (h *AuthHandler) GetAuthURL(c *gin.Context) {
	var req struct {
		CompanyID   string `json:"company_id" binding:"required"`
		RedirectURL string `json:"redirect_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate state parameter for security
	state := uuid.New().String()

	// Store state in cache for validation (expires in 10 minutes)
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	h.services.Cache.Set(stateKey, req.CompanyID, 10*time.Minute)

	// Build Nango OAuth URL
	baseURL := "https://api.nango.dev/oauth/authorize"
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {"your-nango-client-id"}, // Should come from config
		"redirect_uri":  {req.RedirectURL},
		"scope":         {"read write"},
		"state":         {state},
		"company_id":    {req.CompanyID},
	}

	authURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
		"expires_at": time.Now().Add(10 * time.Minute).Unix(),
	})
}

// HandleOAuthCallback processes the OAuth callback from Nango
func (h *AuthHandler) HandleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	error := c.Query("error")

	// Check for OAuth errors
	if error != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "OAuth authorization failed",
			"details": error,
		})
		return
	}

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state parameter"})
		return
	}

	// Validate state parameter
	stateKey := fmt.Sprintf("oauth_state:%s", state)
	cachedCompanyID := h.services.Cache.Get(stateKey)
	if cachedCompanyID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired state parameter"})
		return
	}

	companyID, ok := cachedCompanyID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid state data"})
		return
	}

	// Process OAuth callback with Nango service
	result, err := h.services.Nango.ProcessOAuthCallback(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process OAuth callback",
			"details": err.Error(),
		})
		return
	}

	// Clean up state from cache
	h.services.Cache.Delete(stateKey)

	// Generate JWT token for the company
	jwtToken, err := h.generateJWTToken(companyID, result.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authorization successful",
		"company_id": companyID,
		"access_token": jwtToken,
		"token_type": "Bearer",
		"expires_in": 3600, // 1 hour
		"locations_synced": len(result.Locations),
	})
}

// ExchangeToken exchanges a company token for locations and data
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req struct {
		CompanyID    string `json:"company_id" binding:"required"`
		CompanyToken string `json:"company_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch locations using the company token
	locations, err := h.services.Nango.GetLocations(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch locations",
			"details": err.Error(),
		})
		return
	}

	// Generate JWT token
	jwtToken, err := h.generateJWTToken(req.CompanyID, req.CompanyToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token exchange successful",
		"access_token": jwtToken,
		"token_type": "Bearer",
		"expires_in": 3600,
		"locations": locations,
		"locations_count": len(locations),
	})
}

// RefreshToken manually refreshes tokens for a company
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Refresh token using token service
	err := h.services.Token.RefreshTokenForCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh token",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"company_id": companyID,
		"refreshed_at": time.Now().Unix(),
	})
}

// GetTokenStatus returns token status for a company
func (h *AuthHandler) GetTokenStatus(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Get token expiry information
	expiry, err := h.services.Token.GetTokenExpiryInfo(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get token status",
			"details": err.Error(),
		})
		return
	}

	// Check if token is valid
	isValid, _ := h.services.Token.ValidateToken(companyID)

	c.JSON(http.StatusOK, gin.H{
		"company_id": companyID,
		"is_valid": isValid,
		"expires_at": expiry.TokenExpiry.Unix(),
		"expires_in": int64(time.Until(expiry.TokenExpiry).Seconds()),
		"status": func() string {
			if isValid {
				return "active"
			}
			return "expired"
		}(),
	})
}

// ValidateToken validates the current JWT token
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	// Get token from context (set by auth middleware)
	companyID, exists := c.Get("company_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token context"})
		return
	}

	tokenExp, exists := c.Get("token_exp")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token expiration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"company_id": companyID,
		"expires_at": tokenExp,
		"validated_at": time.Now().Unix(),
	})
}

// Helper function to generate JWT tokens
func (h *AuthHandler) generateJWTToken(companyID, companyToken string) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"company_id": companyID,
		"user_id": companyID, // Using company_id as user_id for simplicity
		"company_token": companyToken,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(), // 1 hour expiration
		"iss": "marketplace-app",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key (should come from config)
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}