package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// FilterParams enthält die Filter-Parameter für die Statistik
type FilterParams struct {
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	EmployeeIDs  []string  `json:"employeeIds"`
	ProjectID    string    `json:"projectId"`
	DateRangeKey string    `json:"dateRangeKey"` // z.B. 'this-month', 'last-month', etc.
}

// StatisticsAPIHandler verwaltet alle API-Anfragen für die Statistik-Seite
type StatisticsAPIHandler struct {
	employeeRepo *repository.EmployeeRepository
}

// NewStatisticsAPIHandler erstellt einen neuen StatisticsAPIHandler
func NewStatisticsAPIHandler() *StatisticsAPIHandler {
	return &StatisticsAPIHandler{
		employeeRepo: repository.NewEmployeeRepository(),
	}
}

// GetFilteredStatistics liefert gefilterte Statistikdaten zurück
func (h *StatisticsAPIHandler) GetFilteredStatistics(c *gin.Context) {
	// Parameter aus der Anfrage extrahieren
	var filter FilterParams
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültige Filter-Parameter: " + err.Error(),
		})
		return
	}

	// Alle Mitarbeiter abrufen
	employees, err := h.employeeRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Abrufen der Mitarbeiter: " + err.Error(),
		})
		return
	}

	// Filter anwenden

	// Mitarbeiter filtern
	filteredEmployees := employees
	if len(filter.EmployeeIDs) > 0 {
		filteredEmployees = []*model.Employee{}
		for _, emp := range employees {
			for _, id := range filter.EmployeeIDs {
				if emp.ID.Hex() == id {
					filteredEmployees = append(filteredEmployees, emp)
					break
				}
			}
		}
	}

	// Gesamtarbeitszeit berechnen
	var totalHours float64
	var projectHours = make(map[string]float64)
	var weekdayHours = make(map[int]float64)
	var weekdayCounts = make(map[int]int)

	for _, emp := range filteredEmployees {
		for _, entry := range emp.TimeEntries {
			// Zeitfilter anwenden
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}

			// Projektfilter anwenden
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			// Zeit zu Gesamtzeit hinzufügen
			totalHours += entry.Duration

			// Zeit nach Projekt aufschlüsseln
			projectHours[entry.ProjectID] += entry.Duration

			// Zeit nach Wochentag aufschlüsseln
			weekday := int(entry.Date.Weekday())
			weekdayHours[weekday] += entry.Duration
			weekdayCounts[weekday]++
		}
	}

	// Durchschnittliche Stunden pro Wochentag berechnen
	weekdayAvg := make(map[string]float64)
	weekdays := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}

	for day := 0; day < 7; day++ {
		if weekdayCounts[day] > 0 {
			weekdayAvg[weekdays[day]] = weekdayHours[day] / float64(weekdayCounts[day])
		} else {
			weekdayAvg[weekdays[day]] = 0
		}
	}

	// Abwesenheitstage berechnen
	var totalAbsenceDays float64
	var absenceByType = make(map[string]float64)
	var absenceByMonth = make(map[string]map[string]float64)

	for _, emp := range filteredEmployees {
		for _, absence := range emp.Absences {
			// Nur genehmigte Abwesenheiten zählen
			if absence.Status != "approved" {
				continue
			}

			// Zeitfilter anwenden
			if !filter.StartDate.IsZero() && absence.StartDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && absence.EndDate.After(filter.EndDate) {
				continue
			}

			// Tage zur Gesamtzahl hinzufügen
			totalAbsenceDays += absence.Days

			// Nach Typ aufschlüsseln
			absenceType := absence.Type
			absenceByType[absenceType] += absence.Days

			// Nach Monat aufschlüsseln
			monthKey := absence.StartDate.Format("2006-01")

			if absenceByMonth[monthKey] == nil {
				absenceByMonth[monthKey] = make(map[string]float64)
			}
			absenceByMonth[monthKey][absenceType] += absence.Days
		}
	}

	// Beispielhafte Produktivitätsrate (in einer echten Anwendung würde dies
	// auf tatsächlichen Leistungsdaten, Zielvereinbarungen, etc. basieren)
	productivityRate := 85.3 // 85.3%

	// Projektinformationen zusammenstellen
	type ProjectSummary struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Hours float64 `json:"hours"`
		Share float64 `json:"share"`
	}

	var projectSummaries []ProjectSummary
	for projectID, hours := range projectHours {
		projectName := "Unbekannt"

		// Projektname suchen
		for _, emp := range employees {
			for _, proj := range emp.ProjectAssignments {
				if proj.ProjectID == projectID {
					projectName = proj.ProjectName
					break
				}
			}
			if projectName != "Unbekannt" {
				break
			}
		}

		share := 0.0
		if totalHours > 0 {
			share = hours / totalHours * 100
		}

		projectSummaries = append(projectSummaries, ProjectSummary{
			ID:    projectID,
			Name:  projectName,
			Hours: hours,
			Share: share,
		})
	}

	// Nach Stunden sortieren
	sort.Slice(projectSummaries, func(i, j int) bool {
		return projectSummaries[i].Hours > projectSummaries[j].Hours
	})

	// Daten für die Antwort aufbereiten
	result := gin.H{
		"success":          true,
		"totalHours":       totalHours,
		"totalAbsenceDays": totalAbsenceDays,
		"productivityRate": productivityRate,
		"weekdayHours":     weekdayAvg,
		"projectHours":     projectSummaries,
		"absenceByType":    absenceByType,
		"absenceByMonth":   absenceByMonth,
	}

	c.JSON(http.StatusOK, result)
}

