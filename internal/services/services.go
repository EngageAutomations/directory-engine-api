package services

import (
	"gorm.io/gorm"
	"marketplace-app/internal/config"
)

// Services holds all application services
type Services struct {
	Nango     *NangoService
	Business  *BusinessService
	Token     *TokenService
	Cache     *CacheService
	Scheduler *SchedulerService
}

// NewServices creates and initializes all services
func NewServices(db *gorm.DB, cfg *config.Config) *Services {
	// Initialize cache service
	cacheService := NewCacheService(cfg)

	// Initialize core services
	nangoService := NewNangoService(db, cfg)
	businessService := NewBusinessService(db, nangoService, cacheService)
	tokenService := NewTokenService(db, nangoService)

	// Initialize scheduler service
	schedulerService := NewSchedulerService(tokenService)

	return &Services{
		Nango:     nangoService,
		Business:  businessService,
		Token:     tokenService,
		Cache:     cacheService,
		Scheduler: schedulerService,
	}
}

// Start initializes background services
func (s *Services) Start() error {
	// Start the token refresh scheduler
	return s.Scheduler.Start()
}

// Stop gracefully shuts down all services
func (s *Services) Stop() {
	if s.Scheduler != nil {
		s.Scheduler.Stop()
	}
	if s.Cache != nil {
		s.Cache.Close()
	}
}