// backend/service/hrService.go
package service

import (
	"PeopleFlow/backend/model"
	"sort"
	"time"
)

// HRService verwaltet HR-spezifische Berechnungen und Statistiken
type HRService struct{}

// NewHRService erstellt einen neuen HRService
func NewHRService() *HRService {
	return &HRService{}
}

// HRDashboardData enthält alle Daten für das HR-Dashboard
type HRDashboardData struct {
	// Grundstatistiken
	TotalEmployees        int
	ActiveEmployees       int
	OnLeaveEmployees      int
	SickEmployees         int // Geändert von RemoteEmployees zu SickEmployees
	InactiveEmployees     int
	NewEmployeesThisMonth int

	// Abwesenheitsstatistiken
	CurrentAbsences  int
	UpcomingAbsences int
	AbsenceRate      float64
	SickRate         float64

	// Gespräche und Reviews
	UpcomingReviews int
	OverdueReviews  int

	// Fluktuationsstatistiken
	TurnoverRate      float64
	AverageEmployment float64
	RetentionRate     float64

	// Chart-Daten
	StatusDistribution map[string]int
	DepartmentCounts   map[string]int
	MonthlyHires       []int
	MonthlyDepartures  []int
	AgeDistribution    map[string]int
	TenureDistribution map[string]int
	AbsenceByMonth     []int
	SicknessByMonth    []int
}

// CalculateHRDashboardData berechnet alle HR-relevanten Statistiken
func (h *HRService) CalculateHRDashboardData(employees []*model.Employee) *HRDashboardData {
	now := time.Now()
	data := &HRDashboardData{
		StatusDistribution: make(map[string]int),
		DepartmentCounts:   make(map[string]int),
		AgeDistribution:    make(map[string]int),
		TenureDistribution: make(map[string]int),
		MonthlyHires:       make([]int, 12),
		MonthlyDepartures:  make([]int, 12),
		AbsenceByMonth:     make([]int, 12),
		SicknessByMonth:    make([]int, 12),
	}

	// Grundstatistiken berechnen
	data.TotalEmployees = len(employees)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	var activeEmployees []*model.Employee

	for _, emp := range employees {
		// Status-Verteilung
		switch emp.Status {
		case model.EmployeeStatusActive:
			data.ActiveEmployees++
			data.StatusDistribution["Aktiv"]++
			activeEmployees = append(activeEmployees, emp)
		case model.EmployeeStatusOnLeave:
			data.OnLeaveEmployees++
			data.StatusDistribution["Im Urlaub"]++
		case model.EmployeeStatusRemote:
			// Remote wird jetzt als "Krank" behandelt
			data.SickEmployees++
			data.StatusDistribution["Krank"]++
		case model.EmployeeStatusInactive:
			data.InactiveEmployees++
			data.StatusDistribution["Inaktiv"]++
		}

		// Abteilungsverteilung
		if emp.Department != "" {
			data.DepartmentCounts[string(emp.Department)]++
		}

		// Neue Mitarbeiter diesen Monat
		if emp.HireDate.After(startOfMonth) {
			data.NewEmployeesThisMonth++
		}

		// Monatliche Einstellungen (für Chart)
		if emp.HireDate.After(startOfYear) {
			month := int(emp.HireDate.Month()) - 1
			if month >= 0 && month < 12 {
				data.MonthlyHires[month]++
			}
		}

		// Altersverteilung
		age := h.calculateAge(emp.DateOfBirth)
		ageGroup := h.getAgeGroup(age)
		data.AgeDistribution[ageGroup]++

		// Betriebszugehörigkeit
		tenure := h.calculateTenure(emp.HireDate)
		tenureGroup := h.getTenureGroup(tenure)
		data.TenureDistribution[tenureGroup]++
	}

	// Abwesenheitsstatistiken berechnen
	data.CurrentAbsences = data.OnLeaveEmployees + data.SickEmployees
	data.AbsenceRate = float64(data.OnLeaveEmployees) / float64(data.TotalEmployees) * 100
	data.SickRate = float64(data.SickEmployees) / float64(data.TotalEmployees) * 100

	// Anstehende Reviews berechnen
	data.UpcomingReviews = h.calculateUpcomingReviews(employees)
	data.OverdueReviews = h.calculateOverdueReviews(employees)

	// Fluktuationsrate berechnen
	data.TurnoverRate = h.calculateTurnoverRate(employees)
	data.AverageEmployment = h.calculateAverageEmployment(activeEmployees)
	data.RetentionRate = 100 - data.TurnoverRate

	// Monatliche Abgänge simulieren (vereinfacht)
	for i := 0; i < 12; i++ {
		data.MonthlyDepartures[i] = data.InactiveEmployees / 12
		data.AbsenceByMonth[i] = data.OnLeaveEmployees
		data.SicknessByMonth[i] = data.SickEmployees
	}

	return data
}