// ExtendedStatisticsResponse represents the complete statistics data structure
type ExtendedStatisticsResponse struct {
	// Overview data
	TotalHours       float64 `json:"totalHours"`
	ProductivityRate float64 `json:"productivityRate"`
	TotalAbsenceDays float64 `json:"totalAbsenceDays"`
	ActiveProjects   int     `json:"activeProjects"`

	// Charts data
	WeekdayHours         map[string]float64    `json:"weekdayHours"`
	ProjectHours         []ProjectHourSummary  `json:"projectHours"`
	ProductivityTimeline []MonthlyProductivity `json:"productivityTimeline"`
	AbsenceTypes         map[string]float64    `json:"absenceTypes"`

	// Productivity data
	ProjectProductivity  []ProjectProductivity  `json:"projectProductivity"`
	EmployeeProductivity []EmployeeProductivity `json:"employeeProductivity"`
	ProductivityRanking  []ProductivityRanking  `json:"productivityRanking"`

	// Project data
	ProjectProgress    []ProjectProgress    `json:"projectProgress"`
	ResourceAllocation []ResourceAllocation `json:"resourceAllocation"`
	ProjectDetails     []ProjectDetail      `json:"projectDetails"`

	// Absence data
	AbsenceTypeDetail map[string]float64 `json:"absenceTypeDetail"`
	AbsenceTimeline   []MonthlyAbsence   `json:"absenceTimeline"`
	CurrentAbsences   []CurrentAbsence   `json:"currentAbsences"`
}

// Various data structures for the response
type ProjectHourSummary struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Hours float64 `json:"hours"`
	Share float64 `json:"share"`
}

type MonthlyProductivity struct {
	Month string  `json:"month"`
	Rate  float64 `json:"rate"`
}

type ProjectProductivity struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

type EmployeeProductivity struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}

type ProductivityRanking struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Department       string  `json:"department"`
	Hours            float64 `json:"hours"`
	ProductivityRate float64 `json:"productivityRate"`
	Trend            float64 `json:"trend"`
	TrendFormatted   string  `json:"trendFormatted"`
	IsTrendPositive  bool    `json:"isTrendPositive"`
	IsTrendNegative  bool    `json:"isTrendNegative"`
	HasProfileImage  bool    `json:"hasProfileImage"`
}

type ProjectProgress struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
}

type ResourceAllocation struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TeamSize int    `json:"teamSize"`
}

type ProjectDetail struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Status              string  `json:"status"`
	TeamSize            int     `json:"teamSize"`
	Hours               float64 `json:"hours"`
	HoursFormatted      string  `json:"hoursFormatted"`
	Efficiency          float64 `json:"efficiency"`
	EfficiencyFormatted string  `json:"efficiencyFormatted"`
	EfficiencyClass     string  `json:"efficiencyClass"`
}

type MonthlyAbsence struct {
	Month    string  `json:"month"`
	Vacation float64 `json:"vacation"`
	Sick     float64 `json:"sick"`
	Other    float64 `json:"other"`
}

type CurrentAbsence struct {
	ID               string    `json:"id"`
	EmployeeID       string    `json:"employeeId"`
	EmployeeName     string    `json:"employeeName"`
	Type             string    `json:"type"`
	StartDate        time.Time `json:"startDate"`
	EndDate          time.Time `json:"endDate"`
	Days             float64   `json:"days"`
	Status           string    `json:"status"`
	HasProfileImage  bool      `json:"hasProfileImage"`
	AffectedProjects []string  `json:"affectedProjects"`
}

