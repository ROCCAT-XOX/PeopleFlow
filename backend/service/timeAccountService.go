// backend/service/timeAccountService.go
package service

import (
	"fmt"
	"sort"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
)

// EmployeeOvertimeSummary repräsentiert eine Überstunden-Zusammenfassung für einen Mitarbeiter
type EmployeeOvertimeSummary struct {
	EmployeeID         string                  `json:"employeeId"`
	EmployeeName       string                  `json:"employeeName"`
	WeeklyTargetHours  float64                 `json:"weeklyTargetHours"`
	CurrentBalance     float64                 `json:"currentBalance"`
	LastCalculated     time.Time               `json:"lastCalculated"`
	WeeklyEntries      []model.WeeklyTimeEntry `json:"weeklyEntries"`
	TotalWorkedHours   float64                 `json:"totalWorkedHours"`
	TotalPlannedHours  float64                 `json:"totalPlannedHours"`
	AverageWeeklyHours float64                 `json:"averageWeeklyHours"`
}

// TimeAccountService verwaltet Zeitkonten und Überstunden-Berechnungen
type TimeAccountService struct {
	employeeRepo   *repository.EmployeeRepository
	holidayService *HolidayService
	settingsRepo   *repository.SystemSettingsRepository
}

// NewTimeAccountService erstellt einen neuen TimeAccountService
func NewTimeAccountService() *TimeAccountService {
	return &TimeAccountService{
		employeeRepo:   repository.NewEmployeeRepository(),
		holidayService: NewHolidayService(),
		settingsRepo:   repository.NewSystemSettingsRepository(),
	}
}

