package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"time"
)

// OvertimeHandler verwaltet alle Überstunden-bezogenen Anfragen
type OvertimeHandler struct {
	employeeRepo       *repository.EmployeeRepository
	timeAccountService *service.TimeAccountService
}

// NewOvertimeHandler erstellt einen neuen OvertimeHandler
func NewOvertimeHandler() *OvertimeHandler {
	return &OvertimeHandler{
		employeeRepo:       repository.NewEmployeeRepository(),
		timeAccountService: service.NewTimeAccountService(),
	}
}

// OvertimeEmployeeSummary repräsentiert die Überstunden-Daten für einen Mitarbeiter
type OvertimeEmployeeSummary struct {
	EmployeeID      string    `json:"employeeId"`
	EmployeeName    string    `json:"employeeName"`
	Department      string    `json:"department"`
	HasProfileImage bool      `json:"hasProfileImage"`
	WeeklyTarget    float64   `json:"weeklyTarget"`
	TotalHours      float64   `json:"totalHours"`
	OvertimeBalance float64   `json:"overtimeBalance"`
	OvertimeStatus  string    `json:"overtimeStatus"`
	LastCalculated  time.Time `json:"lastCalculated"`
	WorkTimeModel   string    `json:"workTimeModel"`
}

// GetOvertimeView zeigt die Überstunden-Übersicht an
func (h *OvertimeHandler) GetOvertimeView(c *gin.Context) {
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

	// Überstunden-Zusammenfassung für alle Mitarbeiter erstellen
	var overtimeEmployees []OvertimeEmployeeSummary
	var totalOvertimeBalance float64
	var positiveCount, negativeCount, neutralCount int

	for _, emp := range employees {
		// Nur Mitarbeiter mit Zeiteinträgen berücksichtigen
		if len(emp.TimeEntries) == 0 {
			continue
		}

		// Gesamtstunden berechnen
		var totalHours float64
		for _, entry := range emp.TimeEntries {
			totalHours += entry.Duration
		}

		// Überstunden-Status bestimmen
		status := emp.GetOvertimeStatus()
		switch status {
		case "positive":
			positiveCount++
		case "negative":
			negativeCount++
		default:
			neutralCount++
		}

		totalOvertimeBalance += emp.OvertimeBalance

		// Mitarbeiter-Zusammenfassung erstellen
		overtimeSummary := OvertimeEmployeeSummary{
			EmployeeID:      emp.ID.Hex(),
			EmployeeName:    emp.FirstName + " " + emp.LastName,
			Department:      string(emp.Department),
			HasProfileImage: len(emp.ProfileImageData.Data) > 0,
			WeeklyTarget:    emp.GetWeeklyTargetHours(),
			TotalHours:      totalHours,
			OvertimeBalance: emp.OvertimeBalance,
			OvertimeStatus:  status,
			LastCalculated:  emp.LastTimeCalculated,
			WorkTimeModel:   emp.WorkTimeModel.GetDisplayName(),
		}

		overtimeEmployees = append(overtimeEmployees, overtimeSummary)
	}

	// Nach Überstunden-Saldo sortieren (höchste zuerst)
	sort.Slice(overtimeEmployees, func(i, j int) bool {
		return overtimeEmployees[i].OvertimeBalance > overtimeEmployees[j].OvertimeBalance
	})

	// Durchschnittliche Wochenstunden berechnen
	var averageWeeklyHours float64
	if len(overtimeEmployees) > 0 {
		var totalTargetHours float64
		for _, emp := range overtimeEmployees {
			totalTargetHours += emp.WeeklyTarget
		}
		averageWeeklyHours = totalTargetHours / float64(len(overtimeEmployees))
	}

	// Abteilungen für Filter sammeln
	departmentMap := make(map[string]bool)
	for _, emp := range overtimeEmployees {
		if emp.Department != "" {
			departmentMap[emp.Department] = true
		}
	}

	var departments []string
	for dept := range departmentMap {
		departments = append(departments, dept)
	}
	sort.Strings(departments)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "overtime.html", gin.H{
		"title":                       "Überstunden",
		"active":                      "overtime",
		"user":                        userModel.FirstName + " " + userModel.LastName,
		"email":                       userModel.Email,
		"year":                        time.Now().Year(),
		"userRole":                    userRole,
		"employeeSummaryWithOvertime": overtimeEmployees,
		"totalEmployees":              len(overtimeEmployees),
		"totalOvertimeBalance":        totalOvertimeBalance,
		"positiveCount":               positiveCount,
		"negativeCount":               negativeCount,
		"neutralCount":                neutralCount,
		"averageWeeklyHours":          averageWeeklyHours,
		"departments":                 departments,
	})
}