// GetExtendedStatistics provides comprehensive statistics data with filters
func (h *StatisticsAPIHandler) GetExtendedStatistics(c *gin.Context) {
	// Get filter parameters from the request
	var filter FilterParams
	if err := c.ShouldBindJSON(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültige Filter-Parameter: " + err.Error(),
		})
		return
	}

	// Retrieve all employees
	employees, err := h.employeeRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Abrufen der Mitarbeiter: " + err.Error(),
		})
		return
	}

	// Apply filters to get filtered employees
	filteredEmployees := filterEmployees(employees, filter)

	// Prepare the response structure
	response := ExtendedStatisticsResponse{}

	// Calculate overview metrics
	response.TotalHours = calculateTotalHours(filteredEmployees, filter)
	response.ProductivityRate = calculateProductivityRate(filteredEmployees, filter)
	response.TotalAbsenceDays = calculateTotalAbsenceDays(filteredEmployees, filter)
	response.ActiveProjects = countActiveProjects(filteredEmployees)

	// Calculate weekday hours distribution
	response.WeekdayHours = calculateWeekdayHours(filteredEmployees, filter)

	// Calculate project hours distribution
	response.ProjectHours = calculateProjectHours(filteredEmployees, filter)

	// Calculate productivity timeline
	response.ProductivityTimeline = calculateProductivityTimeline(filteredEmployees, filter)

	// Calculate absence types
	response.AbsenceTypes = calculateAbsenceTypes(filteredEmployees, filter)

	// Calculate project productivity
	response.ProjectProductivity = calculateProjectProductivity(filteredEmployees, filter)

	// Calculate employee productivity
	response.EmployeeProductivity = calculateEmployeeProductivity(filteredEmployees, filter)

	// Calculate productivity ranking
	response.ProductivityRanking = calculateProductivityRanking(filteredEmployees, filter)

	// Calculate project progress
	response.ProjectProgress = calculateProjectProgress(filteredEmployees, filter)

	// Calculate resource allocation
	response.ResourceAllocation = calculateResourceAllocation(filteredEmployees, filter)

	// Calculate project details
	response.ProjectDetails = calculateProjectDetails(filteredEmployees, filter)

	// Calculate detailed absence types
	response.AbsenceTypeDetail = calculateAbsenceTypeDetail(filteredEmployees, filter)

	// Calculate absence timeline
	response.AbsenceTimeline = calculateAbsenceTimeline(filteredEmployees, filter)

	// Calculate current absences
	response.CurrentAbsences = calculateCurrentAbsences(filteredEmployees, filter)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// Helper function to filter employees based on filter parameters
func filterEmployees(employees []*model.Employee, filter FilterParams) []*model.Employee {
	var filteredEmployees []*model.Employee

	for _, emp := range employees {
		include := true

		// Filter by employee ID
		if len(filter.EmployeeIDs) > 0 {
			found := false
			for _, id := range filter.EmployeeIDs {
				if emp.ID.Hex() == id {
					found = true
					break
				}
			}
			if !found {
				include = false
			}
		}

		// Include this employee in results
		if include {
			filteredEmployees = append(filteredEmployees, emp)
		}
	}

	return filteredEmployees
}

// Implementation of calculation functions
func calculateTotalHours(employees []*model.Employee, filter FilterParams) float64 {
	var totalHours float64

	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			// Apply date filter
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}

			// Apply project filter
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			totalHours += entry.Duration
		}
	}

	return totalHours
}

// Calculate productivity rate
func calculateProductivityRate(employees []*model.Employee, filter FilterParams) float64 {
	var totalProductivity float64
	var totalEmployees int

	for _, emp := range employees {
		// Zählen nur Mitarbeiter mit Zeiteinträgen
		var hasTimeEntries bool
		for _, entry := range emp.TimeEntries {
			// Filter anwenden
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			hasTimeEntries = true
			break
		}

		if !hasTimeEntries {
			continue
		}

		// Berechnung der individuellen Produktivitätsrate
		var employeeProductivity float64

		// 1. Berechnung basierend auf dem Verhältnis der tatsächlichen Arbeitszeit zur erwarteten Zeit
		var actualHours float64
		var expectedHours float64 = 40.0 * 4 // Beispiel: 40 Stunden pro Woche für 4 Wochen

		for _, entry := range emp.TimeEntries {
			// Filter anwenden
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}

			actualHours += entry.Duration
		}

		// 2. Berechnung basierend auf Projektzuordnungen
		projectEfficiency := 1.0
		projectCount := 0

		for _, project := range emp.ProjectAssignments {
			// Filter anwenden
			if !filter.StartDate.IsZero() && project.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && project.StartDate.After(filter.EndDate) {
				continue
			}

			// Hier könnten weitere Faktoren einfließen, z.B. Projektbewertungen
			projectCount++
		}

		if projectCount > 0 {
			projectEfficiency = float64(minInt(projectCount, 3)) / 3.0 // Max 3 Projekte als optimal
		}

		// 3. Berücksichtigung von Abwesenheiten
		absenceRatio := 1.0
		totalAbsenceDays := 0.0

		for _, absence := range emp.Absences {
			// Nur genehmigte Abwesenheiten zählen
			if absence.Status != "approved" {
				continue
			}

			// Filter anwenden
			if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
				continue
			}

			// Krankheit reduziert die Produktivität nicht
			if absence.Type != "sick" {
				totalAbsenceDays += absence.Days
			}
		}

		// Arbeitszeit im Zeitraum (z.B. 20 Arbeitstage)
		workingDaysInPeriod := 20.0
		if totalAbsenceDays > 0 {
			absenceRatio = (workingDaysInPeriod - totalAbsenceDays) / workingDaysInPeriod
			if absenceRatio < 0 {
				absenceRatio = 0
			}
		}

		// Gewichtete Berechnung der Gesamtproduktivität
		timeWeight := 0.5
		projectWeight := 0.3
		absenceWeight := 0.2

		timeRatio := min(actualHours/expectedHours, 1.0) // max 100%

		employeeProductivity = (timeRatio * timeWeight) +
			(projectEfficiency * projectWeight) +
			(absenceRatio * absenceWeight)

		// Skalierung auf 0-100%
		employeeProductivity *= 100

		totalProductivity += employeeProductivity
		totalEmployees++
	}

	// Durchschnittliche Produktivitätsrate
	if totalEmployees > 0 {
		return totalProductivity / float64(totalEmployees)
	}

	return 85.0 // Standardwert, falls keine Daten vorhanden sind
}

