package handler

import (
	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"
	"PeoplePilot/backend/service"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
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
			profileImage = "" // Leer lassen
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
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		Phone:           c.PostForm("phone"),
		Address:         c.PostForm("address"),
		DateOfBirth:     birthDate,
		HireDate:        hireDate,
		Position:        position,
		Department:      model.Department(department),
		ManagerID:       managerID,
		Status:          model.EmployeeStatusActive, // Standardmäßig aktiv
		Salary:          salary,
		BankAccount:     c.PostForm("iban"),
		TaxID:           c.PostForm("taxClass"),
		SocialSecID:     c.PostForm("socialSecId"),
		HealthInsurance: c.PostForm("healthInsurance"),
		EmergencyName:   c.PostForm("emergencyName"),
		EmergencyPhone:  c.PostForm("emergencyPhone"),
		Notes:           c.PostForm("notes"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeEmployeeAdded,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		employee.ID,
		"employee",
		employee.FirstName+" "+employee.LastName,
		"Neuer Mitarbeiter hinzugefügt",
	)

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
	userRole, _ := c.Get("userRole")

	// Vorgesetzten des Mitarbeiters abrufen, falls vorhanden
	var manager *model.Employee
	if !employee.ManagerID.IsZero() {
		manager, _ = h.employeeRepo.FindByID(employee.ManagerID.Hex())
	}

	// Format Helpers als Template Funktionen
	formatFileSize := func(size int64) string {
		const unit = 1024
		if size < unit {
			return fmt.Sprintf("%d B", size)
		}
		div, exp := int64(unit), 0
		for n := size / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
	}

	iterate := func(count int) []int {
		var i []int
		for j := 0; j < count; j++ {
			i = append(i, j)
		}
		return i
	}

	// Hilfsfunktion für das aktuelle Datum
	now := time.Now()

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employee_detail_advanced.html", gin.H{
		"title":          employee.FirstName + " " + employee.LastName,
		"active":         "employees",
		"user":           userModel.FirstName + " " + userModel.LastName,
		"email":          userModel.Email,
		"year":           time.Now().Year(),
		"employee":       employee,
		"manager":        manager,
		"userRole":       userRole,
		"formatFileSize": formatFileSize,
		"iterate":        iterate,
		"now":            now,
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
	employee.FirstName = c.PostForm("firstName")
	employee.LastName = c.PostForm("lastName")
	employee.Email = c.PostForm("email")
	employee.Phone = c.PostForm("phone")
	employee.Address = c.PostForm("address")
	employee.Position = c.PostForm("position")
	employee.Department = model.Department(c.PostForm("department"))
	employee.Notes = c.PostForm("notes")

	// Status aktualisieren
	statusStr := c.PostForm("status")
	if statusStr != "" {
		employee.Status = model.EmployeeStatus(statusStr)
	}

	// Manager-ID parsen, falls vorhanden
	managerIDStr := c.PostForm("managerId")
	if managerIDStr != "" {
		managerID, err := primitive.ObjectIDFromHex(managerIDStr)
		if err == nil {
			employee.ManagerID = managerID
		}
	} else {
		// Wenn kein Manager ausgewählt ist, setzen wir eine leere ID
		employee.ManagerID = primitive.NilObjectID
	}

	// Datumsfelder parsen
	hireDateStr := c.PostForm("hireDate")
	if hireDateStr != "" {
		hireDate, err := time.Parse("2006-01-02", hireDateStr)
		if err == nil {
			employee.HireDate = hireDate
		}
	}

	birthDateStr := c.PostForm("birthDate")
	if birthDateStr != "" {
		birthDate, err := time.Parse("2006-01-02", birthDateStr)
		if err == nil {
			employee.DateOfBirth = birthDate
		}
	}

	// Finanzielle Daten aktualisieren (nur für Administratoren)
	userRole, _ := c.Get("userRole")
	if userRole == string(model.RoleAdmin) {
		salaryStr := c.PostForm("salary")
		if salaryStr != "" {
			salary, err := strconv.ParseFloat(salaryStr, 64)
			if err == nil {
				employee.Salary = salary
			}
		}

		employee.BankAccount = c.PostForm("bankAccount")
		employee.TaxID = c.PostForm("taxId")
		employee.SocialSecID = c.PostForm("socialSecId")
		employee.SocialSecID = c.PostForm("socialSecId")
		employee.HealthInsurance = c.PostForm("healthInsurance")
	}

	// Notfallkontakt aktualisieren
	employee.EmergencyName = c.PostForm("emergencyName")
	employee.EmergencyPhone = c.PostForm("emergencyPhone")

	// UpdatedAt aktualisieren
	employee.UpdatedAt = time.Now()

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

// ShowEditEmployeeForm zeigt das Formular zum Bearbeiten eines Mitarbeiters an
func (h *EmployeeHandler) ShowEditEmployeeForm(c *gin.Context) {
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

	// Liste der Manager abrufen
	managers, err := h.employeeRepo.FindManagers()
	if err != nil {
		managers = []*model.Employee{} // Leere Liste, falls ein Fehler auftritt
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employee_edit.html", gin.H{
		"title":    "Mitarbeiter bearbeiten",
		"active":   "employees",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"employee": employee,
		"managers": managers,
		"userRole": c.GetString("userRole"),
	})
}

// UploadProfileImage lädt ein Profilbild für einen Mitarbeiter hoch
// UploadProfileImage lädt ein Profilbild hoch und gibt den Dateipfad zurück
// UploadProfileImage lädt ein Profilbild für einen Mitarbeiter hoch
func (h *EmployeeHandler) UploadProfileImage(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Hochgeladene Datei abrufen
	file, err := c.FormFile("profileImage")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Datei hochgeladen"})
		return
	}

	// Überprüfen, ob es sich um ein Bild handelt
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Die hochgeladene Datei ist kein Bild"})
		return
	}

	// Wenn es bereits ein altes Profilbild gibt, dieses löschen
	if employee.ProfileImage != "" && strings.HasPrefix(employee.ProfileImage, "/static/uploads/") {
		oldPath := "." + employee.ProfileImage
		os.Remove(oldPath) // Ignoriere Fehler, falls die Datei nicht existiert
	}

	// Profilbild hochladen
	fileService := service.NewFileService()
	profileImagePath, err := fileService.UploadProfileImage(file, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Hochladen des Profilbilds: " + err.Error()})
		return
	}

	// Pfad zum Profilbild in der Datenbank aktualisieren
	employee.ProfileImage = profileImagePath
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Profilbild erfolgreich hochgeladen",
		"profileImage": profileImagePath,
	})
}