// CalculateOvertimeForEmployee berechnet Überstunden für einen einzelnen Mitarbeiter
func (s *TimeAccountService) CalculateOvertimeForEmployee(employee *model.Employee) error {
	if len(employee.TimeEntries) == 0 {
		// Keine Zeiteinträge vorhanden
		employee.OvertimeBalance = 0
		employee.LastTimeCalculated = time.Now()
		return s.employeeRepo.Update(employee)
	}

	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	var totalOvertime float64
	var weeklyEntries []model.WeeklyTimeEntry

	// Gruppiere Zeiteinträge nach Wochen
	weeklyData := s.groupTimeEntriesByWeek(employee.TimeEntries)

	// Sortiere Wochen chronologisch
	var weeks []time.Time
	for week := range weeklyData {
		weeks = append(weeks, week)
	}
	sort.Slice(weeks, func(i, j int) bool {
		return weeks[i].Before(weeks[j])
	})

	for _, weekStart := range weeks {
		entries := weeklyData[weekStart]

		// Geplante Stunden für diese Woche (unter Berücksichtigung von Feiertagen)
		plannedHours := s.CalculateTargetHoursForWeek(employee, weekStart, state)

		// Tatsächlich gearbeitete Stunden
		var actualHours float64
		daysWorked := make(map[string]bool)

		for _, entry := range entries {
			actualHours += entry.Duration
			daysWorked[entry.Date.Format("2006-01-02")] = true
		}

		// Überstunden für diese Woche
		overtimeHours := actualHours - plannedHours
		totalOvertime += overtimeHours

		// Wöchentlichen Eintrag erstellen
		weekEnd := weekStart.AddDate(0, 0, 6)
		year, week := weekStart.ISOWeek()

		weeklyEntry := model.WeeklyTimeEntry{
			WeekStartDate: weekStart,
			WeekEndDate:   weekEnd,
			Year:          year,
			WeekNumber:    week,
			PlannedHours:  plannedHours,
			ActualHours:   actualHours,
			OvertimeHours: overtimeHours,
			DaysWorked:    len(daysWorked),
			IsComplete:    true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		weeklyEntries = append(weeklyEntries, weeklyEntry)
	}

	// Mitarbeiter aktualisieren
	employee.OvertimeBalance = totalOvertime
	employee.WeeklyTimeEntries = weeklyEntries
	employee.LastTimeCalculated = time.Now()

	return s.employeeRepo.Update(employee)
}

// CalculateTargetHoursForWeek berechnet die Soll-Arbeitszeit für eine Woche
// unter Berücksichtigung von Feiertagen
func (s *TimeAccountService) CalculateTargetHoursForWeek(employee *model.Employee, weekStart time.Time, state model.GermanState) float64 {
	// Grundlegende Wochenstunden
	weeklyHours := employee.GetWeeklyTargetHours()
	if weeklyHours == 0 {
		return 40.0 // Standard-Vollzeit als Fallback
	}

	dailyHours := employee.GetWorkingHoursPerDay()
	if dailyHours == 0 {
		dailyHours = 8.0 // Standard
	}

	// Zähle Arbeitstage in dieser Woche (Mo-Fr)
	weekEnd := weekStart.AddDate(0, 0, 6) // Sonntag
	workingDaysInWeek := 0
	holidaysInWeek := 0

	for d := weekStart; d.Before(weekEnd.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		// Nur Montag bis Freitag zählen
		if d.Weekday() >= time.Monday && d.Weekday() <= time.Friday {
			if s.holidayService.IsHoliday(d, state) {
				holidaysInWeek++
			} else {
				workingDaysInWeek++
			}
		}
	}

	// Bei Teilzeit: Proportional reduzieren
	if employee.WorkingDaysPerWeek > 0 && employee.WorkingDaysPerWeek < 5 {
		// Berechne die tatsächlichen Arbeitstage basierend auf dem Teilzeit-Modell
		actualWorkingDays := float64(workingDaysInWeek) * (float64(employee.WorkingDaysPerWeek) / 5.0)
		return actualWorkingDays * dailyHours
	}

	// Vollzeit: Reduziere um Feiertage
	return float64(workingDaysInWeek) * dailyHours
}

// CalculateTargetHoursForMonth berechnet die Soll-Arbeitszeit für einen Monat
func (s *TimeAccountService) CalculateTargetHoursForMonth(employee *model.Employee, year int, month time.Month) float64 {
	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	// Arbeitstage im Monat berechnen (ohne Wochenenden und Feiertage)
	workingDays := s.holidayService.GetWorkingDaysInMonth(year, month, state)

	// Bei Teilzeit: Proportional reduzieren
	dailyHours := employee.GetWorkingHoursPerDay()
	if dailyHours == 0 {
		dailyHours = 8.0 // Standard
	}

	if employee.WorkingDaysPerWeek > 0 && employee.WorkingDaysPerWeek < 5 {
		// Reduziere die Arbeitstage entsprechend dem Teilzeit-Modell
		workingDaysAdjusted := float64(workingDays) * (float64(employee.WorkingDaysPerWeek) / 5.0)
		return workingDaysAdjusted * dailyHours
	}

	return float64(workingDays) * dailyHours
}

// RecalculateAllEmployeeOvertimes berechnet Überstunden für alle Mitarbeiter neu
func (s *TimeAccountService) RecalculateAllEmployeeOvertimes() error {
	employees, err := s.employeeRepo.FindAll()
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der Mitarbeiter: %w", err)
	}

	var errors []string
	successCount := 0

	for _, employee := range employees {
		// Nur Mitarbeiter mit Zeiteinträgen bearbeiten
		if len(employee.TimeEntries) == 0 {
			continue
		}

		err := s.CalculateOvertimeForEmployee(employee)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Fehler bei %s %s: %v", employee.FirstName, employee.LastName, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("fehler bei %d Mitarbeitern: %v", len(errors), errors)
	}

	return nil
}