// Hilfsfunktion für Min
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Calculate total absence days
func calculateTotalAbsenceDays(employees []*model.Employee, filter FilterParams) float64 {
	var totalDays float64

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Only count approved absences
			if absence.Status != "approved" {
				continue
			}

			// Apply date filter
			if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
				continue
			}

			totalDays += absence.Days
		}
	}

	return totalDays
}

// Count active projects
func countActiveProjects(employees []*model.Employee) int {
	// Use a map to deduplicate projects
	activeProjects := make(map[string]bool)

	for _, emp := range employees {
		for _, proj := range emp.ProjectAssignments {
			// Check if project is active (end date in future or not set)
			if proj.EndDate.IsZero() || proj.EndDate.After(time.Now()) {
				activeProjects[proj.ProjectID] = true
			}
		}
	}

	return len(activeProjects)
}

// Calculate hours worked by weekday
func calculateWeekdayHours(employees []*model.Employee, filter FilterParams) map[string]float64 {
	weekdays := map[string]float64{
		"Montag":     0,
		"Dienstag":   0,
		"Mittwoch":   0,
		"Donnerstag": 0,
		"Freitag":    0,
		"Samstag":    0,
		"Sonntag":    0,
	}

	weekdayCounts := map[string]int{
		"Montag":     0,
		"Dienstag":   0,
		"Mittwoch":   0,
		"Donnerstag": 0,
		"Freitag":    0,
		"Samstag":    0,
		"Sonntag":    0,
	}

	weekdayNames := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}

	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			// Apply filters
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			// Get weekday name
			weekdayIndex := int(entry.Date.Weekday())
			weekdayName := weekdayNames[weekdayIndex]

			// Add hours to the appropriate weekday
			weekdays[weekdayName] += entry.Duration
			weekdayCounts[weekdayName]++
		}
	}

	// Calculate average hours per weekday
	for day, hours := range weekdays {
		if weekdayCounts[day] > 0 {
			weekdays[day] = hours / float64(weekdayCounts[day])
		}
	}

	return weekdays
}

// Calculate hours by project
func calculateProjectHours(employees []*model.Employee, filter FilterParams) []ProjectHourSummary {
	projectHours := make(map[string]float64)
	projectNames := make(map[string]string)

	var totalHours float64

	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			// Apply filters
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			projectHours[entry.ProjectID] += entry.Duration
			projectNames[entry.ProjectID] = entry.ProjectName
			totalHours += entry.Duration
		}
	}

	// Create project hour summaries
	var summaries []ProjectHourSummary

	for id, hours := range projectHours {
		share := 0.0
		if totalHours > 0 {
			share = (hours / totalHours) * 100
		}

		summaries = append(summaries, ProjectHourSummary{
			ID:    id,
			Name:  projectNames[id],
			Hours: hours,
			Share: share,
		})
	}

	// Sort by hours in descending order
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Hours > summaries[j].Hours
	})

	return summaries
}

// Calculate productivity timeline
func calculateProductivityTimeline(employees []*model.Employee, filter FilterParams) []MonthlyProductivity {
	// In a real implementation, this would calculate monthly productivity rates
	// based on actual data. For this example, we'll generate sample data.

	months := []string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}
	var timeline []MonthlyProductivity

	for i, month := range months {
		// Generate a productivity rate between 80% and 90%
		rate := 80.0 + rand.Float64()*10.0

		// Add some trend - higher in later months
		trend := float64(i) * 0.2
		rate += trend

		if rate > 95.0 {
			rate = 95.0
		}

		timeline = append(timeline, MonthlyProductivity{
			Month: month,
			Rate:  rate,
		})
	}

	return timeline
}

// Calculate absence types
func calculateAbsenceTypes(employees []*model.Employee, filter FilterParams) map[string]float64 {
	absenceTypes := make(map[string]float64)

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Only count approved absences
			if absence.Status != "approved" {
				continue
			}

			// Apply date filter
			if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
				continue
			}

			// Map internal type codes to human-readable names
			typeName := "Sonstige"
			switch absence.Type {
			case "vacation":
				typeName = "Urlaub"
			case "sick":
				typeName = "Krankheit"
			case "special":
				typeName = "Sonderurlaub"
			}

			absenceTypes[typeName] += absence.Days
		}
	}

	return absenceTypes
}

// Calculate project productivity
func calculateProjectProductivity(employees []*model.Employee, filter FilterParams) []ProjectProductivity {
	// Get unique projects from employee time entries
	projects := make(map[string]string) // ProjectID -> Name

	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			// Apply date filter
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}

			// Only include the specified project if filter is set
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			projects[entry.ProjectID] = entry.ProjectName
		}
	}

	// Generate productivity rates for each project
	var projectProductivity []ProjectProductivity

	for id, name := range projects {
		// In a real implementation, calculate based on various metrics
		// For demo, generate a rate between 65% and 95%
		rate := 65.0 + rand.Float64()*30.0

		// Special case: Security Audit project has lower productivity
		if name == "Security Audit" {
			rate = 65.0 + rand.Float64()*5.0 // 65-70%
		}

		projectProductivity = append(projectProductivity, ProjectProductivity{
			ID:   id,
			Name: name,
			Rate: rate,
		})
	}

	// Sort by productivity rate in descending order
	sort.Slice(projectProductivity, func(i, j int) bool {
		return projectProductivity[i].Rate > projectProductivity[j].Rate
	})

	return projectProductivity
}

// Calculate employee productivity
func calculateEmployeeProductivity(employees []*model.Employee, filter FilterParams) []EmployeeProductivity {
	var employeeProductivity []EmployeeProductivity

	for _, emp := range employees {
		// Skip employees with no time entries
		hasTimeEntries := false
		for _, entry := range emp.TimeEntries {
			// Apply filters
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			hasTimeEntries = true
			break
		}

		if !hasTimeEntries {
			continue
		}

		// In a real implementation, calculate based on various metrics
		// For demo, generate a rate between 75% and 95%
		rate := 75.0 + rand.Float64()*20.0

		employeeProductivity = append(employeeProductivity, EmployeeProductivity{
			ID:   emp.ID.Hex(),
			Name: emp.FirstName + " " + emp.LastName,
			Rate: rate,
		})
	}

	// Sort by productivity rate in descending order
	sort.Slice(employeeProductivity, func(i, j int) bool {
		return employeeProductivity[i].Rate > employeeProductivity[j].Rate
	})

	return employeeProductivity
}

// Calculate productivity ranking
func calculateProductivityRanking(employees []*model.Employee, filter FilterParams) []ProductivityRanking {
	var ranking []ProductivityRanking

	for _, emp := range employees {
		// Calculate total hours for this employee
		var hours float64
		for _, entry := range emp.TimeEntries {
			// Apply filters
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			hours += entry.Duration
		}

		// Skip employees with no filtered time entries
		if hours == 0 {
			continue
		}

		// Calculate productivity rate (in a real system, based on metrics)
		rate := 75.0 + rand.Float64()*20.0

		// Calculate trend (change from previous period)
		trend := -5.0 + rand.Float64()*10.0

		// Format trend for display
		trendFormatted := fmt.Sprintf("%.1f", trend)
		if trend > 0 {
			trendFormatted = "+" + trendFormatted
		}
		trendFormatted += "%"

		ranking = append(ranking, ProductivityRanking{
			ID:               emp.ID.Hex(),
			Name:             emp.FirstName + " " + emp.LastName,
			Department:       string(emp.Department),
			Hours:            hours,
			ProductivityRate: rate,
			Trend:            trend,
			TrendFormatted:   trendFormatted,
			IsTrendPositive:  trend > 0,
			IsTrendNegative:  trend < 0,
			HasProfileImage:  len(emp.ProfileImageData.Data) > 0,
		})
	}

	// Sort by productivity rate in descending order
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].ProductivityRate > ranking[j].ProductivityRate
	})

	return ranking
}

// Calculate project progress
func calculateProjectProgress(employees []*model.Employee, filter FilterParams) []ProjectProgress {
	// Get unique projects
	projects := make(map[string]string) // ProjectID -> Name

	for _, emp := range employees {
		// From time entries
		for _, entry := range emp.TimeEntries {
			projects[entry.ProjectID] = entry.ProjectName
		}

		// From project assignments
		for _, assignment := range emp.ProjectAssignments {
			projects[assignment.ProjectID] = assignment.ProjectName
		}
	}

	var projectProgress []ProjectProgress

	// Status mapping (would come from actual data in a real implementation)
	statusProgress := map[string]float64{
		"Abgeschlossen": 100.0,
		"In Arbeit":     50.0,
		"Kritisch":      30.0,
		"Geplant":       10.0,
	}

	// Example projects with known status for the demo
	knownProjects := map[string]string{
		"Website Redesign":   "In Arbeit",
		"Mobile App":         "In Arbeit",
		"Datenmigration":     "Abgeschlossen",
		"Security Audit":     "Kritisch",
		"CRM Implementation": "Geplant",
	}

	for id, name := range projects {
		// Look up status (or default to "In Arbeit")
		status := "In Arbeit"
		for knownName, knownStatus := range knownProjects {
			if name == knownName {
				status = knownStatus
				break
			}
		}

		// Get progress from status or generate random progress
		progress := statusProgress[status]
		if progress == 0 {
			// Random progress between 10% and 90%
			progress = 10.0 + rand.Float64()*80.0
		}

		projectProgress = append(projectProgress, ProjectProgress{
			ID:       id,
			Name:     name,
			Progress: progress,
		})
	}

	// Sort by name for consistency
	sort.Slice(projectProgress, func(i, j int) bool {
		return projectProgress[i].Name < projectProgress[j].Name
	})

	return projectProgress
}

// Calculate resource allocation
func calculateResourceAllocation(employees []*model.Employee, filter FilterParams) []ResourceAllocation {
	// Count employees assigned to each project
	projectTeamSizes := make(map[string]map[string]bool) // ProjectID -> Employee IDs
	projectNames := make(map[string]string)              // ProjectID -> Name

	for _, emp := range employees {
		for _, assignment := range emp.ProjectAssignments {
			// Apply date filter if specified
			if !filter.StartDate.IsZero() && assignment.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && assignment.StartDate.After(filter.EndDate) {
				continue
			}

			// Only include the specified project if filter is set
			if filter.ProjectID != "" && assignment.ProjectID != filter.ProjectID {
				continue
			}

			if projectTeamSizes[assignment.ProjectID] == nil {
				projectTeamSizes[assignment.ProjectID] = make(map[string]bool)
			}

			projectTeamSizes[assignment.ProjectID][emp.ID.Hex()] = true
			projectNames[assignment.ProjectID] = assignment.ProjectName
		}
	}

	var resourceAllocation []ResourceAllocation

	for id, assignedEmployees := range projectTeamSizes {
		resourceAllocation = append(resourceAllocation, ResourceAllocation{
			ID:       id,
			Name:     projectNames[id],
			TeamSize: len(assignedEmployees),
		})
	}

	// Sort by team size in descending order
	sort.Slice(resourceAllocation, func(i, j int) bool {
		return resourceAllocation[i].TeamSize > resourceAllocation[j].TeamSize
	})

	return resourceAllocation
}

