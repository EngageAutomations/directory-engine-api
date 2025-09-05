package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"marketplace-app/internal/models"
)

type BusinessService struct {
	db    *gorm.DB
	nango *NangoService
	cache *CacheService
}

func NewBusinessService(db *gorm.DB, nango *NangoService, cache *CacheService) *BusinessService {
	return &BusinessService{
		db:    db,
		nango: nango,
		cache: cache,
	}
}

// GetCompanyByID retrieves a company by its ID
func (bs *BusinessService) GetCompanyByID(companyID string) (*models.Company, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("company:%s", companyID)
	if cached := bs.cache.Get(cacheKey); cached != nil {
		if company, ok := cached.(*models.Company); ok {
			return company, nil
		}
	}

	company := &models.Company{}
	err := bs.db.Where("company_id = ? AND is_active = ?", companyID, true).First(company).Error
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	// Cache the result
	bs.cache.Set(cacheKey, company, 30*time.Minute)
	return company, nil
}

// GetLocationsByCompany retrieves all locations for a company
func (bs *BusinessService) GetLocationsByCompany(companyID string) ([]models.Location, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("locations:%s", companyID)
	if cached := bs.cache.Get(cacheKey); cached != nil {
		if locations, ok := cached.([]models.Location); ok {
			return locations, nil
		}
	}

	company, err := bs.GetCompanyByID(companyID)
	if err != nil {
		return nil, err
	}

	var locations []models.Location
	err = bs.db.Where("company_id = ? AND is_active = ?", company.ID, true).Find(&locations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	// If no locations in DB, fetch from Nango
	if len(locations) == 0 {
		locations, err = bs.nango.GetLocations(companyID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch locations from Nango: %w", err)
		}
	}

	// Cache the result
	bs.cache.Set(cacheKey, locations, 15*time.Minute)
	return locations, nil
}

// GetLocationByID retrieves a specific location
func (bs *BusinessService) GetLocationByID(locationID string) (*models.Location, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("location:%s", locationID)
	if cached := bs.cache.Get(cacheKey); cached != nil {
		if location, ok := cached.(*models.Location); ok {
			return location, nil
		}
	}

	location := &models.Location{}
	err := bs.db.Where("location_id = ? AND is_active = ?", locationID, true).First(location).Error
	if err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	// Cache the result
	bs.cache.Set(cacheKey, location, 30*time.Minute)
	return location, nil
}

// GetContactsByLocation retrieves all contacts for a location
func (bs *BusinessService) GetContactsByLocation(locationID string) ([]models.Contact, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("contacts:%s", locationID)
	if cached := bs.cache.Get(cacheKey); cached != nil {
		if contacts, ok := cached.([]models.Contact); ok {
			return contacts, nil
		}
	}

	location, err := bs.GetLocationByID(locationID)
	if err != nil {
		return nil, err
	}

	var contacts []models.Contact
	err = bs.db.Where("location_id = ?", location.ID).Find(&contacts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contacts: %w", err)
	}

	// If no contacts in DB, fetch from external API and save
	if len(contacts) == 0 {
		contacts, err = bs.fetchAndSaveContacts(location)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch contacts from API: %w", err)
		}
	}

	// Cache the result
	bs.cache.Set(cacheKey, contacts, 15*time.Minute)
	return contacts, nil
}

// GetProductsByLocation retrieves all products for a location
func (bs *BusinessService) GetProductsByLocation(locationID string) ([]models.Product, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("products:%s", locationID)
	if cached := bs.cache.Get(cacheKey); cached != nil {
		if products, ok := cached.([]models.Product); ok {
			return products, nil
		}
	}

	location, err := bs.GetLocationByID(locationID)
	if err != nil {
		return nil, err
	}

	var products []models.Product
	err = bs.db.Where("location_id = ? AND is_active = ?", location.ID, true).Find(&products).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// If no products in DB, fetch from external API and save
	if len(products) == 0 {
		products, err = bs.fetchAndSaveProducts(location)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch products from API: %w", err)
		}
	}

	// Cache the result
	bs.cache.Set(cacheKey, products, 15*time.Minute)
	return products, nil
}

// CreateContact creates a new contact for a location
func (bs *BusinessService) CreateContact(locationID string, contact *models.Contact) error {
	location, err := bs.GetLocationByID(locationID)
	if err != nil {
		return err
	}

	contact.LocationID = location.ID

	if err := bs.db.Create(contact).Error; err != nil {
		return fmt.Errorf("failed to create contact: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("contacts:%s", locationID)
	bs.cache.Delete(cacheKey)

	return nil
}

// CreateProduct creates a new product for a location
func (bs *BusinessService) CreateProduct(locationID string, product *models.Product) error {
	location, err := bs.GetLocationByID(locationID)
	if err != nil {
		return err
	}

	product.LocationID = location.ID

	if err := bs.db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("products:%s", locationID)
	bs.cache.Delete(cacheKey)

	return nil
}

// UpdateLocation updates location information
func (bs *BusinessService) UpdateLocation(locationID string, updates map[string]interface{}) error {
	location, err := bs.GetLocationByID(locationID)
	if err != nil {
		return err
	}

	if err := bs.db.Model(location).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("location:%s", locationID)
	bs.cache.Delete(cacheKey)

	return nil
}

// SyncLocationData syncs location data with external API
func (bs *BusinessService) SyncLocationData(companyID string) error {
	// Fetch fresh data from Nango
	locations, err := bs.nango.GetLocations(companyID)
	if err != nil {
		return fmt.Errorf("failed to sync location data: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("locations:%s", companyID)
	bs.cache.Set(cacheKey, locations, 15*time.Minute)

	// Sync contacts and products for each location
	for _, location := range locations {
		go func(loc models.Location) {
			bs.fetchAndSaveContacts(&loc)
			bs.fetchAndSaveProducts(&loc)
		}(location)
	}

	return nil
}

// Private helper methods

func (bs *BusinessService) fetchAndSaveContacts(location *models.Location) ([]models.Contact, error) {
	// This would typically call an external API to fetch contacts
	// For now, we'll return an empty slice as this depends on the specific API
	var contacts []models.Contact

	// Example: fetch from external API using location token
	// contactsResp, err := bs.nango.fetchContactsFromAPI(location.LocationToken, location.LocationID)
	// if err != nil {
	//     return nil, err
	// }

	// Save contacts to database
	for _, contact := range contacts {
		contact.LocationID = location.ID
		bs.db.FirstOrCreate(&contact, "location_id = ? AND email = ?", contact.LocationID, contact.Email)
	}

	return contacts, nil
}

func (bs *BusinessService) fetchAndSaveProducts(location *models.Location) ([]models.Product, error) {
	// This would typically call an external API to fetch products
	// For now, we'll return an empty slice as this depends on the specific API
	var products []models.Product

	// Example: fetch from external API using location token
	// productsResp, err := bs.nango.fetchProductsFromAPI(location.LocationToken, location.LocationID)
	// if err != nil {
	//     return nil, err
	// }

	// Save products to database
	for _, product := range products {
		product.LocationID = location.ID
		bs.db.FirstOrCreate(&product, "location_id = ? AND sku = ?", product.LocationID, product.SKU)
	}

	return products, nil
}