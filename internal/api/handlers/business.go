package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"marketplace-app/internal/models"
	"marketplace-app/internal/services"
)

type BusinessHandler struct {
	services *services.Services
}

func NewBusinessHandler(services *services.Services) *BusinessHandler {
	return &BusinessHandler{
		services: services,
	}
}

// Company Handlers

// GetCompany retrieves a company by ID
func (h *BusinessHandler) GetCompany(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	company, err := h.services.Business.GetCompanyByID(companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"company": company,
	})
}

// GetCompanies retrieves all companies with pagination
func (h *BusinessHandler) GetCompanies(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// TODO: Implement GetCompanies method or use direct database query
	companies := []interface{}{} // Placeholder
	total := int64(0)
	err := error(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve companies",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": companies,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateCompany creates a new company
func (h *BusinessHandler) CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	company.CreatedAt = now
	company.UpdatedAt = now

	// TODO: Implement CreateCompany method
	err := fmt.Errorf("CreateCompany method not implemented")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create company",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Company created successfully",
		"company": company,
	})
}

// UpdateCompany updates an existing company
func (h *BusinessHandler) UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	var updates models.Company
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement UpdateCompany method
	err := fmt.Errorf("UpdateCompany method not implemented")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update company",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Company updated successfully",
		"company": updates,
	})
}

// SyncCompanyData syncs company data from Nango
func (h *BusinessHandler) SyncCompanyData(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	err := h.services.Business.SyncLocationData(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync company data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Company data synced successfully",
		"company_id": companyID,
		"synced_at": time.Now().Unix(),
	})
}

// Location Handlers

// GetLocations retrieves locations for a company
func (h *BusinessHandler) GetLocations(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	locations, err := h.services.Business.GetLocationsByCompany(companyID)
	total := int64(len(locations))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve locations",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"locations": locations,
		"company_id": companyID,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetLocation retrieves a specific location
func (h *BusinessHandler) GetLocation(c *gin.Context) {
	locationID := c.Param("id")
	if locationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location ID is required"})
		return
	}

	location, err := h.services.Business.GetLocationByID(locationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Location not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"location": location,
	})
}

// CreateLocation creates a new location
func (h *BusinessHandler) CreateLocation(c *gin.Context) {
	var location models.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	location.CreatedAt = now
	location.UpdatedAt = now

	// TODO: Implement CreateLocation method
	err := fmt.Errorf("CreateLocation method not implemented")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create location",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Location created successfully",
		"location": location,
	})
}

// Contact Handlers

// GetContacts retrieves contacts for a company
func (h *BusinessHandler) GetContacts(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// TODO: Implement GetContactsByCompany method
	contacts := []interface{}{}
	total := int64(0)
	err := error(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve contacts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contacts": contacts,
		"company_id": companyID,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateContact creates a new contact
func (h *BusinessHandler) CreateContact(c *gin.Context) {
	var contact models.Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	contact.CreatedAt = now
	contact.UpdatedAt = now

	locationID := c.Param("location_id")
	err := h.services.Business.CreateContact(locationID, &contact)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Contact created successfully",
		"contact": contact,
	})
}

// Product Handlers

// GetProducts retrieves products for a company
func (h *BusinessHandler) GetProducts(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Optional location filter
	locationID := c.Query("location_id")

	// TODO: Implement GetProductsByCompany method
	products := []interface{}{}
	total := int64(0)
	err := error(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve products",
			"details": err.Error(),
		})
		return
	}

	response := gin.H{
		"products": products,
		"company_id": companyID,
		"pagination": gin.H{
			"page": page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	if locationID != "" {
		response["location_id"] = locationID
	}

	c.JSON(http.StatusOK, response)
}

// CreateProduct creates a new product
func (h *BusinessHandler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	locationID := c.Param("location_id")
	err := h.services.Business.CreateProduct(locationID, &product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product,
	})
}

// GetBusinessSummary provides a summary of business data
func (h *BusinessHandler) GetBusinessSummary(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Get company
	company, err := h.services.Business.GetCompanyByID(companyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
			"details": err.Error(),
		})
		return
	}

	// Get counts
	locationsCount := int64(0) // TODO: Implement GetLocationCount method
	contactsCount := int64(0) // TODO: Implement GetContactCount method
	productsCount := int64(0) // TODO: Implement GetProductCount method

	// Get token status
	isTokenValid, _ := h.services.Token.ValidateToken(companyID)
	tokenExpiry, _ := h.services.Token.GetTokenExpiryInfo(companyID)

	c.JSON(http.StatusOK, gin.H{
		"company": company,
		"summary": gin.H{
			"locations_count": locationsCount,
			"contacts_count": contactsCount,
			"products_count": productsCount,
			"token_valid": isTokenValid,
			"token_expires_at": tokenExpiry.TokenExpiry.Unix(),
			"last_sync": company.UpdatedAt,
		},
	})
}