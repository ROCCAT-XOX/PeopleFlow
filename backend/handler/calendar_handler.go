// backend/handler/calendar_handler.go
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

// CalendarHandler verwaltet alle Anfragen zum Kalender und Abwesenheiten
type CalendarHandler struct {
	employeeRepo *repository.EmployeeRepository
}

// NewCalendarHandler erstellt einen neuen CalendarHandler
func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{
		employeeRepo: repository.NewEmployeeRepository(),
	}
}

// AbsenceData repräsentiert die Daten einer Abwesenheit für die Kalenderansicht
type AbsenceData struct {
	ID            string    `json:"id"`
	EmployeeID    string    `json:"employeeId"`
	EmployeeName  string    `json:"employeeName"`
	Type          string    `json:"type"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	Days          float64   `json:"days"`
	Status        string    `json:"status"`
	StatusDisplay string    `json:"statusDisplay"`
	TypeDisplay   string    `json:"typeDisplay"`
	TypeColor     string    `json:"typeColor"`
}

// GetAbsenceCalendar zeigt eine Kalenderansicht mit allen Abwesenheiten an
func (h *CalendarHandler) GetAbsenceCalendar(c *gin.Context) {
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

	// Abwesenheiten sammeln
	var allAbsences []AbsenceData

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			// Status-Anzeige festlegen
			statusDisplay := "Genehmigt"
			switch absence.Status {
			case "requested":
				statusDisplay = "Beantragt"
			case "rejected":
				statusDisplay = "Abgelehnt"
			case "cancelled":
				statusDisplay = "Storniert"
			}

			// Typ-Anzeige und Farbe festlegen
			typeDisplay := "Urlaub"
			typeColor := "bg-green-500"
			switch absence.Type {
			case "sick":
				typeDisplay = "Krankheit"
				typeColor = "bg-red-500"
			case "special":
				typeDisplay = "Sonderurlaub"
				typeColor = "bg-blue-500"
			}

			allAbsences = append(allAbsences, AbsenceData{
				ID:            absence.ID.Hex(),
				EmployeeID:    emp.ID.Hex(),
				EmployeeName:  emp.FirstName + " " + emp.LastName,
				Type:          absence.Type,
				StartDate:     absence.StartDate,
				EndDate:       absence.EndDate,
				Days:          absence.Days,
				Status:        absence.Status,
				StatusDisplay: statusDisplay,
				TypeDisplay:   typeDisplay,
				TypeColor:     typeColor,
			})
		}
	}

	// Sortieren nach Startdatum
	sort.Slice(allAbsences, func(i, j int) bool {
		return allAbsences[i].StartDate.Before(allAbsences[j].StartDate)
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
	c.HTML(http.StatusOK, "calendar.html", gin.H{
		"title":            "Abwesenheitskalender",
		"active":           "calendar",
		"user":             userModel.FirstName + " " + userModel.LastName,
		"email":            userModel.Email,
		"year":             time.Now().Year(),
		"userRole":         userRole,
		"absences":         allAbsences,
		"calendarWeeks":    calendarWeeks,
		"currentYear":      currentMonthDate.Year(),
		"currentMonth":     currentMonthDate,
		"currentMonthName": currentMonthDate.Format("January"),
		"prevMonth":        prevMonth,
		"nextMonth":        nextMonth,
		"today":            now,
	})
}
