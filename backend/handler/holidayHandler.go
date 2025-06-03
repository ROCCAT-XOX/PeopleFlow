package handler

import (
	"net/http"
	"strconv"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"

	"github.com/gin-gonic/gin"
)

// HolidayHandler verwaltet alle Anfragen zu Feiertagen
type HolidayHandler struct {
	holidayService *service.HolidayService
	settingsRepo   *repository.SystemSettingsRepository
}

// NewHolidayHandler erstellt einen neuen HolidayHandler
func NewHolidayHandler() *HolidayHandler {
	return &HolidayHandler{
		holidayService: service.NewHolidayService(),
		settingsRepo:   repository.NewSystemSettingsRepository(),
	}
}

// GetHolidays gibt alle Feiertage für ein Jahr und Bundesland zurück
func (h *HolidayHandler) GetHolidays(c *gin.Context) {
	// Parameter auslesen
	yearParam := c.Query("year")
	stateParam := c.Query("state")

	// Jahr parsen oder aktuelles Jahr verwenden
	year := time.Now().Year()
	if yearParam != "" {
		if parsedYear, err := strconv.Atoi(yearParam); err == nil {
			year = parsedYear
		}
	}

	// Bundesland bestimmen
	var state model.GermanState
	if stateParam != "" {
		state = model.GermanState(stateParam)
	} else {
		// Standard-Bundesland aus den Einstellungen holen
		settings, err := h.settingsRepo.GetSettings()
		if err == nil {
			state = model.GermanState(settings.State)
		} else {
			state = model.StateNordrheinWestfalen // Fallback
		}
	}

	// Feiertage abrufen
	holidays := h.holidayService.GetHolidaysForState(year, state)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"year":      year,
		"state":     string(state),
		"stateName": state.GetDisplayName(),
		"holidays":  holidays,
		"count":     len(holidays),
	})
}

// CheckHoliday prüft, ob ein bestimmtes Datum ein Feiertag ist
func (h *HolidayHandler) CheckHoliday(c *gin.Context) {
	// Parameter auslesen
	dateParam := c.Query("date") // Format: 2024-12-25
	stateParam := c.Query("state")

	// Datum parsen
	date, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültiges Datumsformat. Verwenden Sie YYYY-MM-DD",
		})
		return
	}

	// Bundesland bestimmen
	var state model.GermanState
	if stateParam != "" {
		state = model.GermanState(stateParam)
	} else {
		// Standard-Bundesland aus den Einstellungen holen
		settings, err := h.settingsRepo.GetSettings()
		if err == nil {
			state = model.GermanState(settings.State)
		} else {
			state = model.StateNordrheinWestfalen // Fallback
		}
	}

	// Prüfen ob Feiertag
	isHoliday := h.holidayService.IsHoliday(date, state)
	holidayName := h.holidayService.GetHolidayName(date, state)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"date":        date.Format("2006-01-02"),
		"state":       string(state),
		"stateName":   state.GetDisplayName(),
		"isHoliday":   isHoliday,
		"holidayName": holidayName,
	})
}

// GetWorkingDays berechnet die Arbeitstage in einem Zeitraum
func (h *HolidayHandler) GetWorkingDays(c *gin.Context) {
	// Parameter auslesen
	startDateParam := c.Query("startDate")
	endDateParam := c.Query("endDate")
	stateParam := c.Query("state")

	// Daten parsen
	startDate, err := time.Parse("2006-01-02", startDateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültiges Startdatum. Verwenden Sie YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültiges Enddatum. Verwenden Sie YYYY-MM-DD",
		})
		return
	}

	// Bundesland bestimmen
	var state model.GermanState
	if stateParam != "" {
		state = model.GermanState(stateParam)
	} else {
		// Standard-Bundesland aus den Einstellungen holen
		settings, err := h.settingsRepo.GetSettings()
		if err == nil {
			state = model.GermanState(settings.State)
		} else {
			state = model.StateNordrheinWestfalen // Fallback
		}
	}

	// Arbeitstage berechnen
	workingDays := h.holidayService.GetWorkingDaysBetween(startDate, endDate, state)

	// Zusätzliche Informationen
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1
	weekends := 0
	holidays := 0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			weekends++
		} else if h.holidayService.IsHoliday(d, state) {
			holidays++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"startDate":   startDate.Format("2006-01-02"),
		"endDate":     endDate.Format("2006-01-02"),
		"state":       string(state),
		"stateName":   state.GetDisplayName(),
		"totalDays":   totalDays,
		"workingDays": workingDays,
		"weekends":    weekends,
		"holidays":    holidays,
	})
}

// GetCurrentYearHolidays gibt alle Feiertage für das aktuelle Jahr im eingestellten Bundesland zurück
func (h *HolidayHandler) GetCurrentYearHolidays(c *gin.Context) {
	// Aktuelles Jahr
	year := time.Now().Year()

	// Bundesland aus den Einstellungen holen
	settings, err := h.settingsRepo.GetSettings()
	state := model.StateNordrheinWestfalen // Fallback
	if err == nil {
		state = model.GermanState(settings.State)
	}

	// Feiertage abrufen
	holidays := h.holidayService.GetHolidaysForState(year, state)

	// Für das Frontend aufbereiten
	var holidayList []gin.H
	for _, holiday := range holidays {
		holidayList = append(holidayList, gin.H{
			"name":    holiday.Name,
			"date":    holiday.Date.Format("02.01.2006"),
			"dateISO": holiday.Date.Format("2006-01-02"),
			"weekday": getGermanWeekday(holiday.Date.Weekday()),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"year":      year,
		"state":     string(state),
		"stateName": state.GetDisplayName(),
		"holidays":  holidayList,
		"count":     len(holidays),
	})
}

// Hilfsfunktion für deutsche Wochentage
func getGermanWeekday(weekday time.Weekday) string {
	weekdays := map[time.Weekday]string{
		time.Monday:    "Montag",
		time.Tuesday:   "Dienstag",
		time.Wednesday: "Mittwoch",
		time.Thursday:  "Donnerstag",
		time.Friday:    "Freitag",
		time.Saturday:  "Samstag",
		time.Sunday:    "Sonntag",
	}
	return weekdays[weekday]
}
