package handler

import (
	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmployeeHandler verwaltet alle Anfragen zu Mitarbeitern
type EmployeeHandler struct {
	employeeRepo *repository.EmployeeRepository
	userRepo     *repository.UserRepository
}

// NewEmployeeHandler erstellt einen neuen EmployeeHandler
func NewEmployeeHandler() *EmployeeHandler {
	return &EmployeeHandler{
		employeeRepo: repository.NewEmployeeRepository(),
		userRepo:     repository.NewUserRepository(),
	}
}

// ListEmployees zeigt die Liste aller Mitarbeiter an
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

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

	// Liste der Manager (für Dropdown-Menüs) abrufen
	managers, err := h.employeeRepo.FindManagers()
	if err != nil {
		managers = []*model.Employee{} // Leere Liste, falls ein Fehler auftritt
	}

	// Wir erstellen hier EmployeeViewModel-Strukturen, die für die Anzeige optimiert sind
	var employeeViewModels []gin.H
	for _, emp := range employees {
		// Formatiertes Einstellungsdatum
		hireDateFormatted := emp.HireDate.Format("02.01.2006")

		// Status menschenlesbar machen
		status := "Aktiv"
		switch emp.Status {
		case model.EmployeeStatusInactive:
			status = "Inaktiv"
		case model.EmployeeStatusOnLeave:
			status = "Im Urlaub"
		case model.EmployeeStatusRemote:
			status = "Remote"
		}

		// Standard-Profilbild, falls keines definiert ist
		profileImage := emp.ProfileImage
		if profileImage == "" {
			profileImage = "/static/img/default-avatar.png"
		}

		// ViewModel erstellen
		employeeViewModels = append(employeeViewModels, gin.H{
			"ID":                emp.ID.Hex(),
			"FirstName":         emp.FirstName,
			"LastName":          emp.LastName,
			"Email":             emp.Email,
			"Position":          emp.Position,
			"Department":        emp.Department,
			"HireDateFormatted": hireDateFormatted,
			"Status":            status,
			"ProfileImage":      profileImage,
		})
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employees.html", gin.H{
		"title":          "Mitarbeiter",
		"active":         "employees",
		"user":           userModel.FirstName + " " + userModel.LastName,
		"email":          userModel.Email,
		"year":           time.Now().Year(),
		"employees":      employeeViewModels,
		"totalEmployees": len(employees),
		"managers":       managers,
	})
}

// AddEmployee fügt einen neuen Mitarbeiter hinzu
func (h *EmployeeHandler) AddEmployee(c *gin.Context) {
	// Formulardaten abrufen
	firstName := c.PostForm("firstName")
	lastName := c.PostForm("lastName")
	email := c.PostForm("email")
	position := c.PostForm("position")
	department := c.PostForm("department")

	// Weitere Felder aus dem Formular extrahieren
	// (gekürzt für Übersichtlichkeit)

	// Datumsfelder parsen
	var hireDate time.Time
	hireDateStr := c.PostForm("hireDate")
	if hireDateStr != "" {
		var err error
		hireDate, err = time.Parse("2006-01-02", hireDateStr)
		if err != nil {
			hireDate = time.Now() // Fallback auf aktuelles Datum
		}
	} else {
		hireDate = time.Now()
	}

	var birthDate time.Time
	birthDateStr := c.PostForm("birthDate")
	if birthDateStr != "" {
		birthDate, _ = time.Parse("2006-01-02", birthDateStr)
	}

	// Manager-ID parsen, falls vorhanden
	var managerID primitive.ObjectID
	managerIDStr := c.PostForm("managerId")
	if managerIDStr != "" {
		var err error
		managerID, err = primitive.ObjectIDFromHex(managerIDStr)
		if err != nil {
			// Ignorieren, wenn die ID ungültig ist
			managerID = primitive.NilObjectID
		}
	}

	var salary float64
	salaryStr := c.PostForm("salary")
	if salaryStr != "" {
		// Konvertieren und Fehler ignorieren
		salary, _ = strconv.ParseFloat(salaryStr, 64)
	}

	// Neues Employee-Objekt erstellen
	employee := &model.Employee{
		FirstName:      firstName,
		LastName:       lastName,
		Email:          email,
		Phone:          c.PostForm("phone"),
		Address:        c.PostForm("address"),
		DateOfBirth:    birthDate,
		HireDate:       hireDate,
		Position:       position,
		Department:     model.Department(department),
		ManagerID:      managerID,
		Status:         model.EmployeeStatusActive, // Standardmäßig aktiv
		Salary:         salary,
		BankAccount:    c.PostForm("iban"),
		TaxID:          c.PostForm("taxClass"),
		SocialSecID:    c.PostForm("socialSecId"),
		EmergencyName:  c.PostForm("emergencyName"),
		EmergencyPhone: c.PostForm("emergencyPhone"),
		Notes:          c.PostForm("notes"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Mitarbeiter in der Datenbank speichern
	err := h.employeeRepo.Create(employee)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Erstellen des Mitarbeiters: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Zurück zur Mitarbeiterliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/employees?success=added")
}

// GetEmployeeDetails zeigt die Details eines Mitarbeiters an
func (h *EmployeeHandler) GetEmployeeDetails(c *gin.Context) {
	id := c.Param("id")

	// Mitarbeiter anhand der ID abrufen
	employee, err := h.employeeRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Mitarbeiter nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Vorgesetzten des Mitarbeiters abrufen, falls vorhanden
	var manager *model.Employee
	if !employee.ManagerID.IsZero() {
		manager, _ = h.employeeRepo.FindByID(employee.ManagerID.Hex())
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employee_details.html", gin.H{
		"title":    employee.FirstName + " " + employee.LastName,
		"active":   "employees",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"employee": employee,
		"manager":  manager,
	})
}

// UpdateEmployee aktualisiert einen bestehenden Mitarbeiter
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")

	// Mitarbeiter anhand der ID abrufen
	employee, err := h.employeeRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Mitarbeiter nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Formulardaten abrufen und Mitarbeiter aktualisieren
	// (ähnlich wie bei AddEmployee, jedoch mit einem bestehenden Objekt)

	// Mitarbeiter in der Datenbank aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Zurück zur Mitarbeiterliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/employees?success=updated")
}

// DeleteEmployee löscht einen Mitarbeiter
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")

	// Mitarbeiter löschen
	err := h.employeeRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{"message": "Mitarbeiter erfolgreich gelöscht"})
}
