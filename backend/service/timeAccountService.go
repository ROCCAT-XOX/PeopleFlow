// backend/service/timeAccountService.go - Korrigierte Version

package service

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TimeAccountService verwaltet Zeitkonto-Berechnungen
type TimeAccountService struct {
	employeeRepo *repository.EmployeeRepository
	activityRepo *repository.ActivityRepository
}

// NewTimeAccountService erstellt einen neuen TimeAccountService
func NewTimeAccountService() *TimeAccountService {
	return &TimeAccountService{
		employeeRepo: repository.NewEmployeeRepository(),
		activityRepo: repository.NewActivityRepository(),
	}
}

// CalculateOvertimeForEmployee berechnet Überstunden für einen Mitarbeiter
func (s *TimeAccountService) CalculateOvertimeForEmployee(employee *model.Employee) error {
	if len(employee.TimeEntries) == 0 {
		return nil
	}

	// Zeiteinträge nach Wochen gruppieren
	weeklyData := s.groupTimeEntriesByWeek(employee.TimeEntries)

	// Bestehende WeeklyTimeEntries löschen und neu berechnen
	employee.WeeklyTimeEntries = []model.WeeklyTimeEntry{}
	employee.OvertimeBalance = 0.0

	// Für jede Woche berechnen
	for weekKey, entries := range weeklyData {
		weekEntry := s.calculateWeeklyOvertime(employee, weekKey, entries)
		employee.WeeklyTimeEntries = append(employee.WeeklyTimeEntries, weekEntry)
		employee.OvertimeBalance += weekEntry.OvertimeHours
	}

	// Sortiere WeeklyTimeEntries nach Datum
	sort.Slice(employee.WeeklyTimeEntries, func(i, j int) bool {
		return employee.WeeklyTimeEntries[i].WeekStartDate.Before(employee.WeeklyTimeEntries[j].WeekStartDate)
	})

	employee.LastTimeCalculated = time.Now()

	// Mitarbeiter in der Datenbank aktualisieren
	return s.employeeRepo.Update(employee)
}

// groupTimeEntriesByWeek gruppiert Zeiteinträge nach Kalenderwochen
func (s *TimeAccountService) groupTimeEntriesByWeek(timeEntries []model.TimeEntry) map[string][]model.TimeEntry {
	weeklyData := make(map[string][]model.TimeEntry)

	for _, entry := range timeEntries {
		// Montag der Woche als Schlüssel verwenden
		year, week := entry.Date.ISOWeek()
		monday := s.getMondayOfWeek(year, week)
		weekKey := monday.Format("2006-01-02")

		weeklyData[weekKey] = append(weeklyData[weekKey], entry)
	}

	return weeklyData
}