// GetEmployeeOvertimeSummary erstellt eine detaillierte Überstunden-Zusammenfassung für einen Mitarbeiter
func (s *TimeAccountService) GetEmployeeOvertimeSummary(employeeID string) (*EmployeeOvertimeSummary, error) {
	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("mitarbeiter nicht gefunden: %w", err)
	}

	// Berechne Statistiken
	var totalWorkedHours float64
	var totalPlannedHours float64
	var averageWeeklyHours float64

	for _, entry := range employee.TimeEntries {
		totalWorkedHours += entry.Duration
	}

	for _, weekEntry := range employee.WeeklyTimeEntries {
		totalPlannedHours += weekEntry.PlannedHours
	}

	if len(employee.WeeklyTimeEntries) > 0 {
		averageWeeklyHours = totalWorkedHours / float64(len(employee.WeeklyTimeEntries))
	}

	summary := &EmployeeOvertimeSummary{
		EmployeeID:         employee.ID.Hex(),
		EmployeeName:       employee.FirstName + " " + employee.LastName,
		WeeklyTargetHours:  employee.GetWeeklyTargetHours(),
		CurrentBalance:     employee.OvertimeBalance,
		LastCalculated:     employee.LastTimeCalculated,
		WeeklyEntries:      employee.WeeklyTimeEntries,
		TotalWorkedHours:   totalWorkedHours,
		TotalPlannedHours:  totalPlannedHours,
		AverageWeeklyHours: averageWeeklyHours,
	}

	return summary, nil
}

// IsWorkingDay prüft, ob ein Tag ein Arbeitstag ist (kein Wochenende, kein Feiertag)
func (s *TimeAccountService) IsWorkingDay(date time.Time, state model.GermanState) bool {
	// Wochenende
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		return false
	}

	// Feiertag
	if s.holidayService.IsHoliday(date, state) {
		return false
	}

	return true
}

// GetHolidaysInPeriod gibt alle Feiertage in einem Zeitraum zurück
func (s *TimeAccountService) GetHolidaysInPeriod(startDate, endDate time.Time) ([]Holiday, error) {
	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	if err != nil {
		return nil, err
	}

	state := model.GermanState(settings.State)
	var holidays []Holiday

	// Für jedes Jahr im Zeitraum Feiertage sammeln
	startYear := startDate.Year()
	endYear := endDate.Year()

	for year := startYear; year <= endYear; year++ {
		yearHolidays := s.holidayService.GetHolidaysForState(year, state)
		for _, holiday := range yearHolidays {
			// Nur Feiertage im gewünschten Zeitraum
			if !holiday.Date.Before(startDate) && !holiday.Date.After(endDate) {
				holidays = append(holidays, holiday)
			}
		}
	}

	return holidays, nil
}

// groupTimeEntriesByWeek gruppiert Zeiteinträge nach Wochen (Montag als Wochenstart)
func (s *TimeAccountService) groupTimeEntriesByWeek(timeEntries []model.TimeEntry) map[time.Time][]model.TimeEntry {
	weeklyData := make(map[time.Time][]model.TimeEntry)

	for _, entry := range timeEntries {
		// Finde den Montag dieser Woche
		weekStart := entry.Date
		for weekStart.Weekday() != time.Monday {
			weekStart = weekStart.AddDate(0, 0, -1)
		}

		weeklyData[weekStart] = append(weeklyData[weekStart], entry)
	}

	return weeklyData
}

// CalculateOvertimeForPeriod berechnet Überstunden für einen bestimmten Zeitraum
func (s *TimeAccountService) CalculateOvertimeForPeriod(employee *model.Employee, startDate, endDate time.Time) (float64, error) {
	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	// Filtere Zeiteinträge für den gewünschten Zeitraum
	var periodEntries []model.TimeEntry
	for _, entry := range employee.TimeEntries {
		if !entry.Date.Before(startDate) && !entry.Date.After(endDate) {
			periodEntries = append(periodEntries, entry)
		}
	}

	if len(periodEntries) == 0 {
		return 0, nil
	}

	// Gruppiere nach Wochen
	weeklyData := s.groupTimeEntriesByWeek(periodEntries)
	var totalOvertime float64

	for weekStart, entries := range weeklyData {
		// Geplante Stunden für diese Woche
		plannedHours := s.CalculateTargetHoursForWeek(employee, weekStart, state)

		// Tatsächlich gearbeitete Stunden
		var actualHours float64
		for _, entry := range entries {
			actualHours += entry.Duration
		}

		// Überstunden für diese Woche
		overtimeHours := actualHours - plannedHours
		totalOvertime += overtimeHours
	}

	return totalOvertime, nil
}

