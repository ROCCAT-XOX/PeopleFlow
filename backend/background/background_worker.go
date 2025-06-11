// backend/background/worker.go
package background

import (
	"log"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
)

// Worker repräsentiert einen Hintergrundprozess für regelmäßige Aufgaben
type Worker struct {
	stopChan chan struct{}
	running  bool
	lastEmailReport time.Time
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

	// Timer für wöchentliche E-Mail-Berichte (jeden Freitag um 17:00)
	emailTicker := time.NewTicker(1 * time.Hour)
	defer emailTicker.Stop()

	// Sofort erste Synchronisierung durchführen
	w.performSynchronization()

	for {
		select {
		case <-ticker.C:
			// Synchronisierungsaufgaben ausführen
			w.performSynchronization()
		case <-emailTicker.C:
			// Prüfen, ob wöchentliche E-Mail-Berichte gesendet werden sollen
			w.checkWeeklyEmailReports()
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

// checkWeeklyEmailReports prüft, ob wöchentliche E-Mail-Berichte gesendet werden sollen
func (w *Worker) checkWeeklyEmailReports() {
	now := time.Now()
	
	// Nur Freitags um 17:00 Uhr
	if now.Weekday() != time.Friday || now.Hour() != 17 {
		return
	}
	
	// Prüfen, ob bereits heute ein Bericht gesendet wurde
	if w.lastEmailReport.Year() == now.Year() && 
	   w.lastEmailReport.YearDay() == now.YearDay() {
		return
	}
	
	log.Println("Sending weekly email reports...")
	
	// E-Mail-Service erstellen
	emailService := service.NewEmailService()
	if !emailService.IsEmailConfigured() {
		log.Println("Email service not configured, skipping weekly reports")
		return
	}
	
	// Mitarbeiter-Repository
	employeeRepo := repository.NewEmployeeRepository()
	employees, _, err := employeeRepo.FindAll(0, 1000, "lastName", 1) // Get up to 1000 employees
	if err != nil {
		log.Printf("Error getting employees for weekly reports: %v", err)
		return
	}
	
	// Aktivitäts-Repository
	activityRepo := repository.NewActivityRepository()
	
	// Wochenstart und -ende berechnen (Montag bis Sonntag)
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1) // Letzter Montag
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 6) // Sonntag
	weekEnd = time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 23, 59, 59, 0, weekEnd.Location())
	
	sentCount := 0
	errorCount := 0
	
	// Für jeden Mitarbeiter einen Bericht senden
	for _, employee := range employees {
		if employee.Email == "" {
			log.Printf("Skipping employee %s %s - no email address", employee.FirstName, employee.LastName)
			continue
		}
		
		// Aktivitäten der Woche für diesen Mitarbeiter abrufen
		activities, err := activityRepo.GetActivitiesForEmployeeInDateRange(employee.ID, weekStart, weekEnd)
		if err != nil {
			log.Printf("Error getting activities for employee %s %s: %v", employee.FirstName, employee.LastName, err)
			errorCount++
			continue
		}
		
		// Gesamtarbeitszeit berechnen (vereinfacht - basierend auf Aktivitäten)
		totalHours := w.calculateTotalHours(activities)
		
		// E-Mail senden
		err = emailService.SendWeeklyReport(employee, weekStart, weekEnd, totalHours, activities)
		if err != nil {
			log.Printf("Error sending weekly report to %s %s (%s): %v", 
				employee.FirstName, employee.LastName, employee.Email, err)
			errorCount++
		} else {
			log.Printf("Weekly report sent to %s %s (%s)", 
				employee.FirstName, employee.LastName, employee.Email)
			sentCount++
		}
	}
	
	// Letzten E-Mail-Bericht-Zeitpunkt aktualisieren
	w.lastEmailReport = now
	
	log.Printf("Weekly email reports completed: %d sent, %d errors", sentCount, errorCount)
}

// calculateTotalHours berechnet die Gesamtarbeitszeit basierend auf Aktivitäten
func (w *Worker) calculateTotalHours(activities []model.Activity) float64 {
	// Vereinfachte Berechnung - zählt Arbeitstage * 8 Stunden
	// In einer realen Implementierung würde man echte Zeiterfassungsdaten verwenden
	workDays := 0
	
	// Eindeutige Arbeitstage zählen
	workDayMap := make(map[string]bool)
	for _, activity := range activities {
		// Verwende alle Aktivitäten als Arbeitsnachweis (da ActivityTypeTimeEntry nicht existiert)
		dateKey := activity.Timestamp.Format("2006-01-02")
		workDayMap[dateKey] = true
	}
	
	workDays = len(workDayMap)
	
	// Wenn keine Zeiteinträge vorhanden sind, Standard-Arbeitswoche annehmen (5 Tage)
	if workDays == 0 {
		workDays = 5
	}
	
	return float64(workDays) * 8.0 // 8 Stunden pro Arbeitstag
}