// calculateWeeklyOvertime berechnet Überstunden für eine Woche
func (s *TimeAccountService) calculateWeeklyOvertime(employee *model.Employee, weekKey string, entries []model.TimeEntry) model.WeeklyTimeEntry {
	weekStartDate, _ := time.Parse("2006-01-02", weekKey)
	weekEndDate := weekStartDate.AddDate(0, 0, 6) // Sonntag
	year, weekNumber := weekStartDate.ISOWeek()

	// Tatsächlich gearbeitete Stunden berechnen
	var actualHours float64
	daysWorked := make(map[string]bool)

	for _, entry := range entries {
		actualHours += entry.Duration
		dayKey := entry.Date.Format("2006-01-02")
		daysWorked[dayKey] = true
	}

	// Geplante Stunden für die Woche
	plannedHours := employee.GetWeeklyTargetHours()

	// Überstunden berechnen
	overtimeHours := actualHours - plannedHours

	return model.WeeklyTimeEntry{
		ID:            primitive.NewObjectID(),
		WeekStartDate: weekStartDate,
		WeekEndDate:   weekEndDate,
		Year:          year,
		WeekNumber:    weekNumber,
		PlannedHours:  plannedHours,
		ActualHours:   actualHours,
		OvertimeHours: overtimeHours,
		DaysWorked:    len(daysWorked),
		IsComplete:    true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// getMondayOfWeek gibt den Montag einer bestimmten Kalenderwoche zurück
func (s *TimeAccountService) getMondayOfWeek(year, week int) time.Time {
	// 4. Januar ist immer in Kalenderwoche 1
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)

	// Montag der ersten Kalenderwoche finden
	mondayWeek1 := jan4.AddDate(0, 0, -int(jan4.Weekday())+1)
	if jan4.Weekday() == time.Sunday {
		mondayWeek1 = mondayWeek1.AddDate(0, 0, -6)
	}

	// Montag der gewünschten Woche berechnen
	return mondayWeek1.AddDate(0, 0, 7*(week-1))
}

// RecalculateAllEmployeeOvertimes berechnet Überstunden für alle Mitarbeiter neu
func (s *TimeAccountService) RecalculateAllEmployeeOvertimes() error {
	employees, err := s.employeeRepo.FindAll()
	if err != nil {
		return err
	}

	for _, employee := range employees {
		if err := s.CalculateOvertimeForEmployee(employee); err != nil {
			fmt.Printf("Fehler bei der Überstunden-Berechnung für %s %s: %v\n",
				employee.FirstName, employee.LastName, err)
			continue
		}
	}

	return nil
}

// EmployeeOvertimeSummary repräsentiert die Überstunden-Übersicht eines Mitarbeiters
type EmployeeOvertimeSummary struct {
	EmployeeID        string                  `json:"employeeId"`
	EmployeeName      string                  `json:"employeeName"`
	CurrentBalance    float64                 `json:"currentBalance"`
	WeeklyTargetHours float64                 `json:"weeklyTargetHours"`
	LastCalculated    time.Time               `json:"lastCalculated"`
	WeeklyEntries     []model.WeeklyTimeEntry `json:"weeklyEntries"`
	CurrentWeek       *model.WeeklyTimeEntry  `json:"currentWeek,omitempty"`
}

// GetEmployeeOvertimeSummary erstellt eine Übersicht der Überstunden
func (s *TimeAccountService) GetEmployeeOvertimeSummary(employeeID string) (*EmployeeOvertimeSummary, error) {
	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		return nil, err
	}

	// Aktuelle Berechnung durchführen
	if err := s.CalculateOvertimeForEmployee(employee); err != nil {
		return nil, err
	}

	// Zusammenfassung erstellen
	summary := &EmployeeOvertimeSummary{
		EmployeeID:        employee.ID.Hex(),
		EmployeeName:      employee.FirstName + " " + employee.LastName,
		CurrentBalance:    employee.OvertimeBalance,
		WeeklyTargetHours: employee.GetWeeklyTargetHours(),
		LastCalculated:    employee.LastTimeCalculated,
		WeeklyEntries:     employee.WeeklyTimeEntries,
	}

	// Aktuelle Woche hinzufügen
	currentWeek := s.getCurrentWeekEntry(employee)
	if currentWeek != nil {
		summary.CurrentWeek = currentWeek
	}

	return summary, nil
}

// getCurrentWeekEntry erstellt einen Eintrag für die aktuelle Woche
func (s *TimeAccountService) getCurrentWeekEntry(employee *model.Employee) *model.WeeklyTimeEntry {
	now := time.Now()
	year, week := now.ISOWeek()
	weekStart := s.getMondayOfWeek(year, week)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// Zeiteinträge der aktuellen Woche sammeln
	var currentWeekEntries []model.TimeEntry
	for _, entry := range employee.TimeEntries {
		if entry.Date.After(weekStart.AddDate(0, 0, -1)) && entry.Date.Before(weekEnd.AddDate(0, 0, 1)) {
			currentWeekEntries = append(currentWeekEntries, entry)
		}
	}

	if len(currentWeekEntries) == 0 {
		return nil
	}

	weekKey := weekStart.Format("2006-01-02")
	weekEntry := s.calculateWeeklyOvertime(employee, weekKey, currentWeekEntries)
	return &weekEntry
}
