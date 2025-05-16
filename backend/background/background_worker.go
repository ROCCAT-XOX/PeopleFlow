// backend/background/worker.go
package background

import (
	"log"
	"time"

	"PeopleFlow/backend/service"
)

// Worker repräsentiert einen Hintergrundprozess für regelmäßige Aufgaben
type Worker struct {
	stopChan chan struct{}
	running  bool
}

// NewWorker erstellt einen neuen Worker
func NewWorker() *Worker {
	return &Worker{
		stopChan: make(chan struct{}),
		running:  false,
	}
}

// Start startet den Worker
func (w *Worker) Start() {
	if w.running {
		return
	}

	w.running = true
	go w.run()
}

// Stop stoppt den Worker
func (w *Worker) Stop() {
	if !w.running {
		return
	}

	w.running = false
	w.stopChan <- struct{}{}
}

// run führt die regelmäßigen Aufgaben aus
func (w *Worker) run() {
	// Timer für regelmäßige Aufgaben (alle 5 Minuten)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Sofort erste Synchronisierung durchführen
	w.performSynchronization()

	for {
		select {
		case <-ticker.C:
			// Synchronisierungsaufgaben ausführen
			w.performSynchronization()
		case <-w.stopChan:
			log.Println("Background worker stopped")
			return
		}
	}
}

// performSynchronization führt Synchronisierungsaufgaben aus
func (w *Worker) performSynchronization() {
	log.Println("Performing background synchronization tasks...")

	// Timebutler-Synchronisierung
	timebutlerService := service.NewTimebutlerService()
	if connected := timebutlerService.IsConnected(); connected {
		log.Println("Synchronizing Timebutler data...")

		// Benutzer synchronisieren
		if count, err := timebutlerService.SyncTimebutlerUsers(); err != nil {
			log.Printf("Error synchronizing Timebutler users: %v", err)
		} else {
			log.Printf("Synchronized %d Timebutler users", count)
		}

		// Aktuelle Urlaubsansprüche synchronisieren
		currentYear := time.Now().Format("2006")
		if count, err := timebutlerService.SyncHolidayEntitlements(currentYear); err != nil {
			log.Printf("Error synchronizing Timebutler holiday entitlements: %v", err)
		} else {
			log.Printf("Synchronized %d Timebutler holiday entitlements", count)
		}

		// Abwesenheiten synchronisieren
		if count, err := timebutlerService.SyncTimebutlerAbsences(currentYear); err != nil {
			log.Printf("Error synchronizing Timebutler absences: %v", err)
		} else {
			log.Printf("Synchronized %d Timebutler absences", count)
		}
	}

	// 123erfasst-Synchronisierung
	erfasst123Service := service.NewErfasst123Service()
	if connected := erfasst123Service.IsConnected(); connected {
		// Nur synchronisieren, wenn Auto-Sync aktiviert ist
		autoSync, err := erfasst123Service.IsAutoSyncEnabled()
		if err != nil || !autoSync {
			if err != nil {
				log.Printf("Error checking 123erfasst auto-sync setting: %v", err)
			}
			return
		}

		log.Println("Synchronizing 123erfasst data...")

		// Startdatum für die Synchronisierung abrufen
		syncStartDate, err := erfasst123Service.GetSyncStartDate()
		if err != nil {
			log.Printf("Error getting 123erfasst sync start date: %v", err)
			return
		}

		// Aktuelles Datum für das Ende der Synchronisierung
		now := time.Now()
		endDate := now.Format("2006-01-02")

		// Mitarbeiter synchronisieren
		if count, err := erfasst123Service.SyncErfasst123Employees(); err != nil {
			log.Printf("Error synchronizing 123erfasst employees: %v", err)
		} else {
			log.Printf("Synchronized %d 123erfasst employees", count)
		}

		// Projekte synchronisieren
		if count, err := erfasst123Service.SyncErfasst123Projects(syncStartDate, endDate); err != nil {
			log.Printf("Error synchronizing 123erfasst projects: %v", err)
		} else {
			log.Printf("Synchronized %d 123erfasst projects", count)
		}

		// Zeiteinträge synchronisieren
		if count, err := erfasst123Service.SyncErfasst123TimeEntries(syncStartDate, endDate); err != nil {
			log.Printf("Error synchronizing 123erfasst time entries: %v", err)
		} else {
			log.Printf("Synchronized %d 123erfasst time entries", count)
		}
	}

	log.Println("Background synchronization tasks completed")
}
