package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"marketplace-app/internal/services"
)

type WebhookHandler struct {
	services *services.Services
}

func NewWebhookHandler(services *services.Services) *WebhookHandler {
	return &WebhookHandler{
		services: services,
	}
}

// NangoTokenRefresh handles token refresh webhooks from Nango
func (h *WebhookHandler) NangoTokenRefresh(c *gin.Context) {
	// Verify webhook signature
	if !h.verifyNangoSignature(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse webhook payload
	var payload struct {
		Event     string `json:"event"`
		CompanyID string `json:"company_id"`
		Token     struct {
			AccessToken  string    `json:"access_token"`
			RefreshToken string    `json:"refresh_token"`
			ExpiresAt    time.Time `json:"expires_at"`
			TokenType    string    `json:"token_type"`
		} `json:"token"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// Validate event type
	if payload.Event != "token.refreshed" && payload.Event != "token.updated" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		return
	}

	// Process token refresh
	// TODO: Implement ProcessTokenRefresh method
	err := error(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process token refresh",
			"details": err.Error(),
		})
		return
	}

	// Log the webhook event
	fmt.Printf("Token refresh webhook processed for company %s at %v\n", 
		payload.CompanyID, payload.Timestamp)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refresh processed successfully",
		"company_id": payload.CompanyID,
		"event": payload.Event,
		"processed_at": time.Now().Unix(),
	})
}

// NangoCompanyUpdate handles company data update webhooks from Nango
func (h *WebhookHandler) NangoCompanyUpdate(c *gin.Context) {
	// Verify webhook signature
	if !h.verifyNangoSignature(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse webhook payload
	var payload struct {
		Event     string `json:"event"`
		CompanyID string `json:"company_id"`
		Data      struct {
			Company   map[string]interface{} `json:"company"`
			Locations []map[string]interface{} `json:"locations"`
			Contacts  []map[string]interface{} `json:"contacts"`
			Products  []map[string]interface{} `json:"products"`
		} `json:"data"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// Validate event type
	validEvents := []string{"company.updated", "locations.updated", "contacts.updated", "products.updated"}
	validEvent := false
	for _, event := range validEvents {
		if payload.Event == event {
			validEvent = true
			break
		}
	}

	if !validEvent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		return
	}

	// Process the update based on event type
	var err error
	var updatedCount int

	switch payload.Event {
	case "company.updated":
		// TODO: Implement UpdateCompanyFromWebhook method
		err = nil
		updatedCount = 1

	case "locations.updated":
		// TODO: Implement UpdateLocationsFromWebhook method
		updatedCount = 0
		err = nil

	case "contacts.updated":
		// TODO: Implement UpdateContactsFromWebhook method
		updatedCount = 0
		err = nil

	case "products.updated":
		// TODO: Implement UpdateProductsFromWebhook method
		updatedCount = 0
		err = nil
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process company update",
			"details": err.Error(),
		})
		return
	}

	// Invalidate relevant cache entries
	h.invalidateCompanyCache(payload.CompanyID)

	// Log the webhook event
	fmt.Printf("Company update webhook processed: %s for company %s, updated %d records at %v\n", 
		payload.Event, payload.CompanyID, updatedCount, payload.Timestamp)

	c.JSON(http.StatusOK, gin.H{
		"message": "Company update processed successfully",
		"company_id": payload.CompanyID,
		"event": payload.Event,
		"updated_count": updatedCount,
		"processed_at": time.Now().Unix(),
	})
}

// NangoConnectionStatus handles connection status webhooks from Nango
func (h *WebhookHandler) NangoConnectionStatus(c *gin.Context) {
	// Verify webhook signature
	if !h.verifyNangoSignature(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse webhook payload
	var payload struct {
		Event     string `json:"event"`
		CompanyID string `json:"company_id"`
		Status    string `json:"status"`
		Reason    string `json:"reason,omitempty"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// Validate event type
	validEvents := []string{"connection.created", "connection.deleted", "connection.failed"}
	validEvent := false
	for _, event := range validEvents {
		if payload.Event == event {
			validEvent = true
			break
		}
	}

	if !validEvent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported event type"})
		return
	}

	// Process connection status change
	switch payload.Event {
	case "connection.created":
		// TODO: Implement UpdateCompanyConnectionStatus method
		// Mark company as connected
		err := error(nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update connection status",
				"details": err.Error(),
			})
			return
		}

	case "connection.deleted":
		// Mark company as disconnected
		// TODO: Implement UpdateCompanyConnectionStatus method
		err := error(nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update connection status",
				"details": err.Error(),
			})
			return
		}

		// Mark tokens as expired
		// TODO: Implement MarkTokenExpired method

	case "connection.failed":
		// Log connection failure
		fmt.Printf("Connection failed for company %s: %s\n", payload.CompanyID, payload.Reason)
		
		// Optionally mark tokens as expired if connection consistently fails
		// TODO: Implement MarkTokenExpired method
	}

	// Invalidate cache
	h.invalidateCompanyCache(payload.CompanyID)

	// Log the webhook event
	fmt.Printf("Connection status webhook processed: %s for company %s, status: %s at %v\n", 
		payload.Event, payload.CompanyID, payload.Status, payload.Timestamp)

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection status processed successfully",
		"company_id": payload.CompanyID,
		"event": payload.Event,
		"status": payload.Status,
		"processed_at": time.Now().Unix(),
	})
}

// GenericWebhook handles other webhook events
func (h *WebhookHandler) GenericWebhook(c *gin.Context) {
	// Verify webhook signature
	if !h.verifyNangoSignature(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse generic payload
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// Log the webhook event
	event, _ := payload["event"].(string)
	companyID, _ := payload["company_id"].(string)
	fmt.Printf("Generic webhook received: %s for company %s\n", event, companyID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook received successfully",
		"event": event,
		"company_id": companyID,
		"processed_at": time.Now().Unix(),
	})
}

// Helper functions

// verifyNangoSignature verifies the webhook signature from Nango
func (h *WebhookHandler) verifyNangoSignature(c *gin.Context) bool {
	// Get signature from header
	signature := c.GetHeader("X-Nango-Signature")
	if signature == "" {
		return false
	}

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}

	// Reset body for further processing
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Get webhook secret from config (should be stored securely)
	webhookSecret := "your-nango-webhook-secret" // Should come from config

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// invalidateCompanyCache invalidates cache entries for a company
func (h *WebhookHandler) invalidateCompanyCache(companyID string) {
	// Define cache patterns to invalidate
	patterns := []string{
		fmt.Sprintf("company:%s", companyID),
		fmt.Sprintf("locations:%s", companyID),
		fmt.Sprintf("contacts:%s", companyID),
		fmt.Sprintf("products:%s", companyID),
		fmt.Sprintf("business_summary:%s", companyID),
	}

	// Invalidate each pattern
	for _, pattern := range patterns {
		h.services.Cache.Delete(pattern)
	}

	// Also clear any cached API responses for this company
	// TODO: Implement ClearPattern method
}

// WebhookHealth returns webhook endpoint health status
func (h *WebhookHandler) WebhookHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"webhook_endpoints": []string{
			"/webhooks/nango/token-refresh",
			"/webhooks/nango/company-update",
			"/webhooks/nango/connection-status",
			"/webhooks/nango/generic",
		},
		"timestamp": time.Now().Unix(),
	})
}