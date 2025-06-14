// backend/service/timebutler_service.go
package service

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"bufio"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TimebutlerService verwaltet die Integration mit Timebutler
type TimebutlerService struct {
	integrationRepo *repository.IntegrationRepository
}

// NewTimebutlerService erstellt einen neuen TimebutlerService
func NewTimebutlerService() *TimebutlerService {
	return &TimebutlerService{
		integrationRepo: repository.NewIntegrationRepository(),
	}
}

// SaveApiKey speichert den Timebutler API-Schlüssel und testet die Verbindung
func (s *TimebutlerService) SaveApiKey(apiKey string) error {
	fmt.Printf("[DEBUG] TimebutlerService.SaveApiKey called with key length: %d\n", len(apiKey))
	
	// Testen, ob der API-Schlüssel funktioniert
	fmt.Println("[DEBUG] Testing API connection...")
	if err := s.testConnection(apiKey); err != nil {
		fmt.Printf("[ERROR] API connection test failed: %v\n", err)
		return fmt.Errorf("API connection test failed: %w", err)
	}
	fmt.Println("[DEBUG] API connection test successful")

	// API-Schlüssel speichern
	fmt.Println("[DEBUG] Saving API key to repository...")
	if err := s.integrationRepo.SaveApiKey("timebutler", apiKey); err != nil {
		fmt.Printf("[ERROR] Failed to save API key to repository: %v\n", err)
		return fmt.Errorf("failed to save API key: %w", err)
	}
	fmt.Println("[DEBUG] API key saved to repository")

	// Integration als aktiv markieren
	fmt.Println("[DEBUG] Setting integration status to active...")
	err := s.integrationRepo.SetIntegrationStatus("timebutler", true)
	if err != nil {
		fmt.Printf("[ERROR] Failed to set integration status: %v\n", err)
		return fmt.Errorf("failed to set integration status: %w", err)
	}
	fmt.Println("[DEBUG] Integration status set to active")
	
	return nil
}

// testConnection testet die Verbindung zu Timebutler mit dem angegebenen API-Schlüssel
func (s *TimebutlerService) testConnection(apiKey string) error {
	url := "https://app.timebutler.com/api/v1/users"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s", apiKey))

	fmt.Printf("[DEBUG] Testing connection to %s\n", url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create HTTP request: %v\n", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	fmt.Println("[DEBUG] Sending HTTP request...")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] HTTP request failed: %v\n", err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer res.Body.Close()

	fmt.Printf("[DEBUG] Received HTTP response: %d %s\n", res.StatusCode, res.Status)

	if res.StatusCode != http.StatusOK {
		// Read response body for more details
		body, _ := io.ReadAll(res.Body)
		fmt.Printf("[ERROR] API returned non-200 status. Body: %s\n", string(body))
		return fmt.Errorf("Timebutler API error %d: %s", res.StatusCode, res.Status)
	}

	fmt.Println("[DEBUG] Connection test successful")
	return nil
}

// GetUsers ruft Benutzer von Timebutler ab
func (s *TimebutlerService) GetUsers() (string, error) {
	apiKey, err := s.integrationRepo.GetApiKey("timebutler")
	if err != nil {
		return "", err
	}

	url := "https://app.timebutler.com/api/v1/users"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s", apiKey))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Timebutler API Fehler: %s", res.Status))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Integration als aktiv markieren
	s.integrationRepo.SetIntegrationStatus("timebutler", true)

	return string(body), nil
}

// GetAbsences ruft Abwesenheiten von Timebutler ab
func (s *TimebutlerService) GetAbsences(year string) (string, error) {
	fmt.Printf("[DEBUG] GetAbsences called for year: %s\n", year)
	
	apiKey, err := s.integrationRepo.GetApiKey("timebutler")
	if err != nil {
		fmt.Printf("[ERROR] Failed to get API key: %v\n", err)
		return "", fmt.Errorf("failed to get API key: %w", err)
	}
	fmt.Printf("[DEBUG] Retrieved API key (length: %d)\n", len(apiKey))

	// URL für den API-Endpunkt für Abwesenheiten
	url := "https://app.timebutler.com/api/v1/absences"

	// Alle erforderlichen Parameter für die Timebutler API
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s&year=%s&detailed=true", apiKey, year))

	fmt.Printf("Requesting Timebutler absences for year: %s\n", year)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Timebutler API Fehler: %s", res.Status))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Integration als aktiv markieren
	s.integrationRepo.SetIntegrationStatus("timebutler", true)

	return string(body), nil
}