// GetWorkingDaysInPeriod berechnet die Anzahl der Arbeitstage in einem Zeitraum
func (s *TimeAccountService) GetWorkingDaysInPeriod(startDate, endDate time.Time) (int, error) {
	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	return s.holidayService.GetWorkingDaysBetween(startDate, endDate, state), nil
}

// ValidateTimeEntry prüft, ob ein Zeiteintrag gültig ist
func (s *TimeAccountService) ValidateTimeEntry(entry *model.TimeEntry) error {
	if entry.Duration <= 0 {
		return fmt.Errorf("dauer muss größer als 0 sein")
	}

	if entry.Duration > 24 {
		return fmt.Errorf("dauer kann nicht mehr als 24 Stunden betragen")
	}

	if entry.StartTime.After(entry.EndTime) {
		return fmt.Errorf("startzeit muss vor endzeit liegen")
	}

	// Prüfe, ob das Datum in der Zukunft liegt
	if entry.Date.After(time.Now()) {
		return fmt.Errorf("datum kann nicht in der Zukunft liegen")
	}

	return nil
}

// CalculateExpectedHoursForEmployee berechnet die erwarteten Arbeitsstunden für einen Mitarbeiter
// basierend auf seinem Arbeitszeitmodell und dem Zeitraum
func (s *TimeAccountService) CalculateExpectedHoursForEmployee(employee *model.Employee, startDate, endDate time.Time) (float64, error) {
	// Bundesland aus den Einstellungen holen
	settings, err := s.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	workingDays := s.holidayService.GetWorkingDaysBetween(startDate, endDate, state)
	dailyHours := employee.GetWorkingHoursPerDay()

	if employee.WorkingDaysPerWeek > 0 && employee.WorkingDaysPerWeek < 5 {
		// Bei Teilzeit: Proportional reduzieren
		workingDaysAdjusted := float64(workingDays) * (float64(employee.WorkingDaysPerWeek) / 5.0)
		return workingDaysAdjusted * dailyHours, nil
	}

	return float64(workingDays) * dailyHours, nil
}

// GetOvertimeStatistics erstellt Überstunden-Statistiken für alle Mitarbeiter
func (s *TimeAccountService) GetOvertimeStatistics() (map[string]interface{}, error) {
	employees, err := s.employeeRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("fehler beim Abrufen der Mitarbeiter: %w", err)
	}

	var totalBalance float64
	var positiveCount, negativeCount, neutralCount int
	var employeesWithData int

	for _, employee := range employees {
		if len(employee.TimeEntries) == 0 {
			continue
		}

		employeesWithData++
		balance := employee.OvertimeBalance

		totalBalance += balance

		if balance > 0 {
			positiveCount++
		} else if balance < 0 {
			negativeCount++
		} else {
			neutralCount++
		}
	}

	averageBalance := float64(0)
	if employeesWithData > 0 {
		averageBalance = totalBalance / float64(employeesWithData)
	}

	statistics := map[string]interface{}{
		"totalEmployees": employeesWithData,
		"totalBalance":   totalBalance,
		"averageBalance": averageBalance,
		"positiveCount":  positiveCount,
		"negativeCount":  negativeCount,
		"neutralCount":   neutralCount,
		"lastCalculated": time.Now(),
	}

	return statistics, nil
}
