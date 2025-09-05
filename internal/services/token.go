package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"marketplace-app/internal/models"
)

type TokenService struct {
	db    *gorm.DB
	nango *NangoService
}

func NewTokenService(db *gorm.DB, nango *NangoService) *TokenService {
	return &TokenService{
		db:    db,
		nango: nango,
	}
}

// RefreshExpiredTokens finds and refreshes tokens that are about to expire
func (ts *TokenService) RefreshExpiredTokens() error {
	log.Println("Starting token refresh job...")

	// Find tokens that need refreshing (within 24 hours of expiry)
	var tokenRefreshes []models.TokenRefresh
	err := ts.db.Where("next_refresh <= ? AND status = ?", time.Now(), "active").
		Preload("Company").
		Find(&tokenRefreshes).Error

	if err != nil {
		return fmt.Errorf("failed to fetch tokens for refresh: %w", err)
	}

	log.Printf("Found %d tokens to refresh", len(tokenRefreshes))

	successCount := 0
	failureCount := 0

	for _, tokenRefresh := range tokenRefreshes {
		if err := ts.refreshSingleToken(&tokenRefresh); err != nil {
			log.Printf("Failed to refresh token for company %s: %v", 
				tokenRefresh.Company.CompanyID, err)
			failureCount++
			
			// Update token refresh record with error
			tokenRefresh.Status = "failed"
			tokenRefresh.ErrorMessage = err.Error()
			ts.db.Save(&tokenRefresh)
		} else {
			log.Printf("Successfully refreshed token for company %s", 
				tokenRefresh.Company.CompanyID)
			successCount++
		}
	}

	log.Printf("Token refresh job completed. Success: %d, Failures: %d", 
		successCount, failureCount)

	return nil
}

// RefreshTokenForCompany manually refreshes token for a specific company
func (ts *TokenService) RefreshTokenForCompany(companyID string) error {
	company := &models.Company{}
	err := ts.db.Where("company_id = ?", companyID).First(company).Error
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Check if token actually needs refreshing
	if time.Until(company.TokenExpiry) > 24*time.Hour {
		return fmt.Errorf("token for company %s does not need refreshing yet", companyID)
	}

	return ts.nango.RefreshToken(companyID)
}

// ValidateToken checks if a token is still valid
func (ts *TokenService) ValidateToken(companyID string) (bool, error) {
	company := &models.Company{}
	err := ts.db.Where("company_id = ? AND is_active = ?", companyID, true).First(company).Error
	if err != nil {
		return false, fmt.Errorf("company not found: %w", err)
	}

	// Check if token is expired
	if time.Now().After(company.TokenExpiry) {
		return false, nil
	}

	// Check if token expires within the next hour (consider it invalid for safety)
	if time.Until(company.TokenExpiry) < time.Hour {
		return false, nil
	}

	return true, nil
}

// GetTokenExpiryInfo returns token expiry information for a company
func (ts *TokenService) GetTokenExpiryInfo(companyID string) (*TokenExpiryInfo, error) {
	company := &models.Company{}
	err := ts.db.Where("company_id = ?", companyID).First(company).Error
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	tokenRefresh := &models.TokenRefresh{}
	err = ts.db.Where("company_id = ?", company.ID).First(tokenRefresh).Error
	if err != nil {
		return nil, fmt.Errorf("token refresh record not found: %w", err)
	}

	return &TokenExpiryInfo{
		CompanyID:    companyID,
		CompanyName:  company.CompanyName,
		TokenExpiry:  company.TokenExpiry,
		TimeToExpiry: time.Until(company.TokenExpiry),
		IsExpired:    time.Now().After(company.TokenExpiry),
		NeedsRefresh: time.Until(company.TokenExpiry) < 24*time.Hour,
		LastRefresh:  tokenRefresh.LastRefresh,
		NextRefresh:  tokenRefresh.NextRefresh,
		RefreshCount: tokenRefresh.RefreshCount,
		Status:       tokenRefresh.Status,
	}, nil
}

// GetAllTokenStatuses returns token status for all companies
func (ts *TokenService) GetAllTokenStatuses() ([]TokenExpiryInfo, error) {
	var companies []models.Company
	err := ts.db.Where("is_active = ?", true).Find(&companies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch companies: %w", err)
	}

	var statuses []TokenExpiryInfo
	for _, company := range companies {
		info, err := ts.GetTokenExpiryInfo(company.CompanyID)
		if err != nil {
			log.Printf("Failed to get token info for company %s: %v", 
				company.CompanyID, err)
			continue
		}
		statuses = append(statuses, *info)
	}

	return statuses, nil
}

// MarkTokenAsExpired marks a token as expired (for manual intervention)
func (ts *TokenService) MarkTokenAsExpired(companyID string) error {
	company := &models.Company{}
	err := ts.db.Where("company_id = ?", companyID).First(company).Error
	if err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Update company status
	company.IsActive = false
	if err := ts.db.Save(company).Error; err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	// Update token refresh record
	tokenRefresh := &models.TokenRefresh{}
	err = ts.db.Where("company_id = ?", company.ID).First(tokenRefresh).Error
	if err == nil {
		tokenRefresh.Status = "expired"
		tokenRefresh.ErrorMessage = "Manually marked as expired"
		ts.db.Save(tokenRefresh)
	}

	return nil
}

// CleanupExpiredTokens removes old expired token records
func (ts *TokenService) CleanupExpiredTokens() error {
	// Delete token refresh records older than 30 days with failed/expired status
	cutoffDate := time.Now().AddDate(0, 0, -30)

	result := ts.db.Where("updated_at < ? AND status IN ?", 
		cutoffDate, []string{"failed", "expired"}).Delete(&models.TokenRefresh{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	log.Printf("Cleaned up %d expired token records", result.RowsAffected)
	return nil
}

// Private helper methods

func (ts *TokenService) refreshSingleToken(tokenRefresh *models.TokenRefresh) error {
	// Attempt to refresh the token
	err := ts.nango.RefreshToken(tokenRefresh.Company.CompanyID)
	if err != nil {
		return fmt.Errorf("nango refresh failed: %w", err)
	}

	// Update the token refresh record
	tokenRefresh.LastRefresh = time.Now()
	tokenRefresh.RefreshCount++
	tokenRefresh.Status = "active"
	tokenRefresh.ErrorMessage = ""

	// Get updated company to set next refresh time
	updatedCompany := &models.Company{}
	err = ts.db.Where("id = ?", tokenRefresh.CompanyID).First(updatedCompany).Error
	if err != nil {
		return fmt.Errorf("failed to get updated company: %w", err)
	}

	// Set next refresh to 24 hours before expiry
	tokenRefresh.NextRefresh = updatedCompany.TokenExpiry.Add(-24 * time.Hour)

	return ts.db.Save(tokenRefresh).Error
}

// TokenExpiryInfo holds information about token expiry status
type TokenExpiryInfo struct {
	CompanyID    string        `json:"company_id"`
	CompanyName  string        `json:"company_name"`
	TokenExpiry  time.Time     `json:"token_expiry"`
	TimeToExpiry time.Duration `json:"time_to_expiry"`
	IsExpired    bool          `json:"is_expired"`
	NeedsRefresh bool          `json:"needs_refresh"`
	LastRefresh  time.Time     `json:"last_refresh"`
	NextRefresh  time.Time     `json:"next_refresh"`
	RefreshCount int           `json:"refresh_count"`
	Status       string        `json:"status"`
}