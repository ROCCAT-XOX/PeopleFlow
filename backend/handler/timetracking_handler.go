package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeTrackingHandler erweitern (bestehende Struktur ersetzen)
type TimeTrackingHandler struct {
	employeeRepo       *repository.EmployeeRepository
	timeAccountService *service.TimeAccountService
}

// NewTimeTrackingHandler korrigieren (bestehende Funktion ersetzen)
func NewTimeTrackingHandler() *TimeTrackingHandler {
	return &TimeTrackingHandler{
		employeeRepo:       repository.NewEmployeeRepository(),
		timeAccountService: service.NewTimeAccountService(),
	}
}

// Neue Struktur für erweiterte Mitarbeiter-Zusammenfassung hinzufügen
type EmployeeSummaryWithOvertime struct {
	EmployeeSummary                                  // Eingebettete bestehende Struktur
	OvertimeBalance float64                          `json:"overtimeBalance"`
	OvertimeStatus  string                           `json:"overtimeStatus"`
	WeeklyTarget    float64                          `json:"weeklyTarget"`
	OvertimeSummary *service.EmployeeOvertimeSummary `json:"overtimeSummary"`
}

// TimeEntryViewModel repräsentiert die Daten für die Darstellung eines Zeiteintrags
type TimeEntryViewModel struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employeeId"`
	EmployeeName string    `json:"employeeName"`
	Date         time.Time `json:"date"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Duration     float64   `json:"duration"`
	ProjectID    string    `json:"projectId"`
	ProjectName  string    `json:"projectName"`
	Activity     string    `json:"activity"`
	Description  string    `json:"description"`
	Source       string    `json:"source"`
}

// ProjectViewModel repräsentiert ein Projekt für die Auswahl
type ProjectViewModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EmployeeSummary repräsentiert die zusammengefassten Zeitdaten für einen Mitarbeiter
type EmployeeSummary struct {
	EmployeeID      string               `json:"employeeId"`
	EmployeeName    string               `json:"employeeName"`
	HasProfileImage bool                 `json:"hasProfileImage"`
	TotalHours      float64              `json:"totalHours"`
	ProjectCount    int                  `json:"projectCount"`
	Projects        []ProjectSummary     `json:"projects"`
	TimeEntries     []TimeEntryViewModel `json:"timeEntries"`
}

// ProjectSummary repräsentiert die zusammengefassten Stunden pro Projekt
type ProjectSummary struct {
	ProjectID   string  `json:"projectId"`
	ProjectName string  `json:"projectName"`
	Hours       float64 `json:"hours"`
}

// GetTimeTrackingView zeigt die Zeiterfassungsübersicht an
func (h *TimeTrackingHandler) GetTimeTrackingView(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Alle Mitarbeiter abrufen
	employees, _, err := h.employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Mitarbeiter: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Zeiteinträge pro Mitarbeiter sammeln und zusammenfassen
	var employeeSummaries []EmployeeSummary
	var totalHours float64
	projects := make(map[string]ProjectViewModel)

	for _, emp := range employees {
		if len(emp.TimeEntries) == 0 {
			continue
		}

		summary := EmployeeSummary{
			EmployeeID:      emp.ID.Hex(),
			EmployeeName:    emp.FirstName + " " + emp.LastName,
			HasProfileImage: len(emp.ProfileImageData.Data) > 0,
			TimeEntries:     []TimeEntryViewModel{},
			Projects:        []ProjectSummary{}, // Explizit initialisieren
		}

		// Projekte und Stunden pro Mitarbeiter sammeln
		projectHours := make(map[string]float64)
		projectNames := make(map[string]string)

		for _, entry := range emp.TimeEntries {
			// Zeiteintrag zum ViewModel hinzufügen
			timeEntryVM := TimeEntryViewModel{
				ID:           entry.ID.Hex(),
				EmployeeID:   emp.ID.Hex(),
				EmployeeName: emp.FirstName + " " + emp.LastName,
				Date:         entry.Date,
				StartTime:    entry.StartTime,
				EndTime:      entry.EndTime,
				Duration:     entry.Duration,
				ProjectID:    entry.ProjectID,
				ProjectName:  entry.ProjectName,
				Activity:     entry.Activity,
				Description:  entry.Description, // NEU: Description hinzufügen
				Source:       entry.Source,
			}

			summary.TimeEntries = append(summary.TimeEntries, timeEntryVM)
			summary.TotalHours += entry.Duration

			// Projekt hinzufügen, wenn es noch nicht in der Map ist
			if entry.ProjectID != "" {
				projects[entry.ProjectID] = ProjectViewModel{
					ID:   entry.ProjectID,
					Name: entry.ProjectName,
				}

				// Stunden pro Projekt summieren
				projectHours[entry.ProjectID] += entry.Duration
				projectNames[entry.ProjectID] = entry.ProjectName
			}
		}

		// Projekte für diesen Mitarbeiter aufbereiten
		for projID, hours := range projectHours {
			if projID != "" {
				summary.Projects = append(summary.Projects, ProjectSummary{
					ProjectID:   projID,
					ProjectName: projectNames[projID],
					Hours:       hours,
				})
			}
		}

		summary.ProjectCount = len(summary.Projects)
		totalHours += summary.TotalHours

		// Nach Datum sortieren
		sort.Slice(summary.TimeEntries, func(i, j int) bool {
			return summary.TimeEntries[i].Date.After(summary.TimeEntries[j].Date)
		})

		employeeSummaries = append(employeeSummaries, summary)
	}

	// Nach Mitarbeiternamen sortieren
	sort.Slice(employeeSummaries, func(i, j int) bool {
		return employeeSummaries[i].EmployeeName < employeeSummaries[j].EmployeeName
	})

	// Projekte für Dropdown aufbereiten
	var projectsList []ProjectViewModel
	for _, p := range projects {
		projectsList = append(projectsList, p)
	}
	// Nach Namen sortieren
	sort.Slice(projectsList, func(i, j int) bool {
		return projectsList[i].Name < projectsList[j].Name
	})

	// Überstunden für alle Mitarbeiter berechnen
	var employeeSummariesWithOvertime []EmployeeSummaryWithOvertime
	for _, summary := range employeeSummaries {
		emp, err := h.employeeRepo.FindByID(summary.EmployeeID)
		if err != nil {
			// Falls Fehler beim Abrufen, verwende Standard-Werte
			enhancedSummary := EmployeeSummaryWithOvertime{
				EmployeeSummary: summary,
				OvertimeBalance: 0.0,
				OvertimeStatus:  "neutral",
				WeeklyTarget:    40.0,
				OvertimeSummary: nil,
			}
			employeeSummariesWithOvertime = append(employeeSummariesWithOvertime, enhancedSummary)
			continue
		}

		// Überstunden berechnen
		overtimeSummary, err := h.timeAccountService.GetEmployeeOvertimeSummary(summary.EmployeeID)
		if err != nil {
			overtimeSummary = nil
		}

		enhancedSummary := EmployeeSummaryWithOvertime{
			EmployeeSummary: summary,
			OvertimeBalance: emp.OvertimeBalance,
			OvertimeStatus:  emp.GetOvertimeStatus(),
			WeeklyTarget:    emp.GetWeeklyTargetHours(),
			OvertimeSummary: overtimeSummary,
		}

		employeeSummariesWithOvertime = append(employeeSummariesWithOvertime, enhancedSummary)
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "timetracking.html", gin.H{
		"title":                       "Zeiterfassung",
		"active":                      "timetracking",
		"user":                        userModel.FirstName + " " + userModel.LastName,
		"email":                       userModel.Email,
		"year":                        time.Now().Year(),
		"userRole":                    userRole,
		"employees":                   employees,
		"projects":                    projectsList,
		"employeeSummary":             employeeSummaries,
		"employeeSummaryWithOvertime": employeeSummariesWithOvertime,
		"totalHours":                  totalHours,
		"totalEmployees":              len(employeeSummaries),
		"totalProjects":               len(projects),
	})
}

// Rest der Funktionen bleibt unverändert...
func (h *TimeTrackingHandler) GetEmployeeTimeEntries(c *gin.Context) {
	employeeID := c.Param("id")

	// Parameter für Filterung
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	projectID := c.Query("projectId")

	// Datumswerte parsen
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Startdatum"})
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Enddatum"})
			return
		}
	}

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Zeiteinträge filtern
	var filteredEntries []TimeEntryViewModel

	for _, entry := range employee.TimeEntries {
		// Datumsfilterung
		if !startDate.IsZero() && entry.Date.Before(startDate) {
			continue
		}
		if !endDate.IsZero() && entry.Date.After(endDate) {
			continue
		}

		// Projektfilterung
		if projectID != "" && entry.ProjectID != projectID {
			continue
		}

		// Zeiteintrag zum ViewModel hinzufügen
		timeEntryVM := TimeEntryViewModel{
			ID:           entry.ID.Hex(),
			EmployeeID:   employee.ID.Hex(),
			EmployeeName: employee.FirstName + " " + employee.LastName,
			Date:         entry.Date,
			StartTime:    entry.StartTime,
			EndTime:      entry.EndTime,
			Duration:     entry.Duration,
			ProjectID:    entry.ProjectID,
			ProjectName:  entry.ProjectName,
			Activity:     entry.Activity,
			Description:  entry.Description,
			Source:       entry.Source,
		}

		filteredEntries = append(filteredEntries, timeEntryVM)
	}

	// Nach Datum sortieren
	sort.Slice(filteredEntries, func(i, j int) bool {
		return filteredEntries[i].Date.After(filteredEntries[j].Date)
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filteredEntries,
	})
}

func (h *TimeTrackingHandler) ExportTimeTracking(c *gin.Context) {
	// Parameter für Filterung
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	projectID := c.Query("projectId")
	employeeIDs := c.QueryArray("employeeIds")

	// Datumswerte parsen
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Startdatum"})
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Enddatum"})
			return
		}
	}

	// Alle Mitarbeiter abrufen, wenn keine spezifischen IDs angegeben wurden
	var employees []*model.Employee
	if len(employeeIDs) == 0 {
		employees, _, err = h.employeeRepo.FindAll(0, 1000, "lastName", 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Mitarbeiter"})
			return
		}
	} else {
		// Nur die angegebenen Mitarbeiter abrufen
		for _, id := range employeeIDs {
			emp, err := h.employeeRepo.FindByID(id)
			if err == nil {
				employees = append(employees, emp)
			}
		}
	}

	// CSV-Header
	csvContent := "Mitarbeiter,Datum,Start,Ende,Dauer,Projekt,Tätigkeit\n"

	// Zeiteinträge filtern und sammeln
	for _, employee := range employees {
		for _, entry := range employee.TimeEntries {
			// Datumsfilterung
			if !startDate.IsZero() && entry.Date.Before(startDate) {
				continue
			}
			if !endDate.IsZero() && entry.Date.After(endDate) {
				continue
			}

			// Projektfilterung
			if projectID != "" && entry.ProjectID != projectID {
				continue
			}

			// CSV-Zeile hinzufügen
			csvContent += employee.FirstName + " " + employee.LastName + ","
			csvContent += entry.Date.Format("02.01.2006") + ","
			csvContent += entry.StartTime.Format("15:04") + ","
			csvContent += entry.EndTime.Format("15:04") + ","
			csvContent += fmt.Sprintf("%.2f", entry.Duration) + ","
			csvContent += entry.ProjectName + ","
			csvContent += entry.Activity + "\n"
		}
	}

	// CSV-Datei senden
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=zeiterfassung.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, csvContent)
}

func (h *TimeTrackingHandler) RecalculateOvertime(c *gin.Context) {
	err := h.timeAccountService.RecalculateAllEmployeeOvertimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der Überstunden-Berechnung: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Überstunden für alle Mitarbeiter wurden neu berechnet",
	})
}

func (h *TimeTrackingHandler) GetEmployeeOvertimeDetails(c *gin.Context) {
	employeeID := c.Param("id")

	overtimeSummary, err := h.timeAccountService.GetEmployeeOvertimeSummary(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden oder Fehler bei der Berechnung"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overtimeSummary,
	})
}
