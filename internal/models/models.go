package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Company represents a marketplace company
type Company struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID   string    `gorm:"uniqueIndex;not null" json:"company_id"`
	CompanyName string    `gorm:"not null" json:"company_name"`
	AccessToken string    `gorm:"not null" json:"-"` // Hidden from JSON
	RefreshToken string   `gorm:"not null" json:"-"` // Hidden from JSON
	TokenExpiry time.Time `json:"token_expiry"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Locations []Location `gorm:"foreignKey:CompanyID;references:ID" json:"locations,omitempty"`
}

// Location represents a business location
type Location struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID        uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	LocationID       string    `gorm:"uniqueIndex;not null" json:"location_id"`
	LocationToken    string    `gorm:"not null" json:"-"` // Hidden from JSON
	BusinessName     string    `gorm:"not null" json:"business_name"`
	BusinessType     string    `json:"business_type"`
	Address          string    `json:"address"`
	City             string    `json:"city"`
	State            string    `json:"state"`
	ZipCode          string    `json:"zip_code"`
	Country          string    `json:"country"`
	Phone            string    `json:"phone"`
	Email            string    `json:"email"`
	Website          string    `json:"website"`
	IsActive         bool      `gorm:"default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Company  Company   `gorm:"foreignKey:CompanyID;references:ID" json:"company,omitempty"`
	Contacts []Contact `gorm:"foreignKey:LocationID;references:ID" json:"contacts,omitempty"`
	Products []Product `gorm:"foreignKey:LocationID;references:ID" json:"products,omitempty"`
}

// Contact represents business contact information
type Contact struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LocationID uuid.UUID `gorm:"type:uuid;not null;index" json:"location_id"`
	FirstName  string    `gorm:"not null" json:"first_name"`
	LastName   string    `gorm:"not null" json:"last_name"`
	Title      string    `json:"title"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Mobile     string    `json:"mobile"`
	IsPrimary  bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Location Location `gorm:"foreignKey:LocationID;references:ID" json:"location,omitempty"`
}

// Product represents products/services offered by a business
type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LocationID  uuid.UUID `gorm:"type:uuid;not null;index" json:"location_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       float64   `json:"price"`
	Currency    string    `gorm:"default:USD" json:"currency"`
	SKU         string    `json:"sku"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Location Location `gorm:"foreignKey:LocationID;references:ID" json:"location,omitempty"`
}

// TokenRefresh represents token refresh tracking
type TokenRefresh struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID    uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	LastRefresh  time.Time `json:"last_refresh"`
	NextRefresh  time.Time `json:"next_refresh"`
	RefreshCount int       `gorm:"default:0" json:"refresh_count"`
	Status       string    `gorm:"default:active" json:"status"` // active, failed, expired
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Company Company `gorm:"foreignKey:CompanyID;references:ID" json:"company,omitempty"`
}