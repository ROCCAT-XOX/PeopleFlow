package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AbsenceOverviewHandler verwaltet die Abwesenheitsübersicht
type AbsenceOverviewHandler struct {
	employeeRepo *repository.EmployeeRepository
	userRepo     *repository.UserRepository
	activityRepo *repository.ActivityRepository
}

// NewAbsenceOverviewHandler erstellt einen neuen AbsenceOverviewHandler
func NewAbsenceOverviewHandler() *AbsenceOverviewHandler {
	return &AbsenceOverviewHandler{
		employeeRepo: repository.NewEmployeeRepository(),
		userRepo:     repository.NewUserRepository(),
		activityRepo: repository.NewActivityRepository(),
	}
}

func (h *AbsenceOverviewHandler) GetAbsenceOverview(c *gin.Context) {
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Repositories
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter abrufen
	employees, _, err := employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		employees = []*model.Employee{}
	}

	// Verschiedene Abwesenheitslisten erstellen
	var pendingRequests []gin.H
	var upcomingAbsences []gin.H
	var allAbsences []gin.H

	pendingCount := 0
	approvedCount := 0
	rejectedCount := 0
	upcomingCount := 0

	now := time.Now()

	// Durch alle Mitarbeiter iterieren und Abwesenheiten sammeln
	for _, emp := range employees {
		for _, absence := range emp.Absences {
			absenceData := gin.H{
				"ID":           absence.ID.Hex(),
				"EmployeeID":   emp.ID.Hex(),
				"EmployeeName": emp.FirstName + " " + emp.LastName,
				"Department":   emp.Department,
				"Type":         absence.Type,
				"StartDate":    absence.StartDate,
				"EndDate":      absence.EndDate,
				"Days":         absence.Days,
				"Status":       absence.Status,
				"Reason":       absence.Reason,
				"ApproverName": absence.ApproverName,
				"CreatedAt":    absence.StartDate, // Falls CreatedAt nicht verfügbar
			}

			// In alle Abwesenheiten aufnehmen
			allAbsences = append(allAbsences, absenceData)

			// Nach Status kategorisieren
			switch absence.Status {
			case "requested":
				pendingRequests = append(pendingRequests, absenceData)
				pendingCount++
			case "approved":
				approvedCount++
				// Prüfen ob es eine zukünftige Abwesenheit ist
				if absence.StartDate.After(now) {
					upcomingAbsences = append(upcomingAbsences, absenceData)
					upcomingCount++
				}
			case "rejected":
				rejectedCount++
			}
		}
	}

	// Nach Startdatum sortieren
	sort.Slice(pendingRequests, func(i, j int) bool {
		return pendingRequests[i]["StartDate"].(time.Time).Before(pendingRequests[j]["StartDate"].(time.Time))
	})

	sort.Slice(upcomingAbsences, func(i, j int) bool {
		return upcomingAbsences[i]["StartDate"].(time.Time).Before(upcomingAbsences[j]["StartDate"].(time.Time))
	})

	// Alle Abwesenheiten nach Startdatum sortieren (neueste zuerst)
	sort.Slice(allAbsences, func(i, j int) bool {
		return allAbsences[i]["StartDate"].(time.Time).After(allAbsences[j]["StartDate"].(time.Time))
	})

	c.HTML(http.StatusOK, "absence_overview.html", gin.H{
		"title":            "Abwesenheitsanträge",
		"active":           "absence-overview",
		"user":             userModel.FirstName + " " + userModel.LastName,
		"email":            userModel.Email,
		"userRole":         userRole,
		"year":             time.Now().Year(),
		"currentDate":      time.Now().Format("Monday, 02. January 2006"),
		"employees":        employees,
		"pendingRequests":  pendingRequests,
		"upcomingAbsences": upcomingAbsences,
		"allAbsences":      allAbsences,
		"pendingCount":     pendingCount,
		"approvedCount":    approvedCount,
		"rejectedCount":    rejectedCount,
		"upcomingCount":    upcomingCount,
	})
}

