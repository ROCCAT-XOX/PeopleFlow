package service

import (
	"PeopleFlow/backend/repository"
	"log"
	"sync"
	"time"
)

// Worker represents a background worker that performs periodic tasks
type Worker struct {
	isRunning      bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
	integrationRepo *repository.IntegrationRepository
}

// NewWorker creates a new background worker
func NewWorker() *Worker {
	return &Worker{
		isRunning:      false,
		stopChan:       make(chan struct{}),
		integrationRepo: repository.NewIntegrationRepository(),
	}
}

// Start begins the background processing
func (w *Worker) Start() {
	if w.isRunning {
		log.Println("Background worker is already running")
		return
	}

	w.isRunning = true
	w.stopChan = make(chan struct{})

	log.Println("Starting background worker")

	// Start the sync scheduler
	w.wg.Add(1)
	go w.runSyncScheduler()
}

// Stop terminates all background processing
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	log.Println("Stopping background worker")
	close(w.stopChan)
	w.wg.Wait()
	w.isRunning = false
}

// runSyncScheduler periodically synchronizes data from integrations
func (w *Worker) runSyncScheduler() {
	defer w.wg.Done()

	// Initial sync when starting up
	w.performSync()

	// Create a ticker that fires every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.performSync()
		case <-w.stopChan:
			log.Println("Sync scheduler stopped")
			return
		}
	}
}

// performSync synchronizes data from all active integrations
func (w *Worker) performSync() {
	log.Println("Performing scheduled sync of integrations")

	// Check if 123erfasst integration is active
	active, err := w.integrationRepo.GetIntegrationStatus("123erfasst")
	if err != nil || !active {
		log.Println("123erfasst integration is not active, skipping sync")
		return
	}

	// Initialize services
	erfasst123Service := service.NewErfasst123Service()

	// Get current year and month
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// Format date for sync
	startDate := getStartOfYear(currentYear)
	endDate := now.Format("2006-01-02")

	// Check for custom sync start date
	syncStartDate, err := w.integrationRepo.GetMetadata("123erfasst", "sync_start_date")
	if err == nil && syncStartDate != "" {
		// Use the configured start date if available
		startDate = syncStartDate
	}

	log.Printf("Syncing 123erfasst data from %s to %s", startDate, endDate)

	// Sync employees
	empCount, err := erfasst123Service.SyncErfasst123Employees()
	if err != nil {
		log.Printf("Error syncing 123erfasst employees: %v", err)
	} else {
		log.Printf("Synced %d employees from 123erfasst", empCount)
	}

	// Sync projects
	projCount, err := erfasst123Service.SyncErfasst123Projects(startDate, endDate)
	if err != nil {
		log.Printf("Error syncing 123erfasst projects: %v", err)
	} else {
		log.Printf("Synced %d project assignments from 123erfasst", projCount)
	}

	// Sync time entries
	timeCount, err := erfasst123Service.SyncErfasst123TimeEntries(startDate, endDate)
	if err != nil {
		log.Printf("Error syncing 123erfasst time entries: %v", err)
	} else {
		log.Printf("Synced %d time entries from 123erfasst", timeCount)
	}