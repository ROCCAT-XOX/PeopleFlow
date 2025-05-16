package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
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
	projects := make(map[string]struct {
		ID   string
		Name string
	})
	for _, emp := range employees {
		for _, proj := range emp.ProjectAssignments {
			projects[proj.ProjectID] = struct {
				ID   string
				Name string
			}{
				ID:   proj.ProjectID,
				Name: proj.ProjectName,
			}
		}
	}

	// Für die Dropdown-Liste sortiert aufbereiten
	var projectsList []gin.H
	for _, p := range projects {
		projectsList = append(projectsList, gin.H{
			"ID":   p.ID,
			"Name": p.Name,
		})
	}

	sort.Slice(projectsList, func(i, j int) bool {
		return projectsList[i]["Name"].(string) < projectsList[j]["Name"].(string)
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

	// Beispielhafte Produktivitätsrate
	productivityRate := 85.3 // 85.3%

	// Beispielhafte Mitarbeiter-Produktivitätsdaten
	var productivityRanking []gin.H
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

			// Trend formatieren für die Anzeige
			trendFormatted := fmt.Sprintf("%.1f", trend)
			if trend > 0 {
				trendFormatted = "+" + trendFormatted
			}
			trendFormatted += "%"

			productivityRanking = append(productivityRanking, gin.H{
				"ID":               emp.ID.Hex(),
				"Name":             emp.FirstName + " " + emp.LastName,
				"Department":       string(emp.Department),
				"Hours":            empHours,
				"ProductivityRate": prodRate,
				"Trend":            trend,
				"TrendFormatted":   trendFormatted,
				"IsTrendPositive":  trend > 0,
				"IsTrendNegative":  trend < 0,
				"HasProfileImage":  len(emp.ProfileImageData.Data) > 0,
			})
		}
	}

	// Nach Produktivitätsrate sortieren (absteigend)
	sort.Slice(productivityRanking, func(i, j int) bool {
		return productivityRanking[i]["ProductivityRate"].(float64) > productivityRanking[j]["ProductivityRate"].(float64)
	})

	// Beispielhafte Projektdaten - mit vorberechneten Feldern
	projectDetails := []gin.H{
		{
			"ID":                  "proj-1",
			"Name":                "Website Redesign",
			"Status":              "In Arbeit",
			"TeamSize":            5,
			"Hours":               320.0,
			"HoursFormatted":      "320.0 Std",
			"Efficiency":          87.2,
			"EfficiencyFormatted": "87.2%",
			"EfficiencyClass":     "bg-green-600",
		},
		{
			"ID":                  "proj-2",
			"Name":                "Mobile App Entwicklung",
			"Status":              "In Arbeit",
			"TeamSize":            8,
			"Hours":               780.0,
			"HoursFormatted":      "780.0 Std",
			"Efficiency":          92.5,
			"EfficiencyFormatted": "92.5%",
			"EfficiencyClass":     "bg-green-600",
		},
		{
			"ID":                  "proj-3",
			"Name":                "Datenmigration",
			"Status":              "Abgeschlossen",
			"TeamSize":            3,
			"Hours":               150.0,
			"HoursFormatted":      "150.0 Std",
			"Efficiency":          89.0,
			"EfficiencyFormatted": "89.0%",
			"EfficiencyClass":     "bg-green-600",
		},
		{
			"ID":                  "proj-4",
			"Name":                "Security Audit",
			"Status":              "Kritisch",
			"TeamSize":            2,
			"Hours":               95.0,
			"HoursFormatted":      "95.0 Std",
			"Efficiency":          65.0,
			"EfficiencyFormatted": "65.0%",
			"EfficiencyClass":     "bg-yellow-600",
		},
		{
			"ID":                  "proj-5",
			"Name":                "CRM Implementation",
			"Status":              "Geplant",
			"TeamSize":            6,
			"Hours":               80.0,
			"HoursFormatted":      "80.0 Std",
			"Efficiency":          90.0,
			"EfficiencyFormatted": "90.0%",
			"EfficiencyClass":     "bg-green-600",
		},
	}

	// Aktuelle Abwesenheiten identifizieren (für die nächsten 30 Tage)
	now := time.Now()
	endPeriod := now.AddDate(0, 0, 30)
	var currentAbsences []gin.H

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Prüfen, ob Abwesenheit im aktuellen oder zukünftigen Zeitraum liegt
			if (absence.StartDate.After(now) || absence.StartDate.Equal(now)) &&
				absence.StartDate.Before(endPeriod) ||
				(absence.EndDate.After(now) && absence.EndDate.Before(endPeriod)) {

				// Betroffene Projekte identifizieren
				var affectedProjects []string
				for _, proj := range emp.ProjectAssignments {
					// Prüfen, ob Projekt während der Abwesenheit aktiv ist
					if (proj.StartDate.Before(absence.EndDate) || proj.StartDate.Equal(absence.EndDate)) &&
						(proj.EndDate.After(absence.StartDate) || proj.EndDate.Equal(absence.StartDate)) {
						affectedProjects = append(affectedProjects, proj.ProjectName)
					}
				}

				currentAbsences = append(currentAbsences, gin.H{
					"ID":               absence.ID.Hex(),
					"EmployeeID":       emp.ID.Hex(),
					"EmployeeName":     emp.FirstName + " " + emp.LastName,
					"Type":             absence.Type,
					"StartDate":        absence.StartDate,
					"EndDate":          absence.EndDate,
					"Days":             absence.Days,
					"Status":           absence.Status,
					"HasProfileImage":  len(emp.ProfileImageData.Data) > 0,
					"AffectedProjects": affectedProjects,
				})
			}
		}
	}

	// Nach Startdatum sortieren
	sort.Slice(currentAbsences, func(i, j int) bool {
		return currentAbsences[i]["StartDate"].(time.Time).Before(currentAbsences[j]["StartDate"].(time.Time))
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