// calculateAge berechnet das Alter basierend auf dem Geburtsdatum
func (h *HRService) calculateAge(birthDate time.Time) int {
	if birthDate.IsZero() {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.Month() < birthDate.Month() || (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}
	return age
}

// getAgeGroup ordnet ein Alter einer Altersgruppe zu
func (h *HRService) getAgeGroup(age int) string {
	if age < 25 {
		return "< 25"
	} else if age < 35 {
		return "25-34"
	} else if age < 45 {
		return "35-44"
	} else if age < 55 {
		return "45-54"
	} else if age < 65 {
		return "55-64"
	}
	return "≥ 65"
}

// CalculateTenure berechnet die Betriebszugehörigkeit in Jahren (public)
func (h *HRService) CalculateTenure(hireDate time.Time) float64 {
	return h.calculateTenure(hireDate)
}

// calculateTenure berechnet die Betriebszugehörigkeit in Jahren (private)
func (h *HRService) calculateTenure(hireDate time.Time) float64 {
	return time.Since(hireDate).Hours() / (24 * 365.25)
}

// getTenureGroup ordnet eine Betriebszugehörigkeit einer Gruppe zu
func (h *HRService) getTenureGroup(tenure float64) string {
	if tenure < 1 {
		return "< 1 Jahr"
	} else if tenure < 3 {
		return "1-3 Jahre"
	} else if tenure < 5 {
		return "3-5 Jahre"
	} else if tenure < 10 {
		return "5-10 Jahre"
	}
	return "≥ 10 Jahre"
}

// calculateUpcomingReviews berechnet anstehende Mitarbeitergespräche
func (h *HRService) calculateUpcomingReviews(employees []*model.Employee) int {
	upcoming := 0

	for _, emp := range employees {
		if emp.Status != model.EmployeeStatusActive {
			continue
		}

		// Vereinfachte Logik: Reviews alle 12 Monate ab Einstellungsdatum
		monthsSinceHire := int(time.Since(emp.HireDate).Hours() / (24 * 30.44))
		if monthsSinceHire > 0 && monthsSinceHire%12 <= 1 {
			upcoming++
		}
	}

	return upcoming
}

// calculateOverdueReviews berechnet überfällige Mitarbeitergespräche
func (h *HRService) calculateOverdueReviews(employees []*model.Employee) int {
	overdue := 0

	for _, emp := range employees {
		if emp.Status != model.EmployeeStatusActive {
			continue
		}

		// Vereinfachte Logik: Reviews sind überfällig wenn > 13 Monate seit Einstellung
		monthsSinceHire := int(time.Since(emp.HireDate).Hours() / (24 * 30.44))
		if monthsSinceHire > 13 {
			overdue++
		}
	}

	return overdue
}

// calculateTurnoverRate berechnet die Fluktuationsrate
func (h *HRService) calculateTurnoverRate(employees []*model.Employee) float64 {
	if len(employees) == 0 {
		return 0
	}

	// Vereinfachte Berechnung basierend auf inaktiven Mitarbeitern
	inactive := 0
	for _, emp := range employees {
		if emp.Status == model.EmployeeStatusInactive {
			inactive++
		}
	}

	return float64(inactive) / float64(len(employees)) * 100
}

// calculateAverageEmployment berechnet die durchschnittliche Betriebszugehörigkeit
func (h *HRService) calculateAverageEmployment(activeEmployees []*model.Employee) float64 {
	if len(activeEmployees) == 0 {
		return 0
	}

	totalTenure := 0.0
	for _, emp := range activeEmployees {
		totalTenure += h.calculateTenure(emp.HireDate)
	}

	return totalTenure / float64(len(activeEmployees))
}

// GetDepartmentLabelsAndData konvertiert die Department-Map zu Arrays für Charts
func (h *HRService) GetDepartmentLabelsAndData(deptCounts map[string]int) ([]string, []int) {
	var labels []string
	var data []int

	// Sortiere Abteilungen nach Anzahl (absteigend)
	type deptData struct {
		name  string
		count int
	}

	var departments []deptData
	for name, count := range deptCounts {
		departments = append(departments, deptData{name, count})
	}

	sort.Slice(departments, func(i, j int) bool {
		return departments[i].count > departments[j].count
	})

	for _, dept := range departments {
		labels = append(labels, dept.name)
		data = append(data, dept.count)
	}

	return labels, data
}

// GetStatusLabelsAndData konvertiert die Status-Map zu Arrays für Charts
func (h *HRService) GetStatusLabelsAndData(statusCounts map[string]int) ([]string, []int) {
	var labels []string
	var data []int

	// Definierte Reihenfolge für Status
	statusOrder := []string{"Aktiv", "Im Urlaub", "Krank", "Inaktiv"}

	for _, status := range statusOrder {
		if count, exists := statusCounts[status]; exists && count > 0 {
			labels = append(labels, status)
			data = append(data, count)
		}
	}

	return labels, data
}

// GetAgeLabelsAndData konvertiert die Alters-Map zu Arrays für Charts
func (h *HRService) GetAgeLabelsAndData(ageCounts map[string]int) ([]string, []int) {
	var labels []string
	var data []int

	// Definierte Reihenfolge für Altersgruppen
	ageOrder := []string{"< 25", "25-34", "35-44", "45-54", "55-64", "≥ 65"}

	for _, ageGroup := range ageOrder {
		if count, exists := ageCounts[ageGroup]; exists && count > 0 {
			labels = append(labels, ageGroup)
			data = append(data, count)
		}
	}

	return labels, data
}

// GetTenureLabelsAndData konvertiert die Betriebszugehörigkeits-Map zu Arrays für Charts
func (h *HRService) GetTenureLabelsAndData(tenureCounts map[string]int) ([]string, []int) {
	var labels []string
	var data []int

	// Definierte Reihenfolge für Betriebszugehörigkeit
	tenureOrder := []string{"< 1 Jahr", "1-3 Jahre", "3-5 Jahre", "5-10 Jahre", "≥ 10 Jahre"}

	for _, tenureGroup := range tenureOrder {
		if count, exists := tenureCounts[tenureGroup]; exists && count > 0 {
			labels = append(labels, tenureGroup)
			data = append(data, count)
		}
	}

	return labels, data
}