// IsConnected prüft, ob die Timebutler-Integration aktiv ist
func (s *TimebutlerService) IsConnected() bool {
	fmt.Println("[DEBUG] Checking if Timebutler is connected...")
	
	active, err := s.integrationRepo.GetIntegrationStatus("timebutler")
	if err != nil {
		fmt.Printf("[ERROR] Failed to get integration status: %v\n", err)
		return false
	}
	fmt.Printf("[DEBUG] Integration status from DB: %v\n", active)

	// Wenn die Integration aktiv ist, testen wir auch die Verbindung
	if active {
		fmt.Println("[DEBUG] Integration is active, testing API connection...")
		apiKey, err := s.integrationRepo.GetApiKey("timebutler")
		if err != nil {
			fmt.Printf("[ERROR] Failed to get API key: %v\n", err)
			return false
		}
		fmt.Printf("[DEBUG] Retrieved API key (length: %d)\n", len(apiKey))

		// Einfacher Verbindungstest
		if err := s.testConnection(apiKey); err != nil {
			fmt.Printf("[ERROR] Connection test failed, marking integration as inactive: %v\n", err)
			// Bei Fehler setzen wir die Integration auf inaktiv
			s.integrationRepo.SetIntegrationStatus("timebutler", false)
			return false
		}
		fmt.Println("[DEBUG] Connection test passed")
	}

	fmt.Printf("[DEBUG] Returning connection status: %v\n", active)
	return active
}

// GetApiKey ruft den gespeicherten API-Schlüssel ab
func (s *TimebutlerService) GetApiKey() (string, error) {
	return s.integrationRepo.GetApiKey("timebutler")
}

// ParseTimebutlerUsers parst die CSV-Daten von Timebutler und gibt eine Map mit E-Mail als Schlüssel zurück
func (s *TimebutlerService) ParseTimebutlerUsers(data string) (map[string]model.TimebutlerUser, error) {
	userMap := make(map[string]model.TimebutlerUser)

	scanner := bufio.NewScanner(strings.NewReader(data))
	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()

		// Header-Zeile überspringen
		if isFirstLine {
			isFirstLine = false
			continue
		}

		// Leere Zeilen überspringen
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Zeile nach Semikolon aufteilen
		fields := strings.Split(line, ";")

		// Prüfen, ob genügend Felder vorhanden sind
		if len(fields) < 17 {
			continue
		}

		// E-Mail-Adresse extrahieren (Index 4)
		email := strings.TrimSpace(fields[4])
		if email == "" {
			continue
		}

		// Manager-IDs parsen
		var managerIDs []string
		if len(fields) > 12 && fields[12] != "" {
			managerIDs = strings.Split(fields[12], ",")
			for i, id := range managerIDs {
				managerIDs[i] = strings.TrimSpace(id)
			}
		}

		// Datum parsen
		entryDate := parseTimebutlerDate(fields[15])
		separationDate := parseTimebutlerDate(fields[16])
		birthDate := parseTimebutlerDate(fields[17])

		// TimebutlerUser erstellen
		user := model.TimebutlerUser{
			UserID:                      strings.TrimSpace(fields[0]),
			LastName:                    strings.TrimSpace(fields[1]),
			FirstName:                   strings.TrimSpace(fields[2]),
			EmployeeNumber:              strings.TrimSpace(fields[3]),
			EmailAddress:                email,
			Phone:                       strings.TrimSpace(fields[5]),
			MobilePhone:                 strings.TrimSpace(fields[6]),
			CostCenter:                  strings.TrimSpace(fields[7]),
			BranchOffice:                strings.TrimSpace(fields[8]),
			Department:                  strings.TrimSpace(fields[9]),
			UserType:                    strings.TrimSpace(fields[10]),
			Language:                    strings.TrimSpace(fields[11]),
			ManagerIDs:                  managerIDs,
			UserAccountLocked:           strings.TrimSpace(fields[13]) == "true",
			AdditionalInformation:       strings.TrimSpace(fields[14]),
			DateOfEntry:                 entryDate,
			DateOfSeparationFromCompany: separationDate,
			DateOfBirth:                 birthDate,
		}

		// In die Map einfügen
		userMap[email] = user
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return userMap, nil
}

