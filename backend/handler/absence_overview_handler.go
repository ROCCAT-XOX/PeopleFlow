package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
	"net/http"
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

// GetAbsenceOverview zeigt die Abwesenheitsübersicht an
func (h *AbsenceOverviewHandler) GetAbsenceOverview(c *gin.Context) {
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

	// Für normale User und HR: Eigenen Mitarbeiter finden
	var currentEmployee *model.Employee
	if userRole == string(model.RoleUser) || userRole == string(model.RoleHR) {
		for _, emp := range employees {
			if emp.Email == userModel.Email {
				currentEmployee = emp
				break
			}
		}
	}

	// Abwesenheiten sammeln und nach Status gruppieren
	var allAbsences []AbsenceWithEmployee
	var pendingAbsences []AbsenceWithEmployee
	var upcomingAbsences []AbsenceWithEmployee

	currentDate := time.Now()
	currentYear := currentDate.Year()

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			absenceWithEmp := AbsenceWithEmployee{
				Absence:       absence,
				EmployeeID:    emp.ID.Hex(),
				EmployeeName:  emp.FirstName + " " + emp.LastName,
				Department:    string(emp.Department),
				EmployeeEmail: emp.Email,
			}

			allAbsences = append(allAbsences, absenceWithEmp)

			if absence.Status == "requested" {
				pendingAbsences = append(pendingAbsences, absenceWithEmp)
			}

			if absence.Status == "approved" && absence.StartDate.After(currentDate) {
				upcomingAbsences = append(upcomingAbsences, absenceWithEmp)
			}
		}
	}

	// Statistiken berechnen
	var totalVacationDays, totalSickDays float64
	employeesOnVacation := 0
	employeesOnSick := 0

	for _, emp := range employees {
		for _, absence := range emp.Absences {
			if absence.Status == "approved" && absence.StartDate.Year() == currentYear {
				if absence.Type == "vacation" {
					totalVacationDays += absence.Days
				} else if absence.Type == "sick" {
					totalSickDays += absence.Days
				}

				// Prüfen ob aktuell abwesend
				if absence.StartDate.Before(currentDate) && absence.EndDate.After(currentDate) {
					if absence.Type == "vacation" {
						employeesOnVacation++
					} else if absence.Type == "sick" {
						employeesOnSick++
					}
				}
			}
		}
	}

	// Persönliche Statistiken für User/HR
	var personalVacationDays, personalSickDays, personalRemainingVacation float64
	if currentEmployee != nil {
		for _, absence := range currentEmployee.Absences {
			if absence.Status == "approved" && absence.StartDate.Year() == currentYear {
				if absence.Type == "vacation" {
					personalVacationDays += absence.Days
				} else if absence.Type == "sick" {
					personalSickDays += absence.Days
				}
			}
		}
		personalRemainingVacation = float64(currentEmployee.VacationDays) - personalVacationDays
	}

	c.HTML(http.StatusOK, "absence_overview.html", gin.H{
		"title":                     "Abwesenheiten",
		"active":                    "absences",
		"user":                      userModel.FirstName + " " + userModel.LastName,
		"email":                     userModel.Email,
		"year":                      currentYear,
		"userRole":                  userRole,
		"currentEmployee":           currentEmployee,
		"employees":                 employees,
		"allAbsences":               allAbsences,
		"pendingAbsences":           pendingAbsences,
		"upcomingAbsences":          upcomingAbsences,
		"pendingCount":              len(pendingAbsences),
		"totalVacationDays":         totalVacationDays,
		"totalSickDays":             totalSickDays,
		"employeesOnVacation":       employeesOnVacation,
		"employeesOnSick":           employeesOnSick,
		"personalVacationDays":      personalVacationDays,
		"personalSickDays":          personalSickDays,
		"personalRemainingVacation": personalRemainingVacation,
	})
}

// AddAbsenceRequest fügt eine neue Abwesenheitsanfrage hinzu
func (h *AbsenceOverviewHandler) AddAbsenceRequest(c *gin.Context) {
	employeeID := c.PostForm("employeeId")
	absenceType := c.PostForm("type")
	startDateStr := c.PostForm("startDate")
	endDateStr := c.PostForm("endDate")
	reason := c.PostForm("reason")
	notes := c.PostForm("notes")

	// Datumsvalidierung
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Startdatum"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiges Enddatum"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Enddatum muss nach dem Startdatum liegen"})
		return
	}

	// Tage berechnen
	days := endDate.Sub(startDate).Hours()/24 + 1

	// Aktuellen Benutzer abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	// Mitarbeiter bestimmen
	var targetEmployeeID string
	if userRole == string(model.RoleUser) || userRole == string(model.RoleHR) {
		// Eigene Abwesenheit
		employee, err := h.employeeRepo.FindByEmail(userModel.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mitarbeiterdaten nicht gefunden"})
			return
		}
		targetEmployeeID = employee.ID.Hex()
	} else {
		// Manager/Admin kann für andere beantragen
		if employeeID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mitarbeiter muss ausgewählt werden"})
			return
		}
		targetEmployeeID = employeeID
	}

	// Mitarbeiter laden
	employee, err := h.employeeRepo.FindByID(targetEmployeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Bei Urlaub: Verfügbare Tage prüfen
	if absenceType == "vacation" {
		currentYear := time.Now().Year()
		usedDays := 0.0
		for _, absence := range employee.Absences {
			if absence.Status == "approved" && absence.Type == "vacation" && absence.StartDate.Year() == currentYear {
				usedDays += absence.Days
			}
		}

		remainingDays := float64(employee.VacationDays) - usedDays
		if days > remainingDays {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Nicht genügend Urlaubstage. Verfügbar: %.0f, Angefordert: %.0f", remainingDays, days),
			})
			return
		}
	}

	// Neue Abwesenheit erstellen
	absence := model.Absence{
		ID:        primitive.NewObjectID(),
		Type:      absenceType,
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
		Status:    "requested",
		Reason:    reason,
		Notes:     notes,
	}

	// Zur Mitarbeiter-Abwesenheitsliste hinzufügen
	employee.Absences = append(employee.Absences, absence)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der Abwesenheit"})
		return
	}

	// Aktivität loggen
	_, _ = h.activityRepo.LogActivity(
		model.ActivityTypeVacationRequested,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		employee.ID,
		"employee",
		employee.FirstName+" "+employee.LastName,
		fmt.Sprintf("%s-Antrag eingereicht: %.0f Tage", getAbsenceTypeDisplay(absenceType), days),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Abwesenheitsantrag wurde erfolgreich eingereicht",
	})
}

// ApproveAbsenceRequest genehmigt oder lehnt eine Abwesenheit ab
func (h *AbsenceOverviewHandler) ApproveAbsenceRequest(c *gin.Context) {
	employeeID := c.Param("employeeId")
	absenceID := c.Param("absenceId")
	action := c.PostForm("action")

	// Aktuellen Benutzer abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Mitarbeiter laden
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Abwesenheit finden und aktualisieren
	absenceObjID, _ := primitive.ObjectIDFromHex(absenceID)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Abwesenheit nicht gefunden"})
		return
	}

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren"})
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
