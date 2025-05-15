package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// StatisticsHandler verwaltet alle Anfragen zur Statistikseite
type StatisticsHandler struct {
	employeeRepo *repository.EmployeeRepository
}

// NewStatisticsHandler erstellt einen neuen StatisticsHandler
func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{
		employeeRepo: repository.NewEmployeeRepository(),
	}
}

// EmployeeProductivityData enthält Produktivitätsdaten für einen Mitarbeiter
type EmployeeProductivityData struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Department       string  `json:"department"`
	Hours            float64 `json:"hours"`
	ProductivityRate float64 `json:"productivityRate"`
	Trend            float64 `json:"trend"` // Prozentuale Veränderung zum Vormonat
	HasProfileImage  bool    `json:"hasProfileImage"`
}

// ProjectStatData enthält Statistikdaten für ein Projekt
type ProjectStatData struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	TeamSize      int     `json:"teamSize"`
	Hours         float64 `json:"hours"`
	TimeDeviation float64 `json:"timeDeviation"` // Prozentuale Abweichung zum Plan
	Efficiency    float64 `json:"efficiency"`    // Effizienzrate in Prozent
}

// AbsenceStatData enthält Daten für Abwesenheiten
type AbsenceStatData struct {
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

// GetStatisticsView zeigt die Statistikübersicht an
func (h *StatisticsHandler) GetStatisticsView(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Alle Mitarbeiter abrufen
	employees, err := h.employeeRepo.FindAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Mitarbeiter: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Projekte aus allen Mitarbeitern extrahieren und deduplizieren
	projects := make(map[string]ProjectViewModel)
	for _, emp := range employees {
		for _, proj := range emp.ProjectAssignments {
			projects[proj.ProjectID] = ProjectViewModel{
				ID:   proj.ProjectID,
				Name: proj.ProjectName,
			}
		}
	}

	// Für die Dropdown-Liste sortiert aufbereiten
	var projectsList []ProjectViewModel
	for _, p := range projects {
		projectsList = append(projectsList, p)
	}

	sort.Slice(projectsList, func(i, j int) bool {
		return projectsList[i].Name < projectsList[j].Name
	})

	// Berechnung der Gesamtarbeitszeit
	var totalHours float64
	for _, emp := range employees {
		for _, entry := range emp.TimeEntries {
			totalHours += entry.Duration
		}
	}

	// Berechnung der Abwesenheitstage
	var totalAbsenceDays float64
	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Nur genehmigte Abwesenheiten zählen
			if absence.Status == "approved" {
				totalAbsenceDays += absence.Days
			}
		}
	}

	// Beispielhafte Produktivitätsrate (in einer echten Anwendung würde dies
	// auf tatsächlichen Leistungsdaten, Zielvereinbarungen, etc. basieren)
	productivityRate := 85.3 // 85.3%

	// Beispielhafte Mitarbeiter-Produktivitätsdaten
	var productivityRanking []EmployeeProductivityData
	for _, emp := range employees {
		// Berechnung der Gesamtstunden für diesen Mitarbeiter
		var empHours float64
		for _, entry := range emp.TimeEntries {
			empHours += entry.Duration
		}

		// Nur Mitarbeiter mit Zeiteinträgen einbeziehen
		if empHours > 0 {
			// In einer echten Anwendung würde die Produktivitätsrate berechnet werden
			// basierend auf Qualität, Zeitmanagement, Zielerreichung, etc.
			// Hier verwenden wir einen zufälligen Wert zwischen 75 und 95%
			prodRate := 75.0 + float64(emp.ID.Hex()[0]%20) + float64(emp.ID.Hex()[1]%10)
			if prodRate > 95 {
				prodRate = 95
			}

			// Trend berechnen (Beispieldaten)
			trend := -5.0 + float64(emp.ID.Hex()[2]%15)

			productivityRanking = append(productivityRanking, EmployeeProductivityData{
				ID:               emp.ID.Hex(),
				Name:             emp.FirstName + " " + emp.LastName,
				Department:       string(emp.Department),
				Hours:            empHours,
				ProductivityRate: prodRate,
				Trend:            trend,
				HasProfileImage:  len(emp.ProfileImageData.Data) > 0,
			})
		}
	}

	// Nach Produktivitätsrate sortieren (absteigend)
	sort.Slice(productivityRanking, func(i, j int) bool {
		return productivityRanking[i].ProductivityRate > productivityRanking[j].ProductivityRate
	})

	// Beispielhafte Projektdaten
	projectDetails := []ProjectStatData{
		{
			ID:            "proj-1",
			Name:          "Website Redesign",
			Status:        "In Arbeit",
			TeamSize:      5,
			Hours:         320,
			TimeDeviation: 8.5,  // 8.5% über Plan
			Efficiency:    87.2, // 87.2% Effizienz
		},
		{
			ID:            "proj-2",
			Name:          "Mobile App Entwicklung",
			Status:        "In Arbeit",
			TeamSize:      8,
			Hours:         780,
			TimeDeviation: -3.2, // 3.2% unter Plan
			Efficiency:    92.5, // 92.5% Effizienz
		},
		{
			ID:            "proj-3",
			Name:          "Datenmigration",
			Status:        "Abgeschlossen",
			TeamSize:      3,
			Hours:         150,
			TimeDeviation: 5.0,  // 5% über Plan
			Efficiency:    89.0, // 89% Effizienz
		},
		{
			ID:            "proj-4",
			Name:          "Security Audit",
			Status:        "Kritisch",
			TeamSize:      2,
			Hours:         95,
			TimeDeviation: 15.0, // 15% über Plan
			Efficiency:    65.0, // 65% Effizienz
		},
		{
			ID:            "proj-5",
			Name:          "CRM Implementation",
			Status:        "Geplant",
			TeamSize:      6,
			Hours:         80,
			TimeDeviation: 0.0,  // Noch keine Abweichung
			Efficiency:    90.0, // Geschätzte Effizienz
		},
	}

	// Aktuelle Abwesenheiten identifizieren (für die nächsten 30 Tage)
	now := time.Now()
	endPeriod := now.AddDate(0, 0, 30)
	var currentAbsences []AbsenceStatData

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Prüfen, ob Abwesenheit im aktuellen oder zukünftigen Zeitraum liegt
			if (absence.StartDate.After(now) || absence.StartDate.Equal(now)) &&
				absence.StartDate.Before(endPeriod) ||
				(absence.EndDate.After(now) && absence.EndDate.Before(endPeriod)) {

				// Betroffene Projekte identifizieren (in einer echten Anwendung würde
				// dies auf tatsächlichen Zuordnungen basieren)
				var affectedProjects []string
				for _, proj := range emp.ProjectAssignments {
					// Prüfen, ob Projekt während der Abwesenheit aktiv ist
					if (proj.StartDate.Before(absence.EndDate) || proj.StartDate.Equal(absence.EndDate)) &&
						(proj.EndDate.After(absence.StartDate) || proj.EndDate.Equal(absence.StartDate)) {
						affectedProjects = append(affectedProjects, proj.ProjectName)
					}
				}

				currentAbsences = append(currentAbsences, AbsenceStatData{
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

	// Nach Startdatum sortieren
	sort.Slice(currentAbsences, func(i, j int) bool {
		return currentAbsences[i].StartDate.Before(currentAbsences[j].StartDate)
	})

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "statistics.html", gin.H{
		"title":               "Statistiken",
		"active":              "statistics",
		"user":                userModel.FirstName + " " + userModel.LastName,
		"email":               userModel.Email,
		"year":                time.Now().Year(),
		"userRole":            userRole,
		"employees":           employees,
		"projects":            projectsList,
		"totalHours":          totalHours,
		"productivityRate":    productivityRate,
		"totalAbsenceDays":    totalAbsenceDays,
		"productivityRanking": productivityRanking,
		"projectDetails":      projectDetails,
		"currentAbsences":     currentAbsences,
	})
}
