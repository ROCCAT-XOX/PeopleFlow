// backend/service/holidayService.go
package service

import (
	"PeopleFlow/backend/model"
	"time"
)

// Holiday repräsentiert einen Feiertag
type Holiday struct {
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`
	States []string  `json:"states"` // Bundesländer, in denen der Feiertag gilt
}

// HolidayService verwaltet deutsche Feiertage
type HolidayService struct{}

// NewHolidayService erstellt einen neuen HolidayService
func NewHolidayService() *HolidayService {
	return &HolidayService{}
}

// GetHolidaysForYear gibt alle Feiertage für ein bestimmtes Jahr zurück
func (s *HolidayService) GetHolidaysForYear(year int) []Holiday {
	var holidays []Holiday

	// Feste Feiertage (jedes Jahr gleich)
	holidays = append(holidays, s.getFixedHolidays(year)...)

	// Bewegliche Feiertage (basierend auf Ostern)
	easterDate := s.calculateEaster(year)
	holidays = append(holidays, s.getMovableHolidays(year, easterDate)...)

	return holidays
}

// GetHolidaysForState gibt alle Feiertage für ein bestimmtes Jahr und Bundesland zurück
func (s *HolidayService) GetHolidaysForState(year int, state model.GermanState) []Holiday {
	allHolidays := s.GetHolidaysForYear(year)
	var stateHolidays []Holiday

	stateCode := string(state)

	for _, holiday := range allHolidays {
		// Prüfen ob der Feiertag in diesem Bundesland gilt
		for _, holidayState := range holiday.States {
			if holidayState == "ALL" || holidayState == stateCode {
				stateHolidays = append(stateHolidays, holiday)
				break
			}
		}
	}

	return stateHolidays
}

// IsHoliday prüft, ob ein bestimmtes Datum ein Feiertag in einem Bundesland ist
func (s *HolidayService) IsHoliday(date time.Time, state model.GermanState) bool {
	holidays := s.GetHolidaysForState(date.Year(), state)

	for _, holiday := range holidays {
		if holiday.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			return true
		}
	}

	return false
}

// GetHolidayName gibt den Namen des Feiertags zurück, falls das Datum ein Feiertag ist
func (s *HolidayService) GetHolidayName(date time.Time, state model.GermanState) string {
	holidays := s.GetHolidaysForState(date.Year(), state)

	for _, holiday := range holidays {
		if holiday.Date.Format("2006-01-02") == date.Format("2006-01-02") {
			return holiday.Name
		}
	}

	return ""
}

// getFixedHolidays gibt alle festen Feiertage für ein Jahr zurück
func (s *HolidayService) getFixedHolidays(year int) []Holiday {
	return []Holiday{
		{
			Name:   "Neujahr",
			Date:   time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC),
			States: []string{"ALL"},
		},
		{
			Name:   "Heilige Drei Könige",
			Date:   time.Date(year, 1, 6, 0, 0, 0, 0, time.UTC),
			States: []string{"BW", "BY", "ST"},
		},
		{
			Name:   "Tag der Arbeit",
			Date:   time.Date(year, 5, 1, 0, 0, 0, 0, time.UTC),
			States: []string{"ALL"},
		},
		{
			Name:   "Augsburger Friedensfest",
			Date:   time.Date(year, 8, 8, 0, 0, 0, 0, time.UTC),
			States: []string{"BY"}, // Nur in Augsburg, aber wir vereinfachen auf ganz Bayern
		},
		{
			Name:   "Mariä Himmelfahrt",
			Date:   time.Date(year, 8, 15, 0, 0, 0, 0, time.UTC),
			States: []string{"BY", "SL"},
		},
		{
			Name:   "Tag der Deutschen Einheit",
			Date:   time.Date(year, 10, 3, 0, 0, 0, 0, time.UTC),
			States: []string{"ALL"},
		},
		{
			Name:   "Reformationstag",
			Date:   time.Date(year, 10, 31, 0, 0, 0, 0, time.UTC),
			States: []string{"BB", "MV", "SN", "ST", "TH", "HB", "HH", "NI", "SH"},
		},
		{
			Name:   "Allerheiligen",
			Date:   time.Date(year, 11, 1, 0, 0, 0, 0, time.UTC),
			States: []string{"BW", "BY", "NW", "RP", "SL"},
		},
		{
			Name:   "Buß- und Bettag",
			Date:   s.calculateBussUndBettag(year),
			States: []string{"SN"},
		},
		{
			Name:   "1. Weihnachtsfeiertag",
			Date:   time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC),
			States: []string{"ALL"},
		},
		{
			Name:   "2. Weihnachtsfeiertag",
			Date:   time.Date(year, 12, 26, 0, 0, 0, 0, time.UTC),
			States: []string{"ALL"},
		},
	}
}

// getMovableHolidays gibt alle beweglichen Feiertage basierend auf Ostern zurück
func (s *HolidayService) getMovableHolidays(year int, easter time.Time) []Holiday {
	return []Holiday{
		{
			Name:   "Karfreitag",
			Date:   easter.AddDate(0, 0, -2),
			States: []string{"ALL"},
		},
		{
			Name:   "Ostermontag",
			Date:   easter.AddDate(0, 0, 1),
			States: []string{"ALL"},
		},
		{
			Name:   "Christi Himmelfahrt",
			Date:   easter.AddDate(0, 0, 39),
			States: []string{"ALL"},
		},
		{
			Name:   "Pfingstmontag",
			Date:   easter.AddDate(0, 0, 50),
			States: []string{"ALL"},
		},
		{
			Name:   "Fronleichnam",
			Date:   easter.AddDate(0, 0, 60),
			States: []string{"BW", "BY", "HE", "NW", "RP", "SL", "SN", "TH"},
		},
	}
}

// calculateEaster berechnet das Osterdatum für ein Jahr (Gregorianischer Kalender)
func (s *HolidayService) calculateEaster(year int) time.Time {
	// Algorithmus von Carl Friedrich Gauß
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	n := (h + l - 7*m + 114) / 31
	p := (h + l - 7*m + 114) % 31

	return time.Date(year, time.Month(n), p+1, 0, 0, 0, 0, time.UTC)
}

// calculateBussUndBettag berechnet den Buß- und Bettag (letzter Mittwoch vor dem 1. Advent)
func (s *HolidayService) calculateBussUndBettag(year int) time.Time {
	// 1. Advent ist der 4. Sonntag vor dem 1. Weihnachtsfeiertag
	christmas := time.Date(year, 12, 25, 0, 0, 0, 0, time.UTC)

	// Finde den ersten Sonntag vor oder am 25.12.
	daysToSunday := int(christmas.Weekday())
	if daysToSunday == 0 {
		daysToSunday = 7
	}
	firstSunday := christmas.AddDate(0, 0, -daysToSunday)

	// 1. Advent ist 3 Wochen davor
	firstAdvent := firstSunday.AddDate(0, 0, -21)

	// Buß- und Bettag ist der Mittwoch davor (11 Tage)
	bussUndBettag := firstAdvent.AddDate(0, 0, -11)

	return bussUndBettag
}

// GetWorkingDaysInMonth gibt die Anzahl der Arbeitstage in einem Monat zurück
// (ohne Wochenenden und Feiertage)
func (s *HolidayService) GetWorkingDaysInMonth(year int, month time.Month, state model.GermanState) int {
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	holidays := s.GetHolidaysForState(year, state)
	holidayMap := make(map[string]bool)
	for _, holiday := range holidays {
		holidayMap[holiday.Date.Format("2006-01-02")] = true
	}

	workingDays := 0
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		// Überspringe Wochenenden
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}

		// Überspringe Feiertage
		if holidayMap[d.Format("2006-01-02")] {
			continue
		}

		workingDays++
	}

	return workingDays
}

// GetWorkingDaysBetween gibt die Anzahl der Arbeitstage zwischen zwei Daten zurück
func (s *HolidayService) GetWorkingDaysBetween(startDate, endDate time.Time, state model.GermanState) int {
	if startDate.After(endDate) {
		return 0
	}

	// Hole alle Feiertage für die betroffenen Jahre
	startYear := startDate.Year()
	endYear := endDate.Year()

	holidayMap := make(map[string]bool)
	for year := startYear; year <= endYear; year++ {
		holidays := s.GetHolidaysForState(year, state)
		for _, holiday := range holidays {
			holidayMap[holiday.Date.Format("2006-01-02")] = true
		}
	}

	workingDays := 0
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Überspringe Wochenenden
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}

		// Überspringe Feiertage
		if holidayMap[d.Format("2006-01-02")] {
			continue
		}

		workingDays++
	}

	return workingDays
}
