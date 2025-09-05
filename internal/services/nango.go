package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"
	"marketplace-app/internal/config"
	"marketplace-app/internal/models"
)

type NangoService struct {
	db     *gorm.DB
	config *config.Config
	client *http.Client
}

type NangoAuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CompanyID    string    `json:"company_id"`
	CompanyName  string    `json:"company_name"`
}

type NangoLocationResponse struct {
	LocationID    string `json:"location_id"`
	LocationToken string `json:"location_token"`
	BusinessName  string `json:"business_name"`
	BusinessType  string `json:"business_type"`
	Address       string `json:"address"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
	Country       string `json:"country"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Website       string `json:"website"`
}

type NangoContactResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Title     string `json:"title"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Mobile    string `json:"mobile"`
	IsPrimary bool   `json:"is_primary"`
}

type NangoProductResponse struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	SKU         string  `json:"sku"`
}

type NangoTokenRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	CompanyID    string `json:"company_id"`
}

type NangoTokenRefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewNangoService(db *gorm.DB, config *config.Config) *NangoService {
	return &NangoService{
		db:     db,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessOAuthCallback handles the OAuth callback from Nango
func (ns *NangoService) ProcessOAuthCallback(authCode string) (*models.Company, error) {
	// Exchange auth code for tokens
	authResp, err := ns.exchangeAuthCode(authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code: %w", err)
	}

	// Create or update company record
	company := &models.Company{
		CompanyID:    authResp.CompanyID,
		CompanyName:  authResp.CompanyName,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		TokenExpiry:  authResp.ExpiresAt,
		IsActive:     true,
	}

	// Upsert company
	result := ns.db.Where("company_id = ?", company.CompanyID).FirstOrCreate(company)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create/update company: %w", result.Error)
	}

	// Create token refresh record
	tokenRefresh := &models.TokenRefresh{
		CompanyID:   company.ID,
		LastRefresh: time.Now(),
		NextRefresh: authResp.ExpiresAt.Add(-24 * time.Hour), // Refresh 24 hours before expiry
		Status:      "active",
	}

	if err := ns.db.Create(tokenRefresh).Error; err != nil {
		return nil, fmt.Errorf("failed to create token refresh record: %w", err)
	}

	return company, nil
}

// GetLocations fetches all locations for a company
func (ns *NangoService) GetLocations(companyID string) ([]models.Location, error) {
	company := &models.Company{}
	if err := ns.db.Where("company_id = ?", companyID).First(company).Error; err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	// Fetch locations from Nango API
	locationsResp, err := ns.fetchLocationsFromAPI(company.AccessToken, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	var locations []models.Location
	for _, locResp := range locationsResp {
		location := models.Location{
			CompanyID:     company.ID,
			LocationID:    locResp.LocationID,
			LocationToken: locResp.LocationToken,
			BusinessName:  locResp.BusinessName,
			BusinessType:  locResp.BusinessType,
			Address:       locResp.Address,
			City:          locResp.City,
			State:         locResp.State,
			ZipCode:       locResp.ZipCode,
			Country:       locResp.Country,
			Phone:         locResp.Phone,
			Email:         locResp.Email,
			Website:       locResp.Website,
			IsActive:      true,
		}

		// Upsert location
		ns.db.Where("location_id = ?", location.LocationID).FirstOrCreate(&location)
		locations = append(locations, location)
	}

	return locations, nil
}

// RefreshToken refreshes the access token for a company
func (ns *NangoService) RefreshToken(companyID string) error {
	company := &models.Company{}
	if err := ns.db.Where("company_id = ?", companyID).First(company).Error; err != nil {
		return fmt.Errorf("company not found: %w", err)
	}

	// Refresh token via Nango API
	refreshResp, err := ns.refreshTokenAPI(company.RefreshToken, companyID)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update company with new tokens
	company.AccessToken = refreshResp.AccessToken
	company.RefreshToken = refreshResp.RefreshToken
	company.TokenExpiry = refreshResp.ExpiresAt

	if err := ns.db.Save(company).Error; err != nil {
		return fmt.Errorf("failed to update company tokens: %w", err)
	}

	// Update token refresh record
	tokenRefresh := &models.TokenRefresh{}
	if err := ns.db.Where("company_id = ?", company.ID).First(tokenRefresh).Error; err == nil {
		tokenRefresh.LastRefresh = time.Now()
		tokenRefresh.NextRefresh = refreshResp.ExpiresAt.Add(-24 * time.Hour)
		tokenRefresh.RefreshCount++
		tokenRefresh.Status = "active"
		tokenRefresh.ErrorMessage = ""
		ns.db.Save(tokenRefresh)
	}

	return nil
}

// Private helper methods

func (ns *NangoService) exchangeAuthCode(authCode string) (*NangoAuthResponse, error) {
	url := fmt.Sprintf("%s/oauth/token", ns.config.NangoServerURL)
	payload := map[string]string{
		"code":          authCode,
		"client_id":     ns.config.NangoPublicKey,
		"client_secret": ns.config.NangoSecretKey,
		"grant_type":    "authorization_code",
	}

	var result NangoAuthResponse
	err := ns.makeNangoRequest("POST", url, payload, "", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (ns *NangoService) fetchLocationsFromAPI(accessToken, companyID string) ([]NangoLocationResponse, error) {
	url := fmt.Sprintf("%s/api/v2/companies/%s/locations", ns.config.NangoServerURL, companyID)
	var result []NangoLocationResponse
	err := ns.makeNangoRequest("GET", url, nil, accessToken, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ns *NangoService) refreshTokenAPI(refreshToken, companyID string) (*NangoTokenRefreshResponse, error) {
	url := fmt.Sprintf("%s/oauth/refresh", ns.config.NangoServerURL)
	payload := NangoTokenRefreshRequest{
		RefreshToken: refreshToken,
		CompanyID:    companyID,
	}

	var result NangoTokenRefreshResponse
	err := ns.makeNangoRequest("POST", url, payload, "", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (ns *NangoService) makeNangoRequest(method, url string, payload interface{}, accessToken string, result interface{}) error {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}