// AddAbsenceRequest fügt einen neuen Abwesenheitsantrag hinzu
func (h *AbsenceOverviewHandler) AddAbsenceRequest(c *gin.Context) {
	// Benutzer und Rolle abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Prüfen ob der Benutzer berechtigt ist
	if userRole != string(model.RoleAdmin) &&
		userRole != string(model.RoleManager) &&
		userRole != string(model.RoleHR) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Keine Berechtigung, Abwesenheitsanträge zu stellen",
		})
		return
	}

	// Formulardaten abrufen
	employeeID := c.PostForm("employeeId")
	absenceType := c.PostForm("type")
	startDateStr := c.PostForm("startDate")
	endDateStr := c.PostForm("endDate")
	reason := c.PostForm("reason")
	notes := c.PostForm("notes")

	// Mitarbeiter-Repository
	employeeRepo := repository.NewEmployeeRepository()

	// Mitarbeiter abrufen
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Mitarbeiter nicht gefunden",
		})
		return
	}

	// Datumsfelder parsen
	var startDate, endDate time.Time

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Ungültiges Startdatum",
			})
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Ungültiges Enddatum",
			})
			return
		}
	}

	// Validierung
	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Das Startdatum muss vor dem Enddatum liegen",
		})
		return
	}

	// Tage berechnen (inklusive Wochenenden erstmal)
	days := int(endDate.Sub(startDate).Hours()/24) + 1

	// Neue Abwesenheit erstellen
	absence := model.Absence{
		ID:        primitive.NewObjectID(),
		Type:      absenceType,
		StartDate: startDate,
		EndDate:   endDate,
		Days:      float64(days),
		Status:    "requested",
		Reason:    reason,
		Notes:     notes,
	}

	// Abwesenheit zum Mitarbeiter hinzufügen
	employee.Absences = append(employee.Absences, absence)

	// Mitarbeiter aktualisieren
	err = employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Speichern des Abwesenheitsantrags",
		})
		return
	}

	// Aktivität loggen
	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeVacationRequested,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		employee.ID,
		"employee",
		employee.FirstName+" "+employee.LastName,
		fmt.Sprintf("Abwesenheitsantrag gestellt: %s vom %s bis %s",
			absenceType,
			startDate.Format("02.01.2006"),
			endDate.Format("02.01.2006")),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Abwesenheitsantrag wurde erfolgreich gestellt",
	})
}

// ApproveAbsenceRequest genehmigt oder lehnt eine Abwesenheit ab
func (h *AbsenceOverviewHandler) ApproveAbsenceRequest(c *gin.Context) {
	employeeID := c.Param("employeeId")
	absenceID := c.Param("absenceId")
	action := c.PostForm("action")

	// Benutzer und Rolle abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Nur Admin und Manager dürfen genehmigen/ablehnen
	if userRole != string(model.RoleAdmin) && userRole != string(model.RoleManager) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Sie haben keine Berechtigung, Abwesenheitsanträge zu bearbeiten",
		})
		return
	}

	// Mitarbeiter aus der Datenbank laden
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Mitarbeiter nicht gefunden",
		})
		return
	}

	// Abwesenheit finden und aktualisieren
	absenceObjID, err := primitive.ObjectIDFromHex(absenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültige Abwesenheits-ID",
		})
		return
	}

	found := false
	for i, absence := range employee.Absences {
		if absence.ID == absenceObjID {
			if action == "approve" {
				employee.Absences[i].Status = "approved"
			} else {
				employee.Absences[i].Status = "rejected"
			}
			employee.Absences[i].ApprovedBy = userModel.ID
			employee.Absences[i].ApproverName = userModel.FirstName + " " + userModel.LastName
			found = true

			// Aktivität loggen
			activityType := model.ActivityTypeVacationApproved
			if action == "reject" {
				activityType = model.ActivityTypeVacationRejected
			}

			_, _ = h.activityRepo.LogActivity(
				activityType,
				userModel.ID,
				userModel.FirstName+" "+userModel.LastName,
				employee.ID,
				"employee",
				employee.FirstName+" "+employee.LastName,
				fmt.Sprintf("%s-Antrag %s",
					getAbsenceTypeDisplay(absence.Type),
					getActionDisplay(action)),
			)
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Abwesenheit nicht gefunden",
		})
		return
	}

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Aktualisieren",
		})
		return
	}

	message := "Abwesenheit wurde genehmigt"
	if action == "reject" {
		message = "Abwesenheit wurde abgelehnt"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// Hilfsfunktionen
func getAbsenceTypeDisplay(absenceType string) string {
	switch absenceType {
	case "vacation":
		return "Urlaub"
	case "sick":
		return "Krankheit"
	case "special":
		return "Sonderurlaub"
	default:
		return absenceType
	}
}

func getActionDisplay(action string) string {
	if action == "approve" {
		return "genehmigt"
	}
	return "abgelehnt"
}

// AbsenceWithEmployee erweitert Absence um Mitarbeiterinformationen
type AbsenceWithEmployee struct {
	model.Absence
	EmployeeID    string
	EmployeeName  string
	Department    string
	EmployeeEmail string
}
