package services

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	cron        *cron.Cron
	tokenService *TokenService
	isRunning   bool
}

func NewSchedulerService(tokenService *TokenService) *SchedulerService {
	// Create cron with seconds precision and logging
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(cron.VerbosePrintfLogger(log.New(log.Writer(), "CRON: ", log.LstdFlags))),
	)

	return &SchedulerService{
		cron:         c,
		tokenService: tokenService,
		isRunning:    false,
	}
}

// Start initializes and starts all scheduled jobs
func (ss *SchedulerService) Start() error {
	if ss.isRunning {
		return nil
	}

	// Schedule token refresh job - runs every hour
	_, err := ss.cron.AddFunc("0 0 * * * *", func() {
		log.Println("Running scheduled token refresh job")
		if err := ss.tokenService.RefreshExpiredTokens(); err != nil {
			log.Printf("Token refresh job failed: %v", err)
		}
	})
	if err != nil {
		return err
	}

	// Schedule token cleanup job - runs daily at 2 AM
	_, err = ss.cron.AddFunc("0 0 2 * * *", func() {
		log.Println("Running scheduled token cleanup job")
		if err := ss.tokenService.CleanupExpiredTokens(); err != nil {
			log.Printf("Token cleanup job failed: %v", err)
		}
	})
	if err != nil {
		return err
	}

	// Schedule health check job - runs every 5 minutes
	_, err = ss.cron.AddFunc("0 */5 * * * *", func() {
		ss.performHealthCheck()
	})
	if err != nil {
		return err
	}

	// Schedule token status monitoring - runs every 30 minutes
	_, err = ss.cron.AddFunc("0 */30 * * * *", func() {
		ss.monitorTokenStatuses()
	})
	if err != nil {
		return err
	}

	// Start the cron scheduler
	ss.cron.Start()
	ss.isRunning = true

	log.Println("Scheduler service started successfully")
	return nil
}

// Stop gracefully stops the scheduler
func (ss *SchedulerService) Stop() {
	if !ss.isRunning {
		return
	}

	log.Println("Stopping scheduler service...")
	ctx := ss.cron.Stop()
	
	// Wait for running jobs to complete with timeout
	select {
	case <-ctx.Done():
		log.Println("All scheduled jobs completed")
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for jobs to complete")
	}

	ss.isRunning = false
	log.Println("Scheduler service stopped")
}

// AddJob adds a custom job to the scheduler
func (ss *SchedulerService) AddJob(spec string, cmd func()) (cron.EntryID, error) {
	return ss.cron.AddFunc(spec, cmd)
}

// RemoveJob removes a job from the scheduler
func (ss *SchedulerService) RemoveJob(id cron.EntryID) {
	ss.cron.Remove(id)
}

// GetJobEntries returns all scheduled job entries
func (ss *SchedulerService) GetJobEntries() []cron.Entry {
	return ss.cron.Entries()
}

// IsRunning returns whether the scheduler is currently running
func (ss *SchedulerService) IsRunning() bool {
	return ss.isRunning
}

// RunTokenRefreshNow manually triggers token refresh job
func (ss *SchedulerService) RunTokenRefreshNow() error {
	log.Println("Manually triggering token refresh job")
	return ss.tokenService.RefreshExpiredTokens()
}

// RunCleanupNow manually triggers cleanup job
func (ss *SchedulerService) RunCleanupNow() error {
	log.Println("Manually triggering token cleanup job")
	return ss.tokenService.CleanupExpiredTokens()
}

// Private helper methods

func (ss *SchedulerService) performHealthCheck() {
	// Get all token statuses
	statuses, err := ss.tokenService.GetAllTokenStatuses()
	if err != nil {
		log.Printf("Health check failed to get token statuses: %v", err)
		return
	}

	expiredCount := 0
	needsRefreshCount := 0
	activeCount := 0

	for _, status := range statuses {
		if status.IsExpired {
			expiredCount++
		} else if status.NeedsRefresh {
			needsRefreshCount++
		} else {
			activeCount++
		}
	}

	log.Printf("Health Check - Active: %d, Needs Refresh: %d, Expired: %d", 
		activeCount, needsRefreshCount, expiredCount)

	// Alert if too many tokens are expired
	if expiredCount > 0 {
		log.Printf("WARNING: %d companies have expired tokens", expiredCount)
	}

	// Alert if too many tokens need refresh
	if needsRefreshCount > 5 {
		log.Printf("INFO: %d companies need token refresh soon", needsRefreshCount)
	}
}

func (ss *SchedulerService) monitorTokenStatuses() {
	log.Println("Running token status monitoring")

	// Get companies with tokens expiring in the next 48 hours
	statuses, err := ss.tokenService.GetAllTokenStatuses()
	if err != nil {
		log.Printf("Token monitoring failed: %v", err)
		return
	}

	criticalCount := 0
	warningCount := 0

	for _, status := range statuses {
		if status.IsExpired {
			criticalCount++
			log.Printf("CRITICAL: Token expired for company %s (%s)", 
				status.CompanyID, status.CompanyName)
		} else if status.TimeToExpiry < 24*time.Hour {
			criticalCount++
			log.Printf("CRITICAL: Token expires in %v for company %s (%s)", 
				status.TimeToExpiry, status.CompanyID, status.CompanyName)
		} else if status.TimeToExpiry < 48*time.Hour {
			warningCount++
			log.Printf("WARNING: Token expires in %v for company %s (%s)", 
				status.TimeToExpiry, status.CompanyID, status.CompanyName)
		}
	}

	if criticalCount > 0 || warningCount > 0 {
		log.Printf("Token Status Summary - Critical: %d, Warning: %d", 
			criticalCount, warningCount)
	}
}

// SchedulerStats provides statistics about the scheduler
type SchedulerStats struct {
	IsRunning    bool `json:"is_running"`
	JobCount     int  `json:"job_count"`
	NextJobTime  *time.Time `json:"next_job_time,omitempty"`
	LastJobTime  *time.Time `json:"last_job_time,omitempty"`
}

// GetStats returns scheduler statistics
func (ss *SchedulerService) GetStats() SchedulerStats {
	stats := SchedulerStats{
		IsRunning: ss.isRunning,
		JobCount:  len(ss.cron.Entries()),
	}

	// Find next and last job times
	entries := ss.cron.Entries()
	if len(entries) > 0 {
		// Find the earliest next run time
		nextTime := entries[0].Next
		for _, entry := range entries[1:] {
			if entry.Next.Before(nextTime) {
				nextTime = entry.Next
			}
		}
		stats.NextJobTime = &nextTime

		// Find the latest previous run time
		if !entries[0].Prev.IsZero() {
			lastTime := entries[0].Prev
			for _, entry := range entries[1:] {
				if entry.Prev.After(lastTime) {
					lastTime = entry.Prev
				}
			}
			stats.LastJobTime = &lastTime
		}
	}

	return stats
}