// parseTimebutlerDate parst ein Datum im Format dd/mm/yyyy
func parseTimebutlerDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{} // Leeres Datum zurückgeben
	}

	// Datum im Format dd/mm/yyyy parsen
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return time.Time{} // Bei Fehler leeres Datum zurückgeben
	}

	return t
}

// SyncTimebutlerUsers synchronisiert Timebutler-Benutzer mit PeopleFlow-Mitarbeitern
func (s *TimebutlerService) SyncTimebutlerUsers() (int, error) {
	// Timebutler-Benutzer abrufen
	usersData, err := s.GetUsers()
	if err != nil {
		return 0, err
	}

	// Timebutler-Benutzer parsen
	timebutlerUsers, err := s.ParseTimebutlerUsers(usersData)
	if err != nil {
		return 0, err
	}

	// Repository für Mitarbeiter initialisieren
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter abrufen
	employees, _, err := employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		return 0, err
	}

	// Zähler für aktualisierte Mitarbeiter
	updatedCount := 0

	// Mitarbeiter durchgehen und mit Timebutler-Daten abgleichen
	for _, employee := range employees {
		// E-Mail in Kleinbuchstaben umwandeln
		employeeEmail := strings.ToLower(employee.Email)

		// Prüfen, ob ein Timebutler-Benutzer mit dieser E-Mail existiert
		// Wir suchen nach der E-Mail in lowercase
		var matchedUser model.TimebutlerUser
		found := false

		for email, tbUser := range timebutlerUsers {
			if strings.ToLower(email) == employeeEmail {
				matchedUser = tbUser
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Flag, um zu prüfen, ob Änderungen vorgenommen wurden
		updated := false

		// Timebutler UserID hinzufügen oder aktualisieren
		if employee.TimebutlerUserID != matchedUser.UserID {
			employee.TimebutlerUserID = matchedUser.UserID
			updated = true
		}

		// Weitere Felder synchronisieren, die aus Timebutler kommen sollen

		// Telefon aktualisieren, wenn nicht gesetzt
		if employee.Phone == "" && matchedUser.Phone != "" {
			employee.Phone = matchedUser.Phone
			updated = true
		}

		// Abteilung aktualisieren, wenn nicht gesetzt
		if employee.Department == "" && matchedUser.Department != "" {
			employee.Department = model.Department(matchedUser.Department)
			updated = true
		}

		// Eintrittsdatum aktualisieren, wenn nicht gesetzt
		if employee.HireDate.IsZero() && !matchedUser.DateOfEntry.IsZero() {
			employee.HireDate = matchedUser.DateOfEntry
			updated = true
		}

		// Geburtsdatum aktualisieren, wenn nicht gesetzt
		if employee.DateOfBirth.IsZero() && !matchedUser.DateOfBirth.IsZero() {
			employee.DateOfBirth = matchedUser.DateOfBirth
			updated = true
		}

		// Wenn Änderungen vorgenommen wurden, Mitarbeiter aktualisieren
		if updated {
			employee.UpdatedAt = time.Now()
			err := employeeRepo.Update(employee)
			if err != nil {
				return updatedCount, err
			}
			updatedCount++
		}
	}

	return updatedCount, nil
}

// GetTimebutlerAbsences ruft Abwesenheiten von Timebutler ab und ordnet sie Mitarbeitern zu
func (s *TimebutlerService) GetTimebutlerAbsences(year string) (map[string][]model.TimebutlerAbsence, error) {
	// Abwesenheiten von Timebutler abrufen
	absencesData, err := s.GetAbsences(year)
	if err != nil {
		return nil, err
	}

	// Ergebnis-Map initialisieren (E-Mail als Schlüssel)
	absencesMap := make(map[string][]model.TimebutlerAbsence)

	// CSV-Daten parsen
	scanner := bufio.NewScanner(strings.NewReader(absencesData))
	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()

		// Header-Zeile überspringen
		if isFirstLine {
			isFirstLine = false
			continue
		}

		// Leere Zeilen überspringen
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Zeile nach Semikolon aufteilen
		fields := strings.Split(line, ";")

		// Prüfen, ob genügend Felder vorhanden sind
		if len(fields) < 6 {
			continue
		}

		// E-Mail-Adresse extrahieren
		email := strings.TrimSpace(fields[1])
		if email == "" {
			continue
		}

		// Datum parsen
		startDate := parseTimebutlerDate(fields[2])
		endDate := parseTimebutlerDate(fields[3])

		// Wenn Datum nicht geparst werden konnte, überspringen
		if startDate.IsZero() || endDate.IsZero() {
			continue
		}

		// TimebutlerAbsence erstellen
		absence := model.TimebutlerAbsence{
			UserID:       strings.TrimSpace(fields[0]),
			EmailAddress: email,
			StartDate:    startDate,
			EndDate:      endDate,
			AbsenceType:  strings.TrimSpace(fields[4]),
			Status:       strings.TrimSpace(fields[5]),
			Comment:      strings.TrimSpace(fields[6]),
		}

		// Zur Map hinzufügen
		absencesMap[email] = append(absencesMap[email], absence)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return absencesMap, nil
}

// ParseTimebutlerAbsences parst die CSV-Daten von Timebutler-Abwesenheiten
func (s *TimebutlerService) ParseTimebutlerAbsences(data string) (map[string][]model.TimebutlerAbsence, error) {
	// Ergebnis-Map initialisieren (UserID als Schlüssel)
	absencesMap := make(map[string][]model.TimebutlerAbsence)

	// CSV-Daten parsen
	scanner := bufio.NewScanner(strings.NewReader(data))
	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()

		// Header-Zeile überspringen
		if isFirstLine {
			isFirstLine = false
			continue
		}

		// Leere Zeilen überspringen
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Zeile nach Semikolon aufteilen
		fields := strings.Split(line, ";")

		// Prüfen, ob genügend Felder vorhanden sind
		if len(fields) < 15 {
			fmt.Printf("Skipping line with insufficient fields: %s\n", line)
			continue
		}

		// Felder extrahieren (basierend auf der bereitgestellten Struktur)
		absenceID := strings.TrimSpace(fields[0])
		startDateStr := strings.TrimSpace(fields[1])
		endDateStr := strings.TrimSpace(fields[2])
		isHalfDayStr := strings.TrimSpace(fields[3])
		isMorningStr := strings.TrimSpace(fields[4])
		userID := strings.TrimSpace(fields[5])
		employeeNumber := strings.TrimSpace(fields[6])
		absenceType := strings.TrimSpace(fields[7])
		isExtraVacationDayStr := strings.TrimSpace(fields[8])
		status := strings.TrimSpace(fields[9])
		substituteState := strings.TrimSpace(fields[10])
		workdaysStr := strings.TrimSpace(fields[11])
		hoursStr := strings.TrimSpace(fields[12])
		medicalCertificate := strings.TrimSpace(fields[13])
		comment := strings.TrimSpace(fields[14])
		substituteUserID := ""
		if len(fields) > 15 {
			substituteUserID = strings.TrimSpace(fields[15])
		}

		// UserID muss vorhanden sein
		if userID == "" {
			fmt.Printf("Skipping line with empty UserID: %s\n", line)
			continue
		}

		// Datum parsen
		startDate := parseTimebutlerDate(startDateStr)
		endDate := parseTimebutlerDate(endDateStr)

		// Wenn Datum nicht geparst werden konnte, überspringen
		if startDate.IsZero() || endDate.IsZero() {
			fmt.Printf("Skipping line with invalid dates: %s\n", line)
			continue
		}

		// Boolean-Werte parsen
		isHalfDay := strings.ToLower(isHalfDayStr) == "true"
		isMorning := strings.ToLower(isMorningStr) == "true"
		isExtraVacationDay := strings.ToLower(isExtraVacationDayStr) == "true"

		// Numerische Werte parsen
		workdays := 0.0
		if workdaysStr != "" {
			workdays, _ = strconv.ParseFloat(workdaysStr, 64)
		}

		hours := 0.0
		if hoursStr != "" {
			hours, _ = strconv.ParseFloat(hoursStr, 64)
		}

		// TimebutlerAbsence erstellen
		absence := model.TimebutlerAbsence{
			ID:                 absenceID,
			UserID:             userID,
			StartDate:          startDate,
			EndDate:            endDate,
			IsHalfDay:          isHalfDay,
			IsMorning:          isMorning,
			EmployeeNumber:     employeeNumber,
			AbsenceType:        absenceType,
			IsExtraVacationDay: isExtraVacationDay,
			Status:             status,
			SubstituteState:    substituteState,
			Workdays:           workdays,
			Hours:              hours,
			MedicalCertificate: medicalCertificate,
			Comment:            comment,
			SubstituteUserID:   substituteUserID,
		}

		// Zur Map hinzufügen
		absencesMap[userID] = append(absencesMap[userID], absence)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return absencesMap, nil
}

// SyncTimebutlerAbsences synchronizes Timebutler absences with PeopleFlow employees
func (s *TimebutlerService) SyncTimebutlerAbsences(year string) (int, error) {
	fmt.Printf("[DEBUG] SyncTimebutlerAbsences called for year: %s\n", year)
	
	// Timebutler-Abwesenheiten abrufen
	fmt.Println("[DEBUG] Fetching absences from Timebutler API...")
	absencesData, err := s.GetAbsences(year)
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch absences: %v\n", err)
		return 0, fmt.Errorf("failed to fetch absences: %w", err)
	}
	fmt.Printf("[DEBUG] Successfully fetched absences data (length: %d)\n", len(absencesData))

	// Logging für Debugging
	fmt.Println("Received absence data from Timebutler API. First 500 chars:")
	if len(absencesData) > 500 {
		fmt.Println(absencesData[:500] + "...")
	} else {
		fmt.Println(absencesData)
	}

	// Absences nach UserID parsen
	absencesByUserID, err := s.ParseTimebutlerAbsences(absencesData)
	if err != nil {
		return 0, err
	}

	// Logging für Debugging
	fmt.Printf("Parsed %d unique user IDs with absences\n", len(absencesByUserID))
	for userID, absences := range absencesByUserID {
		fmt.Printf("UserID: %s has %d absences\n", userID, len(absences))
	}

	// Repository für Mitarbeiter initialisieren
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter abrufen
	employees, _, err := employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		return 0, err
	}

	// Logging für Debugging
	fmt.Printf("Found %d employees in the database\n", len(employees))
	for _, emp := range employees {
		if emp.TimebutlerUserID != "" {
			fmt.Printf("Employee %s %s has Timebutler UserID: %s\n", emp.FirstName, emp.LastName, emp.TimebutlerUserID)
		}
	}

	// Standardwert für Urlaubstage pro Jahr
	const standardVacationDays = 30

	// Zähler für aktualisierte Mitarbeiter
	updatedCount := 0

	// Mitarbeiter durchgehen und Abwesenheiten zuordnen
	for _, employee := range employees {
		// Prüfen, ob TimebutlerUserID gesetzt ist
		if employee.TimebutlerUserID == "" {
			continue
		}

		// Prüfen, ob Abwesenheiten für diesen Mitarbeiter existieren
		absences, exists := absencesByUserID[employee.TimebutlerUserID]
		if !exists || len(absences) == 0 {
			continue
		}

		fmt.Printf("Processing %d absences for employee %s %s (ID: %s)\n",
			len(absences), employee.FirstName, employee.LastName, employee.TimebutlerUserID)

		// Abwesenheiten zu Mitarbeiter hinzufügen
		abwesenheitenHinzugefuegt := false

		for _, absence := range absences {
			// Prüfen, ob die Abwesenheit bereits existiert
			alreadyExists := false
			for _, existingAbsence := range employee.Absences {
				if existingAbsence.StartDate.Equal(absence.StartDate) &&
					existingAbsence.EndDate.Equal(absence.EndDate) {
					alreadyExists = true
					break
				}
			}

			// Wenn die Abwesenheit nicht existiert, hinzufügen
			if !alreadyExists {
				// Abwesenheitstyp bestimmen
				absenceType := "vacation" // Standard: Urlaub
				if strings.Contains(strings.ToLower(absence.AbsenceType), "sick") ||
					strings.Contains(strings.ToLower(absence.AbsenceType), "krank") {
					absenceType = "sick"
				} else if strings.Contains(strings.ToLower(absence.AbsenceType), "special") {
					absenceType = "special"
				}

				// Status bestimmen
				status := "approved" // Standard: Genehmigt
				if strings.ToLower(absence.Status) == "requested" {
					status = "requested"
				} else if strings.ToLower(absence.Status) == "rejected" ||
					strings.ToLower(absence.Status) == "declined" {
					status = "rejected"
				} else if strings.ToLower(absence.Status) == "cancelled" {
					status = "cancelled"
				}

				// Arbeitstage verwenden, wenn verfügbar, sonst berechnen
				days := absence.Workdays
				if days < 0.1 {
					// Wenn Workdays nicht gesetzt, berechnen wir die Tage
					days = float64(absence.EndDate.Sub(absence.StartDate).Hours() / 24)
					if absence.IsHalfDay {
						days = 0.5
					}
					if days < 0.5 {
						days = 0.5 // Mindestens halber Tag
					}
				}

				// Neue Abwesenheit erstellen
				newAbsence := model.Absence{
					ID:        primitive.NewObjectID(),
					Type:      absenceType,
					StartDate: absence.StartDate,
					EndDate:   absence.EndDate,
					Days:      days,
					Status:    status,
					Reason:    absence.AbsenceType,
					Notes:     absence.Comment,
				}

				// Zur Mitarbeiterabsenzenliste hinzufügen
				employee.Absences = append(employee.Absences, newAbsence)
				abwesenheitenHinzugefuegt = true

				fmt.Printf("Added new absence: %s from %s to %s for employee %s %s\n",
					absence.AbsenceType, absence.StartDate.Format("2006-01-02"),
					absence.EndDate.Format("2006-01-02"), employee.FirstName, employee.LastName)
			}
		}

		// Calculate total vacation days and remaining vacation
		// Aktuelle Jahr berechnen
		currentYear := time.Now().Year()
		var usedVacationDays float64 = 0

		// Berechnen, wie viele Urlaubstage im aktuellen Jahr genommen wurden
		for _, absence := range employee.Absences {
			// Nur genehmigte Urlaubstage (nicht Krankheit oder Sonderurlaub) im aktuellen Jahr zählen
			if absence.Type == "vacation" &&
				absence.Status == "approved" &&
				absence.StartDate.Year() == currentYear {
				usedVacationDays += absence.Days
			}
		}

		// VacationDays-Feld setzen, wenn noch nicht gesetzt
		if employee.VacationDays == 0 {
			employee.VacationDays = standardVacationDays
		}

		// Verbleibende Urlaubstage berechnen (mit Schutz vor negativen Werten)
		remainingVacation := employee.VacationDays - int(usedVacationDays)
		if remainingVacation < 0 {
			employee.RemainingVacation = 0
			fmt.Printf("Warning: Clamped negative remaining vacation days (%d) to 0 for employee %s %s\n", 
				remainingVacation, employee.FirstName, employee.LastName)
		} else {
			employee.RemainingVacation = remainingVacation
		}

		// Mitarbeiter aktualisieren, wenn Abwesenheiten hinzugefügt wurden oder Urlaubstage berechnet wurden
		if abwesenheitenHinzugefuegt || employee.VacationDays > 0 {
			employee.UpdatedAt = time.Now()
			err := employeeRepo.Update(employee)
			if err != nil {
				return updatedCount, err
			}
			updatedCount++
			fmt.Printf("Updated employee %s %s with absences and vacation information\n",
				employee.FirstName, employee.LastName)
		}
	}

	return updatedCount, nil
}

// Add this to backend/service/timebutler_service.go

// GetHolidayEntitlements fetches holiday entitlement data from Timebutler
func (s *TimebutlerService) GetHolidayEntitlements(year string) (string, error) {
	apiKey, err := s.integrationRepo.GetApiKey("timebutler")
	if err != nil {
		return "", err
	}

	// URL for the API endpoint for holiday entitlements
	url := "https://app.timebutler.com/api/v1/holidayentitlement"

	// All required parameters for the Timebutler API
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s&year=%s", apiKey, year))

	fmt.Printf("Requesting Timebutler holiday entitlements for year: %s\n", year)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Timebutler API Error: %s", res.Status))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// Mark integration as active
	s.integrationRepo.SetIntegrationStatus("timebutler", true)

	return string(body), nil
}

// ParseHolidayEntitlements parses the CSV data from Timebutler holiday entitlements
func (s *TimebutlerService) ParseHolidayEntitlements(data string) (map[string]VacationInfo, error) {
	entitlementMap := make(map[string]VacationInfo)

	scanner := bufio.NewScanner(strings.NewReader(data))
	isFirstLine := true

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header line
		if isFirstLine {
			isFirstLine = false
			continue
		}

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split line by semicolon
		fields := strings.Split(line, ";")

		// Check if there are enough fields (at least 3 for userId, total, and remaining)
		if len(fields) < 3 {
			continue
		}

		// Extract userID, total vacation, and remaining vacation
		userID := strings.TrimSpace(fields[0])
		totalVacationStr := strings.TrimSpace(fields[1])
		remainingVacationStr := strings.TrimSpace(fields[2])

		// Parse vacation values to float then to int
		totalVacation, err := strconv.ParseFloat(totalVacationStr, 64)
		if err != nil {
			// If parsing fails, skip this entry
			continue
		}

		remainingVacation, err := strconv.ParseFloat(remainingVacationStr, 64)
		if err != nil {
			// If parsing fails, skip this entry
			continue
		}

		// Store in map with userID as key
		entitlementMap[userID] = VacationInfo{
			TotalVacationDays:     int(totalVacation),
			RemainingVacationDays: int(remainingVacation),
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entitlementMap, nil
}

// VacationInfo holds vacation data
type VacationInfo struct {
	TotalVacationDays     int
	RemainingVacationDays int
}

// SyncHolidayEntitlements synchronizes Timebutler holiday entitlements with PeopleFlow employees
func (s *TimebutlerService) SyncHolidayEntitlements(year string) (int, error) {
	// Fetch holiday entitlements from Timebutler
	entitlementsData, err := s.GetHolidayEntitlements(year)
	if err != nil {
		return 0, err
	}

	// Parse entitlements data
	entitlementsByUserID, err := s.ParseHolidayEntitlements(entitlementsData)
	if err != nil {
		return 0, err
	}

	// Log for debugging
	fmt.Printf("Parsed %d holiday entitlements\n", len(entitlementsByUserID))

	// Repository for employees
	employeeRepo := repository.NewEmployeeRepository()

	// Get all employees
	employees, _, err := employeeRepo.FindAll(0, 1000, "lastName", 1)
	if err != nil {
		return 0, err
	}

	// Counter for updated employees
	updatedCount := 0

	// Go through employees and update vacation days
	for _, employee := range employees {
		// Skip if no Timebutler UserID is set
		if employee.TimebutlerUserID == "" {
			continue
		}

		// Check if entitlement exists for this employee
		vacationInfo, exists := entitlementsByUserID[employee.TimebutlerUserID]
		if !exists {
			continue
		}

		// Check if vacation data needs updating
		if employee.VacationDays != vacationInfo.TotalVacationDays ||
			employee.RemainingVacation != vacationInfo.RemainingVacationDays {

			// Update vacation days with validation
			employee.VacationDays = vacationInfo.TotalVacationDays
			
			// Ensure remaining vacation is not negative (clamp to 0)
			if vacationInfo.RemainingVacationDays < 0 {
				employee.RemainingVacation = 0
				fmt.Printf("Warning: Clamped negative remaining vacation days (%d) to 0 for employee %s %s\n", 
					vacationInfo.RemainingVacationDays, employee.FirstName, employee.LastName)
			} else {
				employee.RemainingVacation = vacationInfo.RemainingVacationDays
			}

			// Update employee
			employee.UpdatedAt = time.Now()
			err := employeeRepo.Update(employee)
			if err != nil {
				return updatedCount, err
			}
			updatedCount++

			fmt.Printf("Updated employee %s %s with vacation days: total=%d, remaining=%d\n",
				employee.FirstName, employee.LastName, vacationInfo.TotalVacationDays, vacationInfo.RemainingVacationDays)
		}
	}

	return updatedCount, nil
}
