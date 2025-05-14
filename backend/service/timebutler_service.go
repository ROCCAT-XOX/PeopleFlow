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
	// Testen, ob der API-Schlüssel funktioniert
	if err := s.testConnection(apiKey); err != nil {
		return err
	}

	// API-Schlüssel speichern
	if err := s.integrationRepo.SaveApiKey("timebutler", apiKey); err != nil {
		return err
	}

	// Integration als aktiv markieren
	return s.integrationRepo.SetIntegrationStatus("timebutler", true)
}

// testConnection testet die Verbindung zu Timebutler mit dem angegebenen API-Schlüssel
func (s *TimebutlerService) testConnection(apiKey string) error {
	url := "https://app.timebutler.com/api/v1/users"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s", apiKey))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Timebutler API Fehler: %s", res.Status))
	}

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
	apiKey, err := s.integrationRepo.GetApiKey("timebutler")
	if err != nil {
		return "", err
	}

	url := "https://app.timebutler.com/api/v1/absences"
	method := "POST"
	payload := strings.NewReader(fmt.Sprintf("auth=%s&year=%s", apiKey, year))

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
	active, err := s.integrationRepo.GetIntegrationStatus("timebutler")
	if err != nil {
		return false
	}

	// Wenn die Integration aktiv ist, testen wir auch die Verbindung
	if active {
		apiKey, err := s.integrationRepo.GetApiKey("timebutler")
		if err != nil {
			return false
		}

		// Einfacher Verbindungstest
		if err := s.testConnection(apiKey); err != nil {
			// Bei Fehler setzen wir die Integration auf inaktiv
			s.integrationRepo.SetIntegrationStatus("timebutler", false)
			return false
		}
	}

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
	employees, err := employeeRepo.FindAll()
	if err != nil {
		return 0, err
	}

	// Zähler für aktualisierte Mitarbeiter
	updatedCount := 0

	// Mitarbeiter durchgehen und mit Timebutler-Daten abgleichen
	for _, employee := range employees {
		// Prüfen, ob ein Timebutler-Benutzer mit dieser E-Mail existiert
		timebutlerUser, exists := timebutlerUsers[employee.Email]
		if !exists {
			continue
		}

		// Flag, um zu prüfen, ob Änderungen vorgenommen wurden
		updated := false

		// Felder synchronisieren, die aus Timebutler kommen sollen

		// Telefon aktualisieren, wenn nicht gesetzt
		if employee.Phone == "" && timebutlerUser.Phone != "" {
			employee.Phone = timebutlerUser.Phone
			updated = true
		}

		// Abteilung aktualisieren, wenn nicht gesetzt
		if employee.Department == "" && timebutlerUser.Department != "" {
			employee.Department = model.Department(timebutlerUser.Department)
			updated = true
		}

		// Eintrittsdatum aktualisieren, wenn nicht gesetzt
		if employee.HireDate.IsZero() && !timebutlerUser.DateOfEntry.IsZero() {
			employee.HireDate = timebutlerUser.DateOfEntry
			updated = true
		}

		// Geburtsdatum aktualisieren, wenn nicht gesetzt
		if employee.DateOfBirth.IsZero() && !timebutlerUser.DateOfBirth.IsZero() {
			employee.DateOfBirth = timebutlerUser.DateOfBirth
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

// SyncTimebutlerAbsences synchronisiert Timebutler-Abwesenheiten mit PeopleFlow-Mitarbeitern
func (s *TimebutlerService) SyncTimebutlerAbsences(year string) (int, error) {
	// Timebutler-Abwesenheiten abrufen
	absencesMap, err := s.GetTimebutlerAbsences(year)
	if err != nil {
		return 0, err
	}

	// Repository für Mitarbeiter initialisieren
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter abrufen
	employees, err := employeeRepo.FindAll()
	if err != nil {
		return 0, err
	}

	// Zähler für aktualisierte Mitarbeiter
	updatedCount := 0

	// Mitarbeiter durchgehen und Abwesenheiten zuordnen
	for _, employee := range employees {
		// Prüfen, ob Abwesenheiten für diesen Mitarbeiter existieren
		absences, exists := absencesMap[employee.Email]
		if !exists || len(absences) == 0 {
			continue
		}

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
				if strings.Contains(strings.ToLower(absence.AbsenceType), "krank") {
					absenceType = "sick"
				} else if strings.Contains(strings.ToLower(absence.AbsenceType), "special") {
					absenceType = "special"
				}

				// Status bestimmen
				status := "approved" // Standard: Genehmigt
				if strings.ToLower(absence.Status) == "requested" {
					status = "requested"
				} else if strings.ToLower(absence.Status) == "rejected" {
					status = "rejected"
				} else if strings.ToLower(absence.Status) == "cancelled" {
					status = "cancelled"
				}

				// Dauer berechnen (in Tagen)
				days := float64(absence.EndDate.Sub(absence.StartDate).Hours() / 24)
				if days < 0.5 {
					days = 0.5 // Mindestens halber Tag
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
			}
		}

		// Mitarbeiter aktualisieren, wenn Abwesenheiten hinzugefügt wurden
		if abwesenheitenHinzugefuegt {
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