// RecalculateAllOvertime berechnet Überstunden für alle Mitarbeiter neu
func (h *OvertimeHandler) RecalculateAllOvertime(c *gin.Context) {
	err := h.timeAccountService.RecalculateAllEmployeeOvertimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fehler bei der Überstunden-Berechnung: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Überstunden für alle Mitarbeiter wurden erfolgreich neu berechnet",
	})
}

// ExportOvertimeData exportiert Überstunden-Daten als CSV
func (h *OvertimeHandler) ExportOvertimeData(c *gin.Context) {
	// Filter-Parameter abrufen
	balanceFilter := c.Query("balanceFilter")
	departmentFilter := c.Query("departmentFilter")

	// Alle Mitarbeiter abrufen
	employees, err := h.employeeRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fehler beim Abrufen der Mitarbeiter",
		})
		return
	}

	// CSV-Header
	csvContent := "Mitarbeiter,Abteilung,Wochenstunden (Soll),Erfasste Stunden,Überstunden-Saldo,Status,Letzte Berechnung\n"

	// Daten filtern und in CSV konvertieren
	for _, emp := range employees {
		if len(emp.TimeEntries) == 0 {
			continue
		}

		// Filter anwenden
		status := emp.GetOvertimeStatus()
		if balanceFilter != "" && balanceFilter != "all" && balanceFilter != status {
			continue
		}

		if departmentFilter != "" && string(emp.Department) != departmentFilter {
			continue
		}

		// Gesamtstunden berechnen
		var totalHours float64
		for _, entry := range emp.TimeEntries {
			totalHours += entry.Duration
		}

		// Status in deutscher Sprache
		var statusGerman string
		switch status {
		case "positive":
			statusGerman = "Überstunden"
		case "negative":
			statusGerman = "Minusstunden"
		default:
			statusGerman = "Ausgeglichen"
		}

		// Letzte Berechnung formatieren
		lastCalculated := "Noch nicht berechnet"
		if !emp.LastTimeCalculated.IsZero() {
			lastCalculated = emp.LastTimeCalculated.Format("02.01.2006 15:04")
		}

		// CSV-Zeile hinzufügen
		csvContent += emp.FirstName + " " + emp.LastName + ","
		csvContent += string(emp.Department) + ","
		csvContent += fmt.Sprintf("%.1f", emp.GetWeeklyTargetHours()) + ","
		csvContent += fmt.Sprintf("%.1f", totalHours) + ","
		csvContent += fmt.Sprintf("%.2f", emp.OvertimeBalance) + ","
		csvContent += statusGerman + ","
		csvContent += lastCalculated + "\n"
	}

	// CSV-Datei senden
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=ueberstunden.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, csvContent)
}

// GetEmployeeOvertimeDetails liefert detaillierte Überstunden-Informationen für einen Mitarbeiter
func (h *OvertimeHandler) GetEmployeeOvertimeDetails(c *gin.Context) {
	employeeID := c.Param("id")

	overtimeSummary, err := h.timeAccountService.GetEmployeeOvertimeSummary(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Mitarbeiter nicht gefunden oder Fehler bei der Berechnung",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    overtimeSummary,
	})
}