// Calculate project details
func calculateProjectDetails(employees []*model.Employee, filter FilterParams) []ProjectDetail {
	// Daten nach Projekt sammeln
	type projectData struct {
		hours           float64
		teamSize        int
		status          string
		plannedDuration float64 // Geplante Dauer des Projekts in Tagen
		actualDuration  float64 // Tatsächliche Dauer in Tagen
		startDate       time.Time
		endDate         time.Time
	}

	projectDataMap := make(map[string]*projectData)
	projectNames := make(map[string]string)
	employeesByProject := make(map[string]map[string]bool)

	// Bekannte Projektstatus
	knownProjects := map[string]string{
		"Website Redesign":   "In Arbeit",
		"Mobile App":         "In Arbeit",
		"Datenmigration":     "Abgeschlossen",
		"Security Audit":     "Kritisch",
		"CRM Implementation": "Geplant",
	}

	// Zeiteinträge verarbeiten
	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			// Filter anwenden
			if !filter.StartDate.IsZero() && entry.Date.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && entry.Date.After(filter.EndDate) {
				continue
			}
			if filter.ProjectID != "" && entry.ProjectID != filter.ProjectID {
				continue
			}

			// Projektdaten initialisieren, falls noch nicht vorhanden
			if projectDataMap[entry.ProjectID] == nil {
				projectDataMap[entry.ProjectID] = &projectData{}
			}

			// Mitarbeiter-Map für dieses Projekt initialisieren
			if employeesByProject[entry.ProjectID] == nil {
				employeesByProject[entry.ProjectID] = make(map[string]bool)
			}

			// Stunden zum Projekt hinzufügen
			projectDataMap[entry.ProjectID].hours += entry.Duration

			// Mitarbeiter zum Projekt hinzufügen
			employeesByProject[entry.ProjectID][emp.ID.Hex()] = true

			// Projektname speichern
			projectNames[entry.ProjectID] = entry.ProjectName
		}
	}

	// Projekttermine verarbeiten für Effizienzberechnung
	for _, emp := range employees {
		for _, assignment := range emp.ProjectAssignments {
			// Filter anwenden
			if !filter.StartDate.IsZero() && assignment.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && assignment.StartDate.After(filter.EndDate) {
				continue
			}

			// Projektdaten initialisieren, falls noch nicht vorhanden
			if projectDataMap[assignment.ProjectID] == nil {
				projectDataMap[assignment.ProjectID] = &projectData{}
			}

			// Aktualisierung der Projekt-Zeiträume
			if projectDataMap[assignment.ProjectID].startDate.IsZero() ||
				assignment.StartDate.Before(projectDataMap[assignment.ProjectID].startDate) {
				projectDataMap[assignment.ProjectID].startDate = assignment.StartDate
			}

			if assignment.EndDate.After(projectDataMap[assignment.ProjectID].endDate) {
				projectDataMap[assignment.ProjectID].endDate = assignment.EndDate
			}

			// Projektname speichern
			projectNames[assignment.ProjectID] = assignment.ProjectName
		}
	}

	// Verbleibende Daten berechnen und Projektdetails erstellen
	var projectDetails []ProjectDetail

	for id, data := range projectDataMap {
		// Projektname abrufen
		name := projectNames[id]

		// Teamgröße festlegen
		data.teamSize = len(employeesByProject[id])

		// Status festlegen (aus bekannten Projekten oder Standard "In Arbeit")
		data.status = "In Arbeit"
		for knownName, status := range knownProjects {
			if name == knownName {
				data.status = status
				break
			}
		}

		// Effizienz berechnen (basierend auf realen Metriken)
		// 1. Verhältnis von geplanter zu tatsächlicher Dauer
		var timeEfficiency float64 = 90.0 // Standardwert

		if !data.startDate.IsZero() && !data.endDate.IsZero() {
			plannedDuration := data.endDate.Sub(data.startDate).Hours() / 24 // in Tagen

			// Aktuelle Dauer berechnen (bis heute oder Enddatum)
			var actualDuration float64
			now := time.Now()
			if now.After(data.endDate) {
				actualDuration = data.endDate.Sub(data.startDate).Hours() / 24
			} else {
				actualDuration = now.Sub(data.startDate).Hours() / 24
			}

			// Falls ein Projekt abgeschlossen ist oder in Arbeit mit überschrittenem Enddatum
			if data.status == "Abgeschlossen" {
				timeEfficiency = 95.0 // Hohe Effizienz für abgeschlossene Projekte
			} else if now.After(data.endDate) {
				// Projekt überzieht Zeitplan
				factor := plannedDuration / max(actualDuration, 1.0)
				timeEfficiency = min(factor*100, 90.0) // max 90% für überzogene Projekte
			} else {
				// Projekt im Zeitplan
				timeEfficiency = 85.0 + (5.0 * float64(data.teamSize) / max(float64(10.0), 10.0))
			}
		}

		// 2. Stunden pro Teamgröße und Dauer
		var resourceEfficiency float64 = 85.0

		if data.teamSize > 0 && data.hours > 0 {
			// Optimales Verhältnis von Stunden pro Person: ~40h/Woche
			hoursPerPerson := data.hours / float64(data.teamSize)

			// Umrechnung in Wochen (grob)
			var weekCount float64 = 4.0 // Standard: 1 Monat

			if !data.startDate.IsZero() && !data.endDate.IsZero() {
				weekCount = data.endDate.Sub(data.startDate).Hours() / (24 * 7)
			}

			if weekCount <= 0 {
				weekCount = 1.0 // Mindestens eine Woche
			}

			optimalHoursPerWeek := 40.0
			actualHoursPerWeek := hoursPerPerson / weekCount

			resourceEfficiency = min((actualHoursPerWeek/optimalHoursPerWeek)*100, 100.0)
		}

		// 3. Status-basierte Effizienz
		var statusEfficiency float64 = 85.0

		switch data.status {
		case "Abgeschlossen":
			statusEfficiency = 95.0
		case "In Arbeit":
			statusEfficiency = 85.0
		case "Kritisch":
			statusEfficiency = 65.0
		case "Geplant":
			statusEfficiency = 90.0 // Geplante Projekte haben noch keine Probleme
		}

		// Gewichtete Gesamteffizienz
		timeWeight := 0.4
		resourceWeight := 0.4
		statusWeight := 0.2

		efficiency := (timeEfficiency * timeWeight) +
			(resourceEfficiency * resourceWeight) +
			(statusEfficiency * statusWeight)

		// Security Audit Projekt hat niedrigere Effizienz
		if name == "Security Audit" {
			efficiency = min(efficiency, 65.0+5.0) // Max 70%
		}

		// Effizienz für die Anzeige formatieren
		efficiencyFormatted := fmt.Sprintf("%.1f%%", efficiency)

		// CSS-Klasse für die Effizienzanzeige bestimmen
		efficiencyClass := "bg-green-600"
		if efficiency < 75.0 {
			efficiencyClass = "bg-red-600"
		} else if efficiency < 85.0 {
			efficiencyClass = "bg-yellow-600"
		}

		projectDetails = append(projectDetails, ProjectDetail{
			ID:                  id,
			Name:                name,
			Status:              data.status,
			TeamSize:            data.teamSize,
			Hours:               data.hours,
			HoursFormatted:      fmt.Sprintf("%.1f Std", data.hours),
			Efficiency:          efficiency,
			EfficiencyFormatted: efficiencyFormatted,
			EfficiencyClass:     efficiencyClass,
		})
	}

	// Nach Stunden absteigend sortieren
	sort.Slice(projectDetails, func(i, j int) bool {
		return projectDetails[i].Hours > projectDetails[j].Hours
	})

	return projectDetails
}

