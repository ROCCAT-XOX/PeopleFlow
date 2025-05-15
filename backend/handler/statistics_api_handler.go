package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"net/http"
	"sort"
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
