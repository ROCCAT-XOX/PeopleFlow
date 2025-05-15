package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PlanningHandler verwaltet alle Anfragen zur Projektplanung
type PlanningHandler struct {
	employeeRepo *repository.EmployeeRepository
}

// NewPlanningHandler erstellt einen neuen PlanningHandler
func NewPlanningHandler() *PlanningHandler {
	return &PlanningHandler{
		employeeRepo: repository.NewEmployeeRepository(),
	}
}

// ProjectData repräsentiert die Daten eines Projekts für die Planungsansicht
type ProjectData struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employeeId"`
	EmployeeName string    `json:"employeeName"`
	ProjectName  string    `json:"projectName"`
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	Role         string    `json:"role"`
	Source       string    `json:"source"`
}

// GetProjectPlanningView zeigt die Projektplanungsansicht an
func (h *PlanningHandler) GetProjectPlanningView(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Monat und Jahr aus der Anfrage holen oder aktuelles Datum verwenden
	now := time.Now()
	yearStr := c.DefaultQuery("year", now.Format("2006"))
	monthStr := c.DefaultQuery("month", now.Format("01"))

	// String in Integer konvertieren
	year, _ := strconv.Atoi(yearStr)
	month, _ := strconv.Atoi(monthStr)

	// Ersten Tag des Monats bestimmen
	currentMonthDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)

	// Letzten Tag des Monats bestimmen
	lastDay := currentMonthDate.AddDate(0, 1, -1)

	// Vorigen und nächsten Monat für Navigation bestimmen
	prevMonth := currentMonthDate.AddDate(0, -1, 0)
	nextMonth := currentMonthDate.AddDate(0, 1, 0)

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

	// Projekte sammeln
	var allProjects []ProjectData

	for _, emp := range employees {
		for _, proj := range emp.ProjectAssignments {
			allProjects = append(allProjects, ProjectData{
				ID:           proj.ID.Hex(),
				EmployeeID:   emp.ID.Hex(),
				EmployeeName: emp.FirstName + " " + emp.LastName,
				ProjectName:  proj.ProjectName,
				StartDate:    proj.StartDate,
				EndDate:      proj.EndDate,
				Role:         proj.Role,
				Source:       proj.Source,
			})
		}
	}

	// Sortieren nach Startdatum
	sort.Slice(allProjects, func(i, j int) bool {
		return allProjects[i].StartDate.Before(allProjects[j].StartDate)
	})

	// Kalenderwochen generieren
	var calendarWeeks [][]time.Time
	var currentWeek []time.Time

	// Wochentag des ersten Tags im Monat (0 = Sonntag, 1 = Montag, ...)
	firstWeekday := int(currentMonthDate.Weekday())
	if firstWeekday == 0 {
		firstWeekday = 7 // Sonntag als 7 betrachten (europäischer Kalender)
	}
	firstWeekday-- // Zur 0-basierten Indexierung wechseln (0 = Montag)

	// Tage vor dem ersten Tag des Monats auffüllen
	for i := 0; i < firstWeekday; i++ {
		prevDate := currentMonthDate.AddDate(0, 0, -firstWeekday+i)
		currentWeek = append(currentWeek, prevDate)
	}

	// Tage des Monats hinzufügen
	for day := 1; day <= lastDay.Day(); day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		currentWeek = append(currentWeek, date)

		// Wenn Sonntag oder letzter Tag, Woche abschließen
		if len(currentWeek) == 7 {
			calendarWeeks = append(calendarWeeks, currentWeek)
			currentWeek = []time.Time{}
		}
	}

	// Restliche Tage nach dem Monatsende auffüllen
	if len(currentWeek) > 0 {
		for i := len(currentWeek); i < 7; i++ {
			nextDate := lastDay.AddDate(0, 0, i-len(currentWeek)+1)
			currentWeek = append(currentWeek, nextDate)
		}
		calendarWeeks = append(calendarWeeks, currentWeek)
	}

	// Zusammenfassen aller Daten für die Anzeige
	c.HTML(http.StatusOK, "project_planning.html", gin.H{
		"title":            "Projektplanung",
		"active":           "planning",
		"user":             userModel.FirstName + " " + userModel.LastName,
		"email":            userModel.Email,
		"year":             time.Now().Year(),
		"userRole":         userRole,
		"projects":         allProjects,
		"calendarWeeks":    calendarWeeks,
		"currentYear":      currentMonthDate.Year(),
		"currentMonth":     currentMonthDate,
		"currentMonthName": currentMonthDate.Format("January"),
		"prevMonth":        prevMonth,
		"nextMonth":        nextMonth,
		"today":            now,
	})
}