// Hilfsfunktion für Max
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Calculate detailed absence types
func calculateAbsenceTypeDetail(employees []*model.Employee, filter FilterParams) map[string]float64 {
	absenceTypeDetail := make(map[string]float64)

	// Default types
	absenceTypeDetail["Urlaub"] = 0
	absenceTypeDetail["Krankheit"] = 0
	absenceTypeDetail["Sonderurlaub"] = 0
	absenceTypeDetail["Elternzeit"] = 0
	absenceTypeDetail["Fortbildung"] = 0

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Only count approved absences
			if absence.Status != "approved" {
				continue
			}

			// Apply date filter
			if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
				continue
			}
			if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
				continue
			}

			// Map internal type codes to detailed types
			switch absence.Type {
			case "vacation":
				absenceTypeDetail["Urlaub"] += absence.Days
			case "sick":
				absenceTypeDetail["Krankheit"] += absence.Days
			case "special":
				reason := strings.ToLower(absence.Reason)
				if strings.Contains(reason, "eltern") {
					absenceTypeDetail["Elternzeit"] += absence.Days
				} else if strings.Contains(reason, "fortbild") || strings.Contains(reason, "training") || strings.Contains(reason, "seminar") {
					absenceTypeDetail["Fortbildung"] += absence.Days
				} else {
					absenceTypeDetail["Sonderurlaub"] += absence.Days
				}
			default:
				absenceTypeDetail["Sonderurlaub"] += absence.Days
			}
		}
	}

	// Remove empty categories
	for key, value := range absenceTypeDetail {
		if value == 0 {
			delete(absenceTypeDetail, key)
		}
	}

	return absenceTypeDetail
}

// Calculate absence timeline
func calculateAbsenceTimeline(employees []*model.Employee, filter FilterParams) []MonthlyAbsence {
	// Initialize monthly data
	months := []string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"}
	absenceByMonth := make(map[string]map[string]float64)

	for _, month := range months {
		absenceByMonth[month] = map[string]float64{
			"vacation": 0,
			"sick":     0,
			"other":    0,
		}
	}

	// Current year
	currentYear := time.Now().Year()

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Only count approved absences
			if absence.Status != "approved" {
				continue
			}

			// Only count absences in the current year or the filtered date range
			if absence.StartDate.Year() != currentYear {
				// If not in filter range, skip
				if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
					continue
				}
				if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
					continue
				}
			}

			// Get month abbreviation
			monthIndex := int(absence.StartDate.Month()) - 1
			if monthIndex < 0 || monthIndex >= len(months) {
				continue // Skip invalid months
			}
			month := months[monthIndex]

			// Add days to the appropriate type
			switch absence.Type {
			case "vacation":
				absenceByMonth[month]["vacation"] += absence.Days
			case "sick":
				absenceByMonth[month]["sick"] += absence.Days
			default:
				absenceByMonth[month]["other"] += absence.Days
			}
		}
	}

	// Create timeline data
	var timeline []MonthlyAbsence

	for _, month := range months {
		timeline = append(timeline, MonthlyAbsence{
			Month:    month,
			Vacation: absenceByMonth[month]["vacation"],
			Sick:     absenceByMonth[month]["sick"],
			Other:    absenceByMonth[month]["other"],
		})
	}

	return timeline
}

// Calculate current absences (for the next 30 days)
func calculateCurrentAbsences(employees []*model.Employee, filter FilterParams) []CurrentAbsence {
	var currentAbsences []CurrentAbsence

	// Time range: now to 30 days in the future
	now := time.Now()
	futureDate := now.AddDate(0, 0, 30)

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Check if absence is in the current/future range
			if (absence.StartDate.After(now) || absence.StartDate.Equal(now)) && absence.StartDate.Before(futureDate) ||
				(absence.EndDate.After(now) && absence.EndDate.Before(futureDate)) {

				// Apply filter if specified
				if !filter.StartDate.IsZero() && absence.EndDate.Before(filter.StartDate) {
					continue
				}
				if !filter.EndDate.IsZero() && absence.StartDate.After(filter.EndDate) {
					continue
				}

				// Find affected projects
				var affectedProjects []string
				for _, proj := range emp.ProjectAssignments {
					// Check if project is active during the absence
					if (proj.StartDate.Before(absence.EndDate) || proj.StartDate.Equal(absence.EndDate)) &&
						(proj.EndDate.After(absence.StartDate) || proj.EndDate.Equal(absence.StartDate) || proj.EndDate.IsZero()) {
						affectedProjects = append(affectedProjects, proj.ProjectName)
					}
				}

				currentAbsences = append(currentAbsences, CurrentAbsence{
					ID:               absence.ID.Hex(),
					EmployeeID:       emp.ID.Hex(),
					EmployeeName:     emp.FirstName + " " + emp.LastName,
					Type:             absence.Type,
					StartDate:        absence.StartDate,
					EndDate:          absence.EndDate,
					Days:             absence.Days,
					Status:           absence.Status,
					HasProfileImage:  len(emp.ProfileImageData.Data) > 0,
					AffectedProjects: affectedProjects,
				})
			}
		}
	}

	// Sort by start date
	sort.Slice(currentAbsences, func(i, j int) bool {
		return currentAbsences[i].StartDate.Before(currentAbsences[j].StartDate)
	})

	return currentAbsences
}
