package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"marketplace-app/internal/models"
)

// Initialize creates a database connection and runs migrations
func Initialize(databaseURL string) (*gorm.DB, error) {
	// Configure GORM logger
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *gorm.DB) error {
	// Enable UUID extension for PostgreSQL
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Could not create uuid-ossp extension: %v", err)
	}

	// Migrate tables individually in dependency order
	if err := db.AutoMigrate(&models.Company{}); err != nil {
		return fmt.Errorf("failed to migrate companies table: %w", err)
	}

	if err := db.AutoMigrate(&models.Location{}); err != nil {
		return fmt.Errorf("failed to migrate locations table: %w", err)
	}

	if err := db.AutoMigrate(&models.TokenRefresh{}); err != nil {
		return fmt.Errorf("failed to migrate token_refreshes table: %w", err)
	}

	// Now migrate tables with foreign keys
	if err := db.AutoMigrate(&models.Contact{}); err != nil {
		return fmt.Errorf("failed to migrate contacts table: %w", err)
	}

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		return fmt.Errorf("failed to migrate products table: %w", err)
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB) error {
	// Company indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_companies_company_id ON companies(company_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_companies_is_active ON companies(is_active)")

	// Location indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_locations_company_id ON locations(company_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_locations_location_id ON locations(location_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_locations_is_active ON locations(is_active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_locations_business_name ON locations(business_name)")

	// Contact indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_location_id ON contacts(location_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_contacts_is_primary ON contacts(is_primary)")

	// Product indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_products_location_id ON products(location_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_products_name ON products(name)")

	// Token refresh indexes
	db.Exec("CREATE INDEX IF NOT EXISTS idx_token_refresh_company_id ON token_refreshes(company_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_token_refresh_next_refresh ON token_refreshes(next_refresh)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_token_refresh_status ON token_refreshes(status)")

	return nil
}