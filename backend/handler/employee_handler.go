package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"fmt"
	"io"
	"net/http"
	"sort"
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

	// Benutzerrolle aus dem Context abrufen
	userRole, _ := c.Get("userRole")

	// Alle Mitarbeiter abrufen
	employees, _, err := h.employeeRepo.FindAll(0, 1000, "lastName", 1)
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

		// Arbeitszeit-Informationen formatieren
		var workingHours string
		var workTimeModel string
		if emp.WorkingHoursPerWeek > 0 {
			workingHours = fmt.Sprintf("%.1f", emp.WorkingHoursPerWeek)
			workTimeModel = emp.WorkTimeModel.GetDisplayName()
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
			"WorkingHours":      workingHours,
			"WorkTimeModel":     workTimeModel,
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
		"userRole":       userRole, // Hier wird die userRole hinzugefügt
	})
}

// AddEmployee fügt einen neuen Mitarbeiter hinzu
func (h *EmployeeHandler) AddEmployee(c *gin.Context) {
	// Formulardaten abrufen
	firstName := strings.TrimSpace(c.PostForm("firstName"))
	lastName := strings.TrimSpace(c.PostForm("lastName"))
	email := strings.TrimSpace(c.PostForm("email"))
	position := strings.TrimSpace(c.PostForm("position"))
	department := strings.TrimSpace(c.PostForm("department"))
	
	// Grundlegende Validierung
	if firstName == "" || lastName == "" || email == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Vorname, Nachname und E-Mail sind Pflichtfelder",
			"year":    time.Now().Year(),
		})
		return
	}
	
	// Repository-Verfügbarkeit prüfen
	if h.employeeRepo == nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Datenbankverbindung nicht verfügbar",
			"year":    time.Now().Year(),
		})
		return
	}
	
	// Prüfen, ob E-Mail bereits existiert
	if email != "" {
		existingEmployee, err := h.employeeRepo.FindByEmail(email)
		if err == nil && existingEmployee != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"title":   "Fehler",
				"message": "Ein Mitarbeiter mit dieser E-Mail-Adresse existiert bereits",
				"year":    time.Now().Year(),
			})
			return
		}
	}

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

	// Arbeitszeit-Daten verarbeiten
	var workingHoursPerWeek float64
	workingHoursStr := c.PostForm("workingHoursPerWeek")
	if workingHoursStr != "" {
		workingHoursPerWeek, _ = strconv.ParseFloat(workingHoursStr, 64)
	}

	var workingDaysPerWeek int
	workingDaysStr := c.PostForm("workingDaysPerWeek")
	if workingDaysStr != "" {
		workingDaysPerWeek, _ = strconv.Atoi(workingDaysStr)
	}

	flexibleWorkingHours := c.PostForm("flexibleWorkingHours") == "true"

	// EmployeeID generieren (falls nicht vom Formular bereitgestellt)
	employeeID := strings.TrimSpace(c.PostForm("employeeID"))
	if employeeID == "" {
		// Generiere eine Employee-ID basierend auf Namen (als Fallback)
		employeeID = strings.ToUpper(string(firstName[0]) + string(lastName[0])) + fmt.Sprintf("%d", time.Now().Unix()%10000)
	}

	// Neues Employee-Objekt erstellen
	employee := &model.Employee{
		EmployeeID:        employeeID,
		FirstName:         firstName,
		LastName:          lastName,
		Email:             email,
		Phone:             c.PostForm("phone"),
		InternalPhone:     c.PostForm("internalPhone"),
		InternalExtension: c.PostForm("internalExtension"),
		Address:           c.PostForm("address"),
		DateOfBirth:       birthDate,
		HireDate:          hireDate,
		Position:          position,
		Department:        model.Department(department),
		ManagerID:         managerID,
		Status:            model.EmployeeStatusActive,

		// Arbeitszeit-Daten hinzufügen
		WorkingHoursPerWeek:  workingHoursPerWeek,
		WorkingDaysPerWeek:   workingDaysPerWeek,
		WorkTimeModel:        model.WorkTimeModel(c.PostForm("workTimeModel")),
		FlexibleWorkingHours: flexibleWorkingHours,
		CoreWorkingTimeStart: c.PostForm("coreWorkingTimeStart"),
		CoreWorkingTimeEnd:   c.PostForm("coreWorkingTimeEnd"),

		// Bestehende Felder...
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
	user, exists := c.Get("user")
	if exists && user != nil {
		userModel := user.(*model.User)
		
		activityRepo := repository.NewActivityRepository()
		if activityRepo != nil {
			_, _ = activityRepo.LogActivity(
				model.ActivityTypeEmployeeAdded,
				userModel.ID,
				userModel.FirstName+" "+userModel.LastName,
				employee.ID,
				"employee",
				employee.FirstName+" "+employee.LastName,
				"Neuer Mitarbeiter hinzugefügt",
			)
		}
	}

	// Zurück zur Mitarbeiterliste mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/employees?success=added")
}

// GetEmployeeDetails zeigt die Details eines Mitarbeiters an
func (h *EmployeeHandler) GetEmployeeDetails(c *gin.Context) {
	id := c.Param("id")

	hideSalary, exists := c.Get("hideSalary")
	if !exists {
		hideSalary = false
	}

	// Mitarbeiter inklusive Anpassungen laden
	employee, err := h.employeeRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":      "Fehler",
			"message":    "Mitarbeiter nicht gefunden",
			"year":       time.Now().Year(),
			"hideSalary": hideSalary,
		})
		return
	}

	// TimeAccountService für Überstunden-Details initialisieren
	timeAccountService := service.NewTimeAccountService()

	// Wenn noch keine Überstunden berechnet wurden, jetzt berechnen
	if employee.LastTimeCalculated.IsZero() && len(employee.TimeEntries) > 0 {
		err = timeAccountService.CalculateOvertimeForEmployee(employee)
		if err != nil {
			fmt.Printf("Error calculating overtime for employee %s: %v\n", employee.ID.Hex(), err)
		}
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer-Session ungültig",
			"year":    time.Now().Year(),
		})
		return
	}
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

	// Helper-Funktionen für Zeitzone-Konvertierung
	location, _ := time.LoadLocation("Europe/Berlin")

	formatTimeInLocalZone := func(t time.Time) string {
		return t.In(location).Format("15:04")
	}

	formatDateInLocalZone := func(t time.Time) string {
		return t.In(location).Format("02.01.2006")
	}

	formatDateTimeInLocalZone := func(t time.Time) string {
		return t.In(location).Format("02.01.2006 15:04")
	}

	// Calculate used and remaining vacation days for the current year
	var usedVacationDays float64 = 0
	currentYear := time.Now().Year()

	for _, absence := range employee.Absences {
		if absence.Type == "vacation" &&
			absence.Status == "approved" &&
			absence.StartDate.Year() == currentYear {
			usedVacationDays += absence.Days
		}
	}

	// If VacationDays is not set, provide a default
	if employee.VacationDays == 0 {
		employee.VacationDays = 30 // Default value if not set
	}

	// Calculate remaining vacation days if not already set
	if employee.RemainingVacation == 0 {
		employee.RemainingVacation = employee.VacationDays - int(usedVacationDays)
	}

	// Prepare time entries data for the view
	var timeEntries []model.TimeEntry
	var totalHours float64
	var projectMap = make(map[string]float64)
	var startDate time.Time
	var endDate time.Time

	// Sort time entries by date (newest first)
	if len(employee.TimeEntries) > 0 {
		// Make a copy to avoid modifying the original
		timeEntries = make([]model.TimeEntry, len(employee.TimeEntries))

		// WICHTIG: Konvertiere die Zeiten beim Kopieren
		for i, entry := range employee.TimeEntries {
			timeEntries[i] = entry
			// Die Zeiten sind bereits in der richtigen Zeitzone in der DB
			// Wir müssen sie nur für die Anzeige vorbereiten
		}

		// Sort time entries by date (newest first)
		sort.Slice(timeEntries, func(i, j int) bool {
			return timeEntries[i].Date.After(timeEntries[j].Date)
		})

		// Initialize with the first entry's date
		startDate = timeEntries[len(timeEntries)-1].Date
		endDate = timeEntries[0].Date

		// Calculate total hours and project distribution
		for _, entry := range timeEntries {
			totalHours += entry.Duration
			projectMap[entry.ProjectName] += entry.Duration

			// Update start and end dates if needed
			if entry.Date.Before(startDate) {
				startDate = entry.Date
			}
			if entry.Date.After(endDate) {
				endDate = entry.Date
			}
		}
	}

	// Convert project map to arrays for chart
	var projectLabels []string
	var projectHours []float64
	for project, hours := range projectMap {
		projectLabels = append(projectLabels, project)
		projectHours = append(projectHours, hours)
	}

	// Überstunden-Details berechnen
	overtimeDetails := employee.GetOvertimeBalanceWithDetails()

	// Basis-Überstunden-Saldo für Template-Attribute
	baseOvertimeBalance := employee.OvertimeBalance

	// Format total hours with 2 decimal places
	totalHoursFormatted := fmt.Sprintf("%.2f", totalHours)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employee_detail_advanced.html", gin.H{
		"title":                     employee.FirstName + " " + employee.LastName,
		"active":                    "employees",
		"user":                      userModel.FirstName + " " + userModel.LastName,
		"email":                     userModel.Email,
		"year":                      time.Now().Year(),
		"employee":                  employee,
		"manager":                   manager,
		"userRole":                  userRole,
		"formatFileSize":            formatFileSize,
		"iterate":                   iterate,
		"now":                       now,
		"hideSalary":                hideSalary,
		"usedVacationDays":          usedVacationDays,
		"remainingVacation":         employee.RemainingVacation,
		"timeEntries":               timeEntries,
		"totalHours":                totalHoursFormatted,
		"projectCount":              len(projectMap),
		"startDate":                 startDate,
		"endDate":                   endDate,
		"projectLabels":             projectLabels,
		"projectHours":              projectHours,
		"overtimeDetails":           overtimeDetails,
		"baseOvertimeBalance":       baseOvertimeBalance,
		"formatTimeInLocalZone":     formatTimeInLocalZone,
		"formatDateInLocalZone":     formatDateInLocalZone,
		"formatDateTimeInLocalZone": formatDateTimeInLocalZone,
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
	employee.InternalPhone = c.PostForm("internalPhone")
	employee.InternalExtension = c.PostForm("internalExtension")
	employee.Address = c.PostForm("address")
	employee.Position = c.PostForm("position")
	employee.Department = model.Department(c.PostForm("department"))
	employee.Notes = c.PostForm("notes")

	// Arbeitszeit-Daten aktualisieren
	workingHoursStr := c.PostForm("workingHoursPerWeek")
	if workingHoursStr != "" {
		workingHours, err := strconv.ParseFloat(workingHoursStr, 64)
		if err == nil {
			employee.WorkingHoursPerWeek = workingHours
		}
	}

	workingDaysStr := c.PostForm("workingDaysPerWeek")
	if workingDaysStr != "" {
		workingDays, err := strconv.Atoi(workingDaysStr)
		if err == nil {
			employee.WorkingDaysPerWeek = workingDays
		}
	}

	employee.WorkTimeModel = model.WorkTimeModel(c.PostForm("workTimeModel"))
	employee.FlexibleWorkingHours = c.PostForm("flexibleWorkingHours") == "true"
	employee.CoreWorkingTimeStart = c.PostForm("coreWorkingTimeStart")
	employee.CoreWorkingTimeEnd = c.PostForm("coreWorkingTimeEnd")

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

	// Finanzielle Daten aktualisieren (abhängig von den Berechtigungen)
	hideSalary, _ := c.Get("hideSalary")

	if hideSalary == nil || hideSalary == false {
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

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeEmployeeUpdated,
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		employee.ID,
		"employee",
		employee.FirstName+" "+employee.LastName,
		"Mitarbeiter aktualisiert",
	)

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

	// Benutzerrolle und Sichtbarkeit des Gehalts aus dem Context abrufen
	userRole, _ := c.Get("userRole")
	hideSalary, exists := c.Get("hideSalary")
	if !exists {
		hideSalary = false
	}

	// Mitarbeiter anhand der ID abrufen
	employee, err := h.employeeRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":      "Fehler",
			"message":    "Mitarbeiter nicht gefunden",
			"year":       time.Now().Year(),
			"hideSalary": hideSalary,
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer-Session ungültig",
			"year":    time.Now().Year(),
		})
		return
	}
	userModel := user.(*model.User)

	// Liste der Manager abrufen
	managers, err := h.employeeRepo.FindManagers()
	if err != nil {
		managers = []*model.Employee{} // Leere Liste, falls ein Fehler auftritt
	}

	fmt.Printf("Before template: hideSalary=%v\n", hideSalary)

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "employee_edit.html", gin.H{
		"title":      "Mitarbeiter bearbeiten",
		"active":     "employees",
		"user":       userModel.FirstName + " " + userModel.LastName,
		"email":      userModel.Email,
		"year":       time.Now().Year(),
		"employee":   employee,
		"managers":   managers,
		"userRole":   userRole,
		"hideSalary": hideSalary,
	})
}

// UploadProfileImage handles profile image uploads for employees
func (h *EmployeeHandler) UploadProfileImage(c *gin.Context) {
	// Get employee ID from URL parameter
	employeeID := c.Param("id")

	// Retrieve employee from database
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("profileImage")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Datei hochgeladen: " + err.Error()})
		return
	}

	// Check file type
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Die hochgeladene Datei ist kein Bild"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Öffnen der hochgeladenen Datei: " + err.Error()})
		return
	}
	defer src.Close()

	// Read file contents
	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Lesen der Datei: " + err.Error()})
		return
	}

	// Store the image data in the employee object
	employee.ProfileImage = contentType // Store the mime type
	employee.ProfileImageData = primitive.Binary{Data: fileData}

	// Update employee in database
	if err := h.employeeRepo.Update(employee); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profilbild erfolgreich hochgeladen",
	})
}

func (h *EmployeeHandler) ListUpcomingConversations(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, exists := c.Get("user")
	if !exists || user == nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer-Session ungültig",
			"year":    time.Now().Year(),
		})
		return
	}
	userModel := user.(*model.User)

	// Alle Mitarbeiter abrufen
	employees, _, err := h.employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Mitarbeiter: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Liste für Mitarbeiter mit anstehenden Gesprächen
	var employeesWithUpcomingConversations []*model.Employee
	var upcomingReviewsList []map[string]string

	// Aktuelle Zeit für den Vergleich
	now := time.Now()

	// Alle Mitarbeiter durchgehen und nach geplanten Gesprächen in der Zukunft suchen
	for _, emp := range employees {
		hasUpcomingConversation := false
		for _, conv := range emp.Conversations {
			// Nur geplante Gespräche und nur solche, die in der Zukunft liegen
			if conv.Status == "planned" && conv.Date.After(now) {
				// Gespräche, die innerhalb der nächsten 14 Tage stattfinden
				if conv.Date.Before(now.AddDate(0, 0, 14)) {
					hasUpcomingConversation = true
					upcomingReviewsList = append(upcomingReviewsList, map[string]string{
						"EmployeeID":   emp.ID.Hex(),
						"EmployeeName": emp.FirstName + " " + emp.LastName,
						"ReviewType":   conv.Title,
						"Date":         conv.Date.Format("02.01.2006"),
						"Description":  conv.Description,
					})
				}
			}
		}
		if hasUpcomingConversation {
			employeesWithUpcomingConversations = append(employeesWithUpcomingConversations, emp)
		}
	}

	// Sortieren nach Datum (die nächsten zuerst)
	if len(upcomingReviewsList) > 0 {
		sort.Slice(upcomingReviewsList, func(i, j int) bool {
			date1, _ := time.Parse("02.01.2006", upcomingReviewsList[i]["Date"])
			date2, _ := time.Parse("02.01.2006", upcomingReviewsList[j]["Date"])
			return date1.Before(date2)
		})
	}

	// Wir erstellen hier EmployeeViewModel-Strukturen, die für die Anzeige optimiert sind
	var employeeViewModels []gin.H
	for _, emp := range employeesWithUpcomingConversations {
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
	c.HTML(http.StatusOK, "upcoming_conversations.html", gin.H{
		"title":           "Anstehende Gespräche",
		"active":          "employees",
		"user":            userModel.FirstName + " " + userModel.LastName,
		"email":           userModel.Email,
		"year":            time.Now().Year(),
		"employees":       employeeViewModels,
		"totalEmployees":  len(employeesWithUpcomingConversations),
		"upcomingReviews": upcomingReviewsList,
	})
}

// Add this to your employee_handler.go if it's not already there
func (h *EmployeeHandler) GetProfileImage(c *gin.Context) {
	employeeID := c.Param("id")

	// Add debug logging
	fmt.Printf("GetProfileImage called for ID: %s\n", employeeID)

	// Retrieve employee from database
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		fmt.Printf("Error finding employee: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// Check if profile image exists
	if len(employee.ProfileImageData.Data) == 0 {
		fmt.Printf("No profile image data found for employee: %s\n", employeeID)
		c.Status(http.StatusNotFound)
		return
	}

	// Log that we're serving the image
	fmt.Printf("Serving profile image for employee: %s, content type: %s, data length: %d bytes\n",
		employeeID, employee.ProfileImage, len(employee.ProfileImageData.Data))

	// Set appropriate content type
	c.Header("Content-Type", employee.ProfileImage)
	c.Header("Cache-Control", "no-cache")

	// Serve the image data
	c.Data(http.StatusOK, employee.ProfileImage, employee.ProfileImageData.Data)
}

// RecalculateEmployeeOvertime berechnet Überstunden für einen spezifischen Mitarbeiter neu
func (h *EmployeeHandler) RecalculateEmployeeOvertime(c *gin.Context) {
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}

	// TimeAccountService initialisieren
	timeAccountService := service.NewTimeAccountService()

	// Überstunden neu berechnen
	err = timeAccountService.CalculateOvertimeForEmployee(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der Überstunden-Berechnung: " + err.Error()})
		return
	}

	// Aktivität loggen
	user, exists := c.Get("user")
	if exists && user != nil {
		userModel := user.(*model.User)

		activityRepo := repository.NewActivityRepository()
		_, _ = activityRepo.LogActivity(
			model.ActivityTypeEmployeeUpdated,
			userModel.ID,
			userModel.FirstName+" "+userModel.LastName,
			employee.ID,
			"employee",
			employee.FirstName+" "+employee.LastName,
			"Überstunden neu berechnet",
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Überstunden erfolgreich neu berechnet",
		"data": gin.H{
			"overtimeBalance": employee.OvertimeBalance,
			"lastCalculated":  employee.LastTimeCalculated,
		},
	})
}

// GetEmployeeName liefert nur den Namen eines Mitarbeiters (für AJAX-Anfragen)
func GetEmployeeName(c *gin.Context) {
	employeeID := c.Param("id")

	employeeRepo := repository.NewEmployeeRepository()
	employee, err := employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Mitarbeiter nicht gefunden",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"name":    employee.FirstName + " " + employee.LastName,
		"id":      employee.ID.Hex(),
	})
}
