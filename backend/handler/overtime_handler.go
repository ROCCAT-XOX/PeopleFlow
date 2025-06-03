package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sort"
	"strconv"
	"time"
)

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
	var totalFinalBalance float64 // Neue Variable für finales Saldo
	var positiveCount, negativeCount, neutralCount int

	for _, emp := range employees {
		// Nur Mitarbeiter mit Zeiteinträgen berücksichtigen
		if len(emp.TimeEntries) == 0 {
			continue
		}

		// Anpassungen für diesen Mitarbeiter laden
		adjustments, err := h.overtimeAdjustmentRepo.FindByEmployeeID(emp.ID.Hex())
		if err == nil {
			emp.OvertimeAdjustments = make([]model.OvertimeAdjustment, len(adjustments))
			for i, adj := range adjustments {
				emp.OvertimeAdjustments[i] = *adj
			}
		}

		// Gesamtstunden berechnen
		var totalHours float64
		for _, entry := range emp.TimeEntries {
			totalHours += entry.Duration
		}

		// Basis-Überstunden-Saldo
		baseBalance := emp.OvertimeBalance
		// Anpassungen-Saldo
		adjustmentsTotal := emp.GetTotalAdjustments()
		// Finales Saldo
		finalBalance := baseBalance + adjustmentsTotal

		// Überstunden-Status basierend auf finalem Saldo bestimmen
		var status string
		if finalBalance > 0 {
			status = "positive"
			positiveCount++
		} else if finalBalance < 0 {
			status = "negative"
			negativeCount++
		} else {
			status = "neutral"
			neutralCount++
		}

		totalOvertimeBalance += baseBalance
		totalFinalBalance += finalBalance

		// Mitarbeiter-Zusammenfassung erstellen
		overtimeSummary := OvertimeEmployeeSummary{
			EmployeeID:      emp.ID.Hex(),
			EmployeeName:    emp.FirstName + " " + emp.LastName,
			Department:      string(emp.Department),
			HasProfileImage: len(emp.ProfileImageData.Data) > 0,
			WeeklyTarget:    emp.GetWeeklyTargetHours(),
			TotalHours:      totalHours,
			OvertimeBalance: finalBalance, // Verwende finales Saldo statt Basis-Saldo
			OvertimeStatus:  status,
			LastCalculated:  emp.LastTimeCalculated,
			WorkTimeModel:   emp.WorkTimeModel.GetDisplayName(),
		}

		overtimeEmployees = append(overtimeEmployees, overtimeSummary)
	}

	// Nach finalem Überstunden-Saldo sortieren (höchste zuerst)
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

	// Ausstehende Anpassungen für Admin/Manager laden
	var pendingAdjustments []*model.OvertimeAdjustment
	var pendingCount int
	if userRole == string(model.RoleAdmin) || userRole == string(model.RoleManager) {
		pendingAdjustments, err = h.overtimeAdjustmentRepo.FindPending()
		if err != nil {
			fmt.Printf("Error loading pending adjustments: %v\n", err)
			pendingAdjustments = []*model.OvertimeAdjustment{}
		}
		pendingCount = len(pendingAdjustments)
	}

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
		"totalOvertimeBalance":        totalFinalBalance,    // Verwende finales Saldo
		"totalBaseBalance":            totalOvertimeBalance, // Zusätzlich Basis-Saldo für Vergleich
		"positiveCount":               positiveCount,
		"negativeCount":               negativeCount,
		"neutralCount":                neutralCount,
		"averageWeeklyHours":          averageWeeklyHours,
		"departments":                 departments,
		"pendingAdjustments":          pendingAdjustments,
		"pendingCount":                pendingCount,
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

// =========================== OVERTIME ADJUSTMENT =========================================

// Ergänzungen für backend/handler/overtime_handler.go

// Neue Struktur für Overtime Handler hinzufügen:
type OvertimeHandler struct {
	employeeRepo           *repository.EmployeeRepository
	timeAccountService     *service.TimeAccountService
	overtimeAdjustmentRepo *repository.OvertimeAdjustmentRepository
}

// NewOvertimeHandler Konstruktor erweitern:
func NewOvertimeHandler() *OvertimeHandler {
	return &OvertimeHandler{
		employeeRepo:           repository.NewEmployeeRepository(),
		timeAccountService:     service.NewTimeAccountService(),
		overtimeAdjustmentRepo: repository.NewOvertimeAdjustmentRepository(),
	}
}

// Neue Handler-Methoden hinzufügen:

// AddOvertimeAdjustment fügt eine manuelle Überstunden-Anpassung hinzu
func (h *OvertimeHandler) AddOvertimeAdjustment(c *gin.Context) {
	employeeID := c.Param("id")

	// Formulardaten abrufen
	adjustmentType := c.PostForm("type")
	hoursStr := c.PostForm("hours")
	reason := c.PostForm("reason")
	description := c.PostForm("description")

	// Validierung
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Begründung ist erforderlich"})
		return
	}

	hours, err := strconv.ParseFloat(hoursStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Stundenangabe"})
		return
	}

	// Mitarbeiter prüfen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Aktuellen Benutzer abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	empObjID, _ := primitive.ObjectIDFromHex(employeeID)

	// Anpassung erstellen
	adjustment := &model.OvertimeAdjustment{
		EmployeeID:   empObjID,
		Type:         model.OvertimeAdjustmentType(adjustmentType),
		Hours:        hours,
		Reason:       reason,
		Description:  description,
		AdjustedBy:   userModel.ID,
		AdjusterName: userModel.FirstName + " " + userModel.LastName,
		Status:       "pending",
	}

	// In Datenbank speichern
	err = h.overtimeAdjustmentRepo.Create(adjustment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der Anpassung"})
		return
	}

	// Aktivität loggen
	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeEmployeeUpdated,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		employee.ID,
		"employee",
		employee.FirstName+" "+employee.LastName,
		fmt.Sprintf("Manuelle Überstunden-Anpassung hinzugefügt: %s", adjustment.FormatHours()),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Überstunden-Anpassung wurde eingereicht und wartet auf Genehmigung",
		"data":    adjustment,
	})
}

// GetEmployeeAdjustments liefert alle Anpassungen für einen Mitarbeiter
func (h *OvertimeHandler) GetEmployeeAdjustments(c *gin.Context) {
	employeeID := c.Param("id")

	adjustments, err := h.overtimeAdjustmentRepo.FindByEmployeeID(employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der Anpassungen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    adjustments,
	})
}

// ApproveAdjustment genehmigt eine Überstunden-Anpassung
func (h *OvertimeHandler) ApproveAdjustment(c *gin.Context) {
	adjustmentID := c.Param("adjustmentId")
	action := c.PostForm("action") // "approve" oder "reject"

	// Aktuellen Benutzer abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Status bestimmen
	status := "approved"
	if action == "reject" {
		status = "rejected"
	}

	// Status aktualisieren
	err := h.overtimeAdjustmentRepo.UpdateStatus(adjustmentID, status, userModel.ID, userModel.FirstName+" "+userModel.LastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Status"})
		return
	}

	// Anpassung abrufen für Logging
	adjustment, err := h.overtimeAdjustmentRepo.FindByID(adjustmentID)
	if err == nil {
		// Mitarbeiter abrufen
		employee, err := h.employeeRepo.FindByID(adjustment.EmployeeID.Hex())
		if err == nil {
			// Aktivität loggen
			activityRepo := repository.NewActivityRepository()
			actionText := "genehmigt"
			if status == "rejected" {
				actionText = "abgelehnt"
			}
			_, _ = activityRepo.LogActivity(
				model.ActivityTypeEmployeeUpdated,
				userModel.ID,
				userModel.FirstName+" "+userModel.LastName,
				employee.ID,
				"employee",
				employee.FirstName+" "+employee.LastName,
				fmt.Sprintf("Überstunden-Anpassung %s: %s", actionText, adjustment.FormatHours()),
			)
		}
	}

	message := "Anpassung wurde genehmigt"
	if status == "rejected" {
		message = "Anpassung wurde abgelehnt"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// GetPendingAdjustments liefert alle ausstehenden Anpassungen
func (h *OvertimeHandler) GetPendingAdjustments(c *gin.Context) {
	adjustments, err := h.overtimeAdjustmentRepo.FindPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Abrufen der ausstehenden Anpassungen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    adjustments,
	})
}
