// backend/service/erfasst123_service.go
package service

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Erfasst123Service verwaltet die Integration mit 123erfasst
type Erfasst123Service struct {
	integrationRepo *repository.IntegrationRepository
}

// NewErfasst123Service erstellt einen neuen Erfasst123Service
func NewErfasst123Service() *Erfasst123Service {
	return &Erfasst123Service{
		integrationRepo: repository.NewIntegrationRepository(),
	}
}

// getGermanLocation gibt die deutsche Zeitzone zurück
func getGermanLocation() *time.Location {
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		// Fallback auf CET/CEST
		location = time.FixedZone("CET", 1*60*60) // UTC+1
	}
	return location
}

// SaveCredentials speichert die 123erfasst Anmeldedaten und testet die Verbindung
func (s *Erfasst123Service) SaveCredentials(email, password, syncStartDate string) error {
	// Testen, ob die Anmeldedaten funktionieren
	if err := s.testConnection(email, password); err != nil {
		return err
	}

	// Anmeldedaten zusammen speichern (werden im Repository verschlüsselt)
	credentials := fmt.Sprintf("%s:%s", email, password)
	if err := s.integrationRepo.SaveApiKey("123erfasst", credentials); err != nil {
		return err
	}

	// Automatische Synchronisierung aktivieren
	if err := s.integrationRepo.SetMetadata("123erfasst", "auto_sync", "true"); err != nil {
		return err
	}

	// Startdatum für die Synchronisierung speichern, falls angegeben
	if syncStartDate != "" {
		// Validiere das Datumsformat (YYYY-MM-DD)
		_, err := time.Parse("2006-01-02", syncStartDate)
		if err != nil {
			return fmt.Errorf("ungültiges Startdatum format, verwende YYYY-MM-DD: %v", err)
		}

		if err := s.integrationRepo.SetMetadata("123erfasst", "sync_start_date", syncStartDate); err != nil {
			return err
		}
	} else {
		// Falls kein Startdatum angegeben, Beginn des aktuellen Jahres verwenden
		startOfYear := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		if err := s.integrationRepo.SetMetadata("123erfasst", "sync_start_date", startOfYear); err != nil {
			return err
		}
	}

	// Integration als aktiv markieren
	return s.integrationRepo.SetIntegrationStatus("123erfasst", true)
}

// testConnection testet die Verbindung zu 123erfasst mit den angegebenen Anmeldedaten
func (s *Erfasst123Service) testConnection(email, password string) error {
	// Basic Auth Token erstellen
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// GraphQL-Anfrage für einen einfachen Test (z.B. Anzahl der Personen)
	query := `
	{
		"query": "query { persons { totalCount } }",
		"variables": {}
	}`

	// HTTP-Anfrage
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://server.123erfasst.de/api/graphql", bytes.NewBufferString(query))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Prüfen ob Anfrage erfolgreich war (Status-Code 200)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("123erfasst API Fehler: %s", res.Status)
	}

	return nil
}

// GetCredentials ruft die gespeicherten Anmeldedaten ab
func (s *Erfasst123Service) GetCredentials() (string, string, error) {
	credentials, err := s.integrationRepo.GetApiKey("123erfasst")
	if err != nil {
		return "", "", err
	}

	// Trennen von E-Mail und Passwort
	parts := strings.Split(credentials, ":")
	if len(parts) != 2 {
		return "", "", errors.New("ungültiges Anmeldedatenformat")
	}

	return parts[0], parts[1], nil
}

// IsConnected prüft, ob die 123erfasst-Integration aktiv ist
func (s *Erfasst123Service) IsConnected() bool {
	active, err := s.integrationRepo.GetIntegrationStatus("123erfasst")
	if err != nil {
		return false
	}

	// Wenn die Integration aktiv ist, testen wir auch die Verbindung
	if active {
		email, password, err := s.GetCredentials()
		if err != nil {
			return false
		}

		// Einfacher Verbindungstest
		if err := s.testConnection(email, password); err != nil {
			// Bei Fehler setzen wir die Integration auf inaktiv
			s.integrationRepo.SetIntegrationStatus("123erfasst", false)
			return false
		}
	}

	return active
}

// IsAutoSyncEnabled prüft, ob die automatische Synchronisierung aktiviert ist
func (s *Erfasst123Service) IsAutoSyncEnabled() (bool, error) {
	autoSync, err := s.integrationRepo.GetMetadata("123erfasst", "auto_sync")
	if err != nil {
		return false, err
	}

	return autoSync == "true", nil
}

// SetAutoSync aktiviert oder deaktiviert die automatische Synchronisierung
func (s *Erfasst123Service) SetAutoSync(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}

	return s.integrationRepo.SetMetadata("123erfasst", "auto_sync", value)
}

// GetSyncStartDate holt das gespeicherte Startdatum für die Synchronisierung
func (s *Erfasst123Service) GetSyncStartDate() (string, error) {
	date, err := s.integrationRepo.GetMetadata("123erfasst", "sync_start_date")
	if err != nil || date == "" {
		// Falls kein Datum gespeichert ist, Beginn des aktuellen Jahres zurückgeben
		return time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), nil
	}

	return date, nil
}

// SetSyncStartDate setzt das Startdatum für die Synchronisierung
func (s *Erfasst123Service) SetSyncStartDate(date string) error {
	// Validiere das Datumsformat (YYYY-MM-DD)
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("ungültiges Startdatum format, verwende YYYY-MM-DD: %v", err)
	}

	return s.integrationRepo.SetMetadata("123erfasst", "sync_start_date", date)
}

// GetLastSyncTime holt den Zeitstempel der letzten Synchronisierung
func (s *Erfasst123Service) GetLastSyncTime() (time.Time, error) {
	return s.integrationRepo.GetLastSync("123erfasst")
}

// GetSyncStatus gibt den aktuellen Synchronisationsstatus zurück
func (s *Erfasst123Service) GetSyncStatus() (map[string]interface{}, error) {
	lastSync, err := s.integrationRepo.GetLastSync("123erfasst")
	if err != nil {
		lastSync = time.Time{}
	}

	autoSync, err := s.IsAutoSyncEnabled()
	if err != nil {
		autoSync = false
	}

	syncStartDate, err := s.GetSyncStartDate()
	if err != nil {
		syncStartDate = ""
	}

	return map[string]interface{}{
		"lastSync":      lastSync,
		"autoSync":      autoSync,
		"syncStartDate": syncStartDate,
		"isConnected":   s.IsConnected(),
	}, nil
}

// GetEmployees ruft Mitarbeiter von 123erfasst ab
func (s *Erfasst123Service) GetEmployees() ([]model.Erfasst123Person, error) {
	email, password, err := s.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Basic Auth Token erstellen
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// GraphQL-Anfrage für Mitarbeiterdaten
	query := `{
		"query": "query GetAllEmployeesDetailed { persons { nodes { ident firstname lastname mail employee { isActive hireDate exitDate } } totalCount } }",
		"variables": {}
	}`

	// HTTP-Anfrage
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://server.123erfasst.de/api/graphql", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("123erfasst API Fehler: %s", res.Status)
	}

	// Antwort parsen
	var response model.Erfasst123Response
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	location := getGermanLocation()

	// Datum konvertieren
	for i, person := range response.Data.Persons.Nodes {
		// HireDate
		if person.Employee.HireDate != "" {
			hireDate, err := time.ParseInLocation("2006-01-02", person.Employee.HireDate, location)
			if err != nil {
				fmt.Printf("Warnung: Fehler beim Parsen des Eintrittsdatums für %s %s: %v\n",
					person.Firstname, person.Lastname, err)
			} else {
				response.Data.Persons.Nodes[i].Employee.HireDateParsed = hireDate
			}
		}

		// ExitDate (kann null sein)
		if person.Employee.ExitDate != nil && *person.Employee.ExitDate != "" {
			exitDate, err := time.ParseInLocation("2006-01-02", *person.Employee.ExitDate, location)
			if err != nil {
				fmt.Printf("Warnung: Fehler beim Parsen des Austrittsdatums für %s %s: %v\n",
					person.Firstname, person.Lastname, err)
			} else {
				exitDateParsed := exitDate
				response.Data.Persons.Nodes[i].Employee.ExitDateParsed = &exitDateParsed
			}
		}
	}

	// Integration als aktiv markieren
	s.integrationRepo.SetIntegrationStatus("123erfasst", true)

	// Letzte Synchronisierung aktualisieren
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return response.Data.Persons.Nodes, nil
}

// SyncErfasst123Employees synchronisiert 123erfasst-Mitarbeiter mit PeopleFlow-Mitarbeitern
func (s *Erfasst123Service) SyncErfasst123Employees() (int, error) {
	// 123erfasst-Mitarbeiter abrufen
	employees, err := s.GetEmployees()
	if err != nil {
		return 0, err
	}

	// Repository für Mitarbeiter initialisieren
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter aus der Datenbank abrufen
	peopleFlowEmployees, err := employeeRepo.FindAll()
	if err != nil {
		return 0, err
	}

	// Zähler für aktualisierte Mitarbeiter
	updatedCount := 0

	// Nur aktive Mitarbeiter berücksichtigen
	var activeEmployees []model.Erfasst123Person
	for _, emp := range employees {
		if emp.Employee.IsActive {
			activeEmployees = append(activeEmployees, emp)
		}
	}

	// Logging für Debugging
	fmt.Printf("Gefunden: %d aktive Mitarbeiter in 123erfasst\n", len(activeEmployees))

	// Mitarbeiter durchgehen und mit 123erfasst-Daten abgleichen
	for _, employee := range peopleFlowEmployees {
		// E-Mail in Kleinbuchstaben umwandeln
		employeeEmail := strings.ToLower(employee.Email)

		// Prüfen, ob ein 123erfasst-Mitarbeiter mit dieser E-Mail existiert
		var matchedEmployee model.Erfasst123Person
		found := false

		for _, erfasst123Emp := range activeEmployees {
			if strings.ToLower(erfasst123Emp.Mail) == employeeEmail {
				matchedEmployee = erfasst123Emp
				found = true
				break
			}
		}

		if !found {
			continue
		}

		// Logging
		fmt.Printf("Gefunden: Übereinstimmung für %s %s mit 123erfasst ID: %s\n",
			employee.FirstName, employee.LastName, matchedEmployee.Ident)

		// Flag, um zu prüfen, ob Änderungen vorgenommen wurden
		updated := false

		// 123erfasst ID hinzufügen oder aktualisieren
		if employee.Erfasst123ID != matchedEmployee.Ident {
			employee.Erfasst123ID = matchedEmployee.Ident
			updated = true
			fmt.Printf("Update: 123erfasst ID für %s %s auf %s gesetzt\n",
				employee.FirstName, employee.LastName, matchedEmployee.Ident)
		}

		// Eintrittsdatum aktualisieren, wenn nicht gesetzt
		if employee.HireDate.IsZero() && !matchedEmployee.Employee.HireDateParsed.IsZero() {
			employee.HireDate = matchedEmployee.Employee.HireDateParsed
			updated = true
			fmt.Printf("Update: Eintrittsdatum für %s %s auf %s gesetzt\n",
				employee.FirstName, employee.LastName, matchedEmployee.Employee.HireDateParsed.Format("2006-01-02"))
		}

		// Nur aktualisieren, wenn Änderungen vorgenommen wurden
		if updated {
			employee.UpdatedAt = time.Now()
			err := employeeRepo.Update(employee)
			if err != nil {
				fmt.Printf("Fehler beim Aktualisieren des Mitarbeiters %s %s: %v\n",
					employee.FirstName, employee.LastName, err)
				continue
			}
			updatedCount++
		}
	}

	fmt.Printf("Synchronisierung abgeschlossen: %d Mitarbeiter aktualisiert\n", updatedCount)
	return updatedCount, nil
}

// parseTimeEntry parst Datum und Zeiten korrekt mit Zeitzonenbehandlung
func (s *Erfasst123Service) parseTimeEntry(timeEntry *model.Erfasst123Time) error {
	location := getGermanLocation()

	// Datum parsen
	if timeEntry.Date != "" {
		date, err := time.ParseInLocation("2006-01-02", timeEntry.Date, location)
		if err != nil {
			return fmt.Errorf("fehler beim Parsen des Datums %s: %v", timeEntry.Date, err)
		}
		// Setze auf Mitternacht in der lokalen Zeitzone
		timeEntry.DateParsed = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
	}

	// Startzeit parsen
	if timeEntry.TimeStart != "" {
		startTime, err := s.parseTimeString(timeEntry.TimeStart, timeEntry.DateParsed, location)
		if err != nil {
			return fmt.Errorf("fehler beim Parsen der Startzeit %s: %v", timeEntry.TimeStart, err)
		}
		timeEntry.TimeStartParsed = startTime
	}

	// Endzeit parsen
	if timeEntry.TimeEnd != "" {
		endTime, err := s.parseTimeString(timeEntry.TimeEnd, timeEntry.DateParsed, location)
		if err != nil {
			return fmt.Errorf("fehler beim Parsen der Endzeit %s: %v", timeEntry.TimeEnd, err)
		}
		timeEntry.TimeEndParsed = endTime
	}

	// Dauer berechnen
	if !timeEntry.TimeStartParsed.IsZero() && !timeEntry.TimeEndParsed.IsZero() {
		timeEntry.Duration = timeEntry.TimeEndParsed.Sub(timeEntry.TimeStartParsed).Hours()

		// Wenn Endzeit vor Startzeit liegt, haben wir einen Tagesüberlauf
		if timeEntry.Duration < 0 {
			// Füge 24 Stunden hinzu (Arbeit über Mitternacht)
			timeEntry.TimeEndParsed = timeEntry.TimeEndParsed.AddDate(0, 0, 1)
			timeEntry.Duration = timeEntry.TimeEndParsed.Sub(timeEntry.TimeStartParsed).Hours()
		}
	}

	return nil
}

// parseTimeString parst einen Zeitstring und kombiniert ihn mit dem Datum
func (s *Erfasst123Service) parseTimeString(timeStr string, date time.Time, location *time.Location) (time.Time, error) {
	// Verschiedene Zeitformate ausprobieren
	formats := []string{
		"15:04:05",
		"15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	// Zuerst versuchen, die Zeit direkt zu parsen
	for _, format := range formats {
		if parsedTime, err := time.ParseInLocation(format, timeStr, location); err == nil {
			// Wenn nur Zeit geparst wurde, mit Datum kombinieren
			if parsedTime.Year() == 0 || parsedTime.Year() == 1 {
				return time.Date(
					date.Year(), date.Month(), date.Day(),
					parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second(), 0,
					location,
				), nil
			}
			return parsedTime, nil
		}
	}

	// Manuelles Parsen für HH:MM(:SS) Format
	parts := strings.Split(timeStr, ":")
	if len(parts) >= 2 {
		hour, err := strconv.Atoi(parts[0])
		if err != nil {
			return time.Time{}, fmt.Errorf("ungültige Stunde: %s", parts[0])
		}

		minute, err := strconv.Atoi(parts[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("ungültige Minute: %s", parts[1])
		}

		second := 0
		if len(parts) >= 3 {
			second, _ = strconv.Atoi(parts[2])
		}

		return time.Date(
			date.Year(), date.Month(), date.Day(),
			hour, minute, second, 0,
			location,
		), nil
	}

	return time.Time{}, fmt.Errorf("unbekanntes Zeitformat: %s", timeStr)
}

// GetTimeEntries - Korrigierte Version mit dem richtigen API-Format
func (s *Erfasst123Service) GetTimeEntries(startDate, endDate string) ([]model.Erfasst123Time, error) {
	email, password, err := s.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Basic Auth Token erstellen
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Datum mit Zeitzone formatieren
	startDateTime := fmt.Sprintf("%sT00:00:00Z", startDate)
	endDateTime := fmt.Sprintf("%sT23:59:59Z", endDate)

	fmt.Printf("Fetching time entries from %s to %s\n", startDateTime, endDateTime)

	// GraphQL-Anfrage mit dem korrekten Format
	query := fmt.Sprintf(`{
		"query": "query GetStaffTimesWithActivityAndWage($filter: TimeCollectionFilter) { times(filter: $filter) { nodes { fid person { ident firstname lastname mail } project { id name } date timeStart timeEnd activity { ident name } wageType { ident name } text } totalCount } }",
		"variables": {
			"filter": {
				"date": {
					"_gte": "%s",
					"_lte": "%s"
				}
			}
		}
	}`, startDateTime, endDateTime)

	// HTTP-Anfrage
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://server.123erfasst.de/api/graphql", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("123erfasst API Fehler: %s - Response: %s", res.Status, string(body))
	}

	// Debug: Ausgabe der Response
	fmt.Printf("API Response Status: %s\n", res.Status)
	if len(body) > 200 {
		fmt.Printf("Response (erste 200 Zeichen): %s...\n", string(body[:200]))
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}

	// Antwort parsen
	var response model.Erfasst123TimeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("fehler beim Parsen der Response: %v", err)
	}

	fmt.Printf("Gefunden: %d Zeiteinträge (totalCount: %d)\n",
		len(response.Data.Times.Nodes), response.Data.Times.TotalCount)

	// Zeit- und Datumsparsierung für alle Einträge
	//location := getGermanLocation()
	validEntries := 0

	for i := range response.Data.Times.Nodes {
		timeEntry := &response.Data.Times.Nodes[i]

		// Verwende die parseTimeEntry Funktion
		if err := s.parseTimeEntry(timeEntry); err != nil {
			fmt.Printf("Warnung beim Parsen des Zeiteintrags %d: %v\n", i, err)
			continue
		}

		// Validierung
		if !timeEntry.DateParsed.IsZero() &&
			!timeEntry.TimeStartParsed.IsZero() &&
			!timeEntry.TimeEndParsed.IsZero() {
			validEntries++
		}
	}

	fmt.Printf("Erfolgreich %d von %d Zeiteinträgen mit gültigen Daten geparst\n",
		validEntries, len(response.Data.Times.Nodes))

	// Integration als aktiv markieren
	s.integrationRepo.SetIntegrationStatus("123erfasst", true)
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return response.Data.Times.Nodes, nil
}

// timeEntriesEqual prüft ob zwei Zeiteinträge gleich sind (mit Toleranz für Zeitzonen)
func timeEntriesEqual(e1, e2 model.TimeEntry) bool {
	// Datum vergleichen (nur Jahr, Monat, Tag)
	if !isSameDay(e1.Date, e2.Date) {
		return false
	}

	// Zeiten in UTC konvertieren und vergleichen
	if !e1.StartTime.UTC().Equal(e2.StartTime.UTC()) {
		return false
	}

	if !e1.EndTime.UTC().Equal(e2.EndTime.UTC()) {
		return false
	}

	// Projekt vergleichen
	if e1.ProjectID != e2.ProjectID {
		return false
	}

	return true
}

// isSameDay prüft ob zwei Zeitpunkte am selben Tag sind (unabhängig von der Zeitzone)
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// SyncErfasst123TimeEntries synchronisiert Zeiteinträge von 123erfasst
func (s *Erfasst123Service) SyncErfasst123TimeEntries(startDate, endDate string) (int, error) {
	fmt.Printf("\n=== START SYNC 123ERFASST ZEITEINTRÄGE ===\n")
	fmt.Printf("Zeitraum: %s bis %s\n", startDate, endDate)

	// Zeiteinträge von 123erfasst abrufen
	timeEntries, err := s.GetTimeEntries(startDate, endDate)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Abgerufene Zeiteinträge: %d\n", len(timeEntries))

	// Parse dates für Filterung
	startDateParsed, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return 0, fmt.Errorf("ungültiges Startdatum: %v", err)
	}

	endDateParsed, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return 0, fmt.Errorf("ungültiges Enddatum: %v", err)
	}

	// Repository für Mitarbeiter initialisieren
	employeeRepo := repository.NewEmployeeRepository()

	// Alle Mitarbeiter abrufen
	allEmployees, err := employeeRepo.FindAll()
	if err != nil {
		return 0, fmt.Errorf("Fehler beim Abrufen der Mitarbeiter: %v", err)
	}

	fmt.Printf("\nGeladene Mitarbeiter aus der Datenbank: %d\n", len(allEmployees))

	// Debug: Mitarbeiter mit 123erfasst IDs ausgeben
	erfasst123Count := 0
	for _, emp := range allEmployees {
		if emp.Erfasst123ID != "" {
			erfasst123Count++
			fmt.Printf("  Mitarbeiter mit 123erfasst ID: %s %s (ID: %s, Email: %s)\n",
				emp.FirstName, emp.LastName, emp.Erfasst123ID, emp.Email)
		}
	}
	fmt.Printf("Mitarbeiter mit 123erfasst ID: %d\n", erfasst123Count)

	// Mitarbeiter-Maps erstellen
	employeeMap := make(map[string]*model.Employee)
	employeeByErfasst123ID := make(map[string]*model.Employee)

	for _, emp := range allEmployees {
		key := strings.ToLower(strings.TrimSpace(emp.Email))
		employeeMap[key] = emp

		if emp.Erfasst123ID != "" {
			employeeByErfasst123ID[emp.Erfasst123ID] = emp
		}
	}

	// Tracking für Updates
	updatedEmployees := make(map[string]*model.Employee)
	notFoundEmployees := make(map[string]bool)

	fmt.Printf("\n=== VERARBEITUNG DER ZEITEINTRÄGE ===\n")

	// Zeiteinträge verarbeiten
	for idx, timeEntry := range timeEntries {
		// Debug für ersten Eintrag
		if idx == 0 {
			fmt.Printf("\nVerarbeite ersten Zeiteintrag im Detail:\n")
			fmt.Printf("  Person: %s %s\n", timeEntry.Person.Firstname, timeEntry.Person.Lastname)
			fmt.Printf("  123erfasst ID: %s\n", timeEntry.Person.Ident)
			fmt.Printf("  Email: %s\n", timeEntry.Person.Mail)
			fmt.Printf("  Datum: %s\n", timeEntry.Date)
			fmt.Printf("  Zeit: %s - %s\n", timeEntry.TimeStart, timeEntry.TimeEnd)
			fmt.Printf("  DateParsed: %v\n", timeEntry.DateParsed)
			fmt.Printf("  TimeStartParsed: %v\n", timeEntry.TimeStartParsed)
			fmt.Printf("  TimeEndParsed: %v\n", timeEntry.TimeEndParsed)
		}

		// Validierung
		if timeEntry.DateParsed.IsZero() || timeEntry.TimeStartParsed.IsZero() || timeEntry.TimeEndParsed.IsZero() {
			fmt.Printf("WARNUNG: Überspringe ungültigen Zeiteintrag für %s %s (Datum: %v, Start: %v, Ende: %v)\n",
				timeEntry.Person.Firstname, timeEntry.Person.Lastname,
				timeEntry.DateParsed.IsZero(), timeEntry.TimeStartParsed.IsZero(), timeEntry.TimeEndParsed.IsZero())
			continue
		}

		// Mitarbeiter suchen - erst über 123erfasst ID, dann über Email
		var employee *model.Employee

		// Methode 1: Über 123erfasst ID
		if timeEntry.Person.Ident != "" {
			employee = employeeByErfasst123ID[timeEntry.Person.Ident]
			if employee != nil && idx < 10 {
				fmt.Printf("✓ Gefunden via 123erfasst ID: %s %s\n", employee.FirstName, employee.LastName)
			}
		}

		// Methode 2: Über Email
		if employee == nil && timeEntry.Person.Mail != "" {
			emailKey := strings.ToLower(strings.TrimSpace(timeEntry.Person.Mail))
			employee = employeeMap[emailKey]
			if employee != nil && idx < 10 {
				fmt.Printf("✓ Gefunden via Email: %s %s\n", employee.FirstName, employee.LastName)
			}
		}

		// Nicht gefunden
		if employee == nil {
			personKey := fmt.Sprintf("%s_%s", timeEntry.Person.Firstname, timeEntry.Person.Lastname)
			if !notFoundEmployees[personKey] {
				notFoundEmployees[personKey] = true
				fmt.Printf("✗ NICHT GEFUNDEN: %s %s (ID: %s, Email: %s)\n",
					timeEntry.Person.Firstname, timeEntry.Person.Lastname,
					timeEntry.Person.Ident, timeEntry.Person.Mail)
			}
			continue
		}

		// Neuen Zeiteintrag erstellen
		newTimeEntry := model.TimeEntry{
			ID:          primitive.NewObjectID(),
			Date:        timeEntry.DateParsed,
			StartTime:   timeEntry.TimeStartParsed,
			EndTime:     timeEntry.TimeEndParsed,
			Duration:    timeEntry.Duration,
			ProjectID:   timeEntry.Project.ID,
			ProjectName: timeEntry.Project.Name,
			Activity:    timeEntry.Activity.Name,
			Source:      "123erfasst",
		}

		if timeEntry.WageType != nil {
			newTimeEntry.WageType = timeEntry.WageType.Name
		}

		// Zeiteintrag hinzufügen
		if updatedEmployees[employee.ID.Hex()] == nil {
			// Erstmalig: Employee kopieren
			empCopy := *employee
			updatedEmployees[employee.ID.Hex()] = &empCopy
		}
		updatedEmployees[employee.ID.Hex()].TimeEntries = append(
			updatedEmployees[employee.ID.Hex()].TimeEntries,
			newTimeEntry,
		)
	}

	fmt.Printf("\n=== BEREINIGUNG UND SPEICHERUNG ===\n")
	fmt.Printf("Mitarbeiter mit neuen Zeiteinträgen: %d\n", len(updatedEmployees))

	// Updates speichern
	updateCount := 0
	for _, employee := range updatedEmployees {
		// WICHTIG: Mitarbeiter aus DB neu laden für sauberen Stand
		dbEmployee, err := employeeRepo.FindByID(employee.ID.Hex())
		if err != nil {
			fmt.Printf("✗ Fehler beim Abrufen von %s %s: %v\n",
				employee.FirstName, employee.LastName, err)
			continue
		}

		// Schritt 1: Alle NICHT-123erfasst Einträge behalten
		var keptEntries []model.TimeEntry
		for _, entry := range dbEmployee.TimeEntries {
			if entry.Source != "123erfasst" {
				keptEntries = append(keptEntries, entry)
			}
		}

		// Schritt 2: Alle 123erfasst-Einträge außerhalb des Sync-Zeitraums behalten
		for _, entry := range dbEmployee.TimeEntries {
			if entry.Source == "123erfasst" &&
				(entry.Date.Before(startDateParsed) || entry.Date.After(endDateParsed)) {
				keptEntries = append(keptEntries, entry)
			}
		}

		// Schritt 3: Neue Einträge aus der aktuellen Synchronisation sammeln
		var newEntries []model.TimeEntry
		for _, entry := range employee.TimeEntries {
			// Nur Einträge im Sync-Zeitraum
			if !entry.Date.Before(startDateParsed) && !entry.Date.After(endDateParsed) {
				newEntries = append(newEntries, entry)
			}
		}

		// Debug-Ausgabe
		fmt.Printf("\nMitarbeiter %s %s:\n", dbEmployee.FirstName, dbEmployee.LastName)
		fmt.Printf("  Einträge in DB gesamt: %d\n", len(dbEmployee.TimeEntries))
		fmt.Printf("  Davon 123erfasst im Sync-Zeitraum: %d\n",
			len(dbEmployee.TimeEntries)-len(keptEntries))
		fmt.Printf("  Neue 123erfasst-Einträge: %d\n", len(newEntries))

		// Schritt 4: Kombiniere alte (gefilterte) und neue Einträge
		dbEmployee.TimeEntries = append(keptEntries, newEntries...)

		// Schritt 5: Duplikate entfernen (mit korrigierter Funktion)
		dbEmployee.TimeEntries = s.removeDuplicateTimeEntries(dbEmployee.TimeEntries)

		fmt.Printf("  Gesamt nach Deduplizierung: %d\n", len(dbEmployee.TimeEntries))

		// Berechne Gesamtstunden für Debug
		var totalHours float64
		for _, entry := range dbEmployee.TimeEntries {
			if entry.Date.Format("2006-01-02") == "2025-06-05" {
				totalHours += entry.Duration
			}
		}
		if totalHours > 0 {
			fmt.Printf("  Stunden am 05.06.2025: %.2f\n", totalHours)
		}

		dbEmployee.UpdatedAt = time.Now()

		// Mitarbeiter aktualisieren
		if err := employeeRepo.Update(dbEmployee); err != nil {
			fmt.Printf("✗ Fehler beim Aktualisieren von %s %s: %v\n",
				dbEmployee.FirstName, dbEmployee.LastName, err)
			continue
		}

		updateCount++
		fmt.Printf("✓ Erfolgreich aktualisiert: %s %s\n", dbEmployee.FirstName, dbEmployee.LastName)
	}

	fmt.Printf("\n=== ZUSAMMENFASSUNG ===\n")
	fmt.Printf("Zeiteinträge von 123erfasst: %d\n", len(timeEntries))
	fmt.Printf("Nicht zugeordnete Personen: %d\n", len(notFoundEmployees))
	fmt.Printf("Aktualisierte Mitarbeiter: %d\n", updateCount)
	fmt.Printf("=== SYNC ENDE ===\n\n")

	// Letzte Synchronisation aktualisieren
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return updateCount, nil
}

func (s *Erfasst123Service) cleanOldTimeEntries(entries []model.TimeEntry, startDate, endDate time.Time) []model.TimeEntry {
	var cleanedEntries []model.TimeEntry
	removedCount := 0
	keptCount := 0

	fmt.Printf("\n  Bereinige alte Einträge (Zeitraum: %s bis %s):\n",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	debugCount := 0
	for _, entry := range entries {
		// Debug für erste 3 Einträge
		if debugCount < 3 {
			fmt.Printf("    Eintrag %d: Datum=%s, Source=%s, Projekt=%s\n",
				debugCount+1, entry.Date.Format("2006-01-02"), entry.Source, entry.ProjectName)
		}

		// Behalte Einträge die:
		// 1. Nicht von 123erfasst sind ODER
		// 2. Außerhalb des Synchronisationszeitraums liegen
		shouldKeep := false
		reason := ""

		if entry.Source != "123erfasst" {
			shouldKeep = true
			reason = "nicht von 123erfasst"
		} else if entry.Date.Before(startDate) {
			shouldKeep = true
			reason = fmt.Sprintf("vor Startdatum (%s < %s)",
				entry.Date.Format("2006-01-02"), startDate.Format("2006-01-02"))
		} else if entry.Date.After(endDate) {
			shouldKeep = true
			reason = fmt.Sprintf("nach Enddatum (%s > %s)",
				entry.Date.Format("2006-01-02"), endDate.Format("2006-01-02"))
		} else {
			// Dieser Eintrag wird entfernt
			reason = "im Sync-Zeitraum und von 123erfasst"
			removedCount++
		}

		if debugCount < 3 {
			if shouldKeep {
				fmt.Printf("      → Behalten: %s\n", reason)
			} else {
				fmt.Printf("      → Entfernen: %s\n", reason)
			}
			debugCount++
		}

		if shouldKeep {
			cleanedEntries = append(cleanedEntries, entry)
			keptCount++
		}
	}

	fmt.Printf("  Zusammenfassung: %d behalten, %d entfernt\n", keptCount, removedCount)

	return cleanedEntries
}

func (s *Erfasst123Service) removeDuplicateTimeEntries(entries []model.TimeEntry) []model.TimeEntry {
	seen := make(map[string]bool)
	var uniqueEntries []model.TimeEntry

	for _, entry := range entries {
		// Eindeutigen Schlüssel für jeden Eintrag erstellen - INKLUSIVE Activity!
		key := fmt.Sprintf("%s_%s_%s_%s_%s_%s",
			entry.Date.Format("2006-01-02"),
			entry.StartTime.UTC().Format("15:04:05"),
			entry.EndTime.UTC().Format("15:04:05"),
			entry.ProjectID,
			entry.Activity, // NEU: Activity ist jetzt Teil des Schlüssels!
			entry.Source)

		if !seen[key] {
			seen[key] = true
			uniqueEntries = append(uniqueEntries, entry)
		}
	}

	return uniqueEntries
}

// GetProjects ruft Projekte von 123erfasst ab
func (s *Erfasst123Service) GetProjects(startDate, endDate string) ([]model.Erfasst123Planning, error) {
	email, password, err := s.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Basic Auth Token erstellen
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// GraphQL-Anfrage für Projekte
	query := fmt.Sprintf(`{
		"query": "query GetPlannings($dateFrom: Date!, $dateTo: Date!) { plannings(dateFrom: $dateFrom, dateTo: $dateTo) { nodes { project { id name } persons { ident firstname lastname mail } dateStart dateEnd } totalCount } }",
		"variables": {
			"dateFrom": "%s",
			"dateTo": "%s"
		}
	}`, startDate, endDate)

	// HTTP-Anfrage
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", "https://server.123erfasst.de/api/graphql", bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("123erfasst API Fehler: %s - %s", res.Status, string(body))
	}

	// Antwort parsen
	var response model.Erfasst123PlanningResponse
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	location := getGermanLocation()

	// Datum konvertieren
	for i, planning := range response.Data.Plannings.Nodes {
		// Start Datum
		if planning.DateStart != "" {
			startDate, err := time.ParseInLocation("2006-01-02", planning.DateStart, location)
			if err == nil {
				response.Data.Plannings.Nodes[i].DateStartParsed = startDate
			}
		}

		// End Datum
		if planning.DateEnd != "" {
			endDate, err := time.ParseInLocation("2006-01-02", planning.DateEnd, location)
			if err == nil {
				response.Data.Plannings.Nodes[i].DateEndParsed = endDate
			}
		}
	}

	// Integration als aktiv markieren
	s.integrationRepo.SetIntegrationStatus("123erfasst", true)
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return response.Data.Plannings.Nodes, nil
}

// SyncErfasst123Projects synchronisiert 123erfasst Projekte mit PeopleFlow Mitarbeitern
func (s *Erfasst123Service) SyncErfasst123Projects(startDate, endDate string) (int, error) {
	// Projekte von 123erfasst abrufen
	projects, err := s.GetProjects(startDate, endDate)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Erhalten: %d Projekte von 123erfasst\n", len(projects))

	// TODO: Implementierung der Projektsynchronisation
	// Dies könnte beinhalten:
	// - Erstellen von Projekten in PeopleFlow
	// - Zuordnung von Mitarbeitern zu Projekten
	// - Aktualisierung von Projektinformationen

	return len(projects), nil
}

// TestEmployeeMapping testet die Zuordnung zwischen 123erfasst und PeopleFlow Mitarbeitern
func (s *Erfasst123Service) TestEmployeeMapping() error {
	fmt.Println("=== TEST: Mitarbeiter-Zuordnung ===")

	// 1. Hole 123erfasst Mitarbeiter
	fmt.Println("\n1. Hole aktive Mitarbeiter von 123erfasst...")
	erfasst123Employees, err := s.GetEmployees()
	if err != nil {
		return fmt.Errorf("Fehler beim Abrufen der 123erfasst Mitarbeiter: %v", err)
	}

	activeCount := 0
	for _, emp := range erfasst123Employees {
		if emp.Employee.IsActive {
			activeCount++
			fmt.Printf("  - %s %s (ID: %s, Email: %s)\n",
				emp.Firstname, emp.Lastname, emp.Ident, emp.Mail)
		}
	}
	fmt.Printf("Gefunden: %d aktive Mitarbeiter\n", activeCount)

	// 2. Hole PeopleFlow Mitarbeiter
	fmt.Println("\n2. Hole Mitarbeiter aus PeopleFlow Datenbank...")
	employeeRepo := repository.NewEmployeeRepository()
	peopleFlowEmployees, err := employeeRepo.FindAll()
	if err != nil {
		return fmt.Errorf("Fehler beim Abrufen der PeopleFlow Mitarbeiter: %v", err)
	}

	fmt.Printf("Gefunden: %d Mitarbeiter\n", len(peopleFlowEmployees))
	for i, emp := range peopleFlowEmployees {
		if i < 10 { // Erste 10 anzeigen
			fmt.Printf("  - %s %s (Email: %s, 123erfasst ID: %s)\n",
				emp.FirstName, emp.LastName, emp.Email, emp.Erfasst123ID)
		}
	}

	// 3. Teste Zuordnung
	fmt.Println("\n3. Teste Zuordnung...")
	matchedCount := 0
	unmatchedErfasst := []model.Erfasst123Person{}

	for _, erfasst123Emp := range erfasst123Employees {
		if !erfasst123Emp.Employee.IsActive {
			continue
		}

		found := false

		// Suche nach Email
		for _, pfEmp := range peopleFlowEmployees {
			if strings.EqualFold(strings.TrimSpace(pfEmp.Email), strings.TrimSpace(erfasst123Emp.Mail)) {
				matchedCount++
				found = true
				fmt.Printf("  ✓ Zuordnung via Email: %s %s ↔ %s %s\n",
					erfasst123Emp.Firstname, erfasst123Emp.Lastname,
					pfEmp.FirstName, pfEmp.LastName)
				break
			}
		}

		// Wenn nicht via Email gefunden, suche nach Name
		if !found {
			for _, pfEmp := range peopleFlowEmployees {
				if strings.EqualFold(strings.TrimSpace(pfEmp.FirstName), strings.TrimSpace(erfasst123Emp.Firstname)) &&
					strings.EqualFold(strings.TrimSpace(pfEmp.LastName), strings.TrimSpace(erfasst123Emp.Lastname)) {
					matchedCount++
					found = true
					fmt.Printf("  ✓ Zuordnung via Name: %s %s ↔ %s %s\n",
						erfasst123Emp.Firstname, erfasst123Emp.Lastname,
						pfEmp.FirstName, pfEmp.LastName)
					break
				}
			}
		}

		if !found {
			unmatchedErfasst = append(unmatchedErfasst, erfasst123Emp)
		}
	}

	fmt.Printf("\n=== ERGEBNIS ===\n")
	fmt.Printf("Erfolgreich zugeordnet: %d von %d\n", matchedCount, activeCount)

	if len(unmatchedErfasst) > 0 {
		fmt.Printf("\nNicht zugeordnete 123erfasst Mitarbeiter:\n")
		for _, emp := range unmatchedErfasst {
			fmt.Printf("  - %s %s (Email: %s, ID: %s)\n",
				emp.Firstname, emp.Lastname, emp.Mail, emp.Ident)
		}
	}

	// 4. Teste einen einzelnen Zeiteintrag
	fmt.Println("\n4. Teste Zeiteintrag-Abruf...")
	today := time.Now()
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	timeEntries, err := s.GetTimeEntries(
		startOfMonth.Format("2006-01-02"),
		endOfMonth.Format("2006-01-02"),
	)

	if err != nil {
		fmt.Printf("Fehler beim Abrufen der Zeiteinträge: %v\n", err)
	} else {
		fmt.Printf("Zeiteinträge für %s: %d\n",
			startOfMonth.Format("January 2006"), len(timeEntries))

		if len(timeEntries) > 0 {
			fmt.Println("Erste 5 Zeiteinträge:")
			for i, entry := range timeEntries {
				if i >= 5 {
					break
				}
				fmt.Printf("  - %s: %s %s, Projekt: %s\n",
					entry.Date,
					entry.Person.Firstname, entry.Person.Lastname,
					entry.Project.Name)
			}
		}
	}

	return nil
}

// CleanupDuplicateTimeEntries bereinigt doppelte Zeiteinträge für alle Mitarbeiter
func (s *Erfasst123Service) CleanupDuplicateTimeEntries() (int, error) {
	fmt.Printf("\n=== START BEREINIGUNG DUPLIKATE ===\n")

	employeeRepo := repository.NewEmployeeRepository()
	employees, err := employeeRepo.FindAll()
	if err != nil {
		return 0, fmt.Errorf("Fehler beim Abrufen der Mitarbeiter: %v", err)
	}

	cleanedCount := 0
	totalDuplicatesRemoved := 0

	for _, employee := range employees {
		originalCount := len(employee.TimeEntries)

		// Nur Mitarbeiter mit Zeiteinträgen bereinigen
		if originalCount == 0 {
			continue
		}

		// Duplikate entfernen
		employee.TimeEntries = s.removeDuplicateTimeEntries(employee.TimeEntries)
		newCount := len(employee.TimeEntries)

		// Nur aktualisieren, wenn sich etwas geändert hat
		if originalCount != newCount {
			duplicatesRemoved := originalCount - newCount
			totalDuplicatesRemoved += duplicatesRemoved

			// Debug-Ausgabe für erste 5 Mitarbeiter
			if cleanedCount < 5 {
				fmt.Printf("Bereinige %s %s: %d -> %d Einträge (%d Duplikate entfernt)\n",
					employee.FirstName, employee.LastName,
					originalCount, newCount, duplicatesRemoved)
			}

			// Mitarbeiter aktualisieren
			employee.UpdatedAt = time.Now()
			if err := employeeRepo.Update(employee); err != nil {
				fmt.Printf("Fehler beim Aktualisieren von %s %s: %v\n",
					employee.FirstName, employee.LastName, err)
				continue
			}

			cleanedCount++
		}
	}

	fmt.Printf("\n=== BEREINIGUNG ABGESCHLOSSEN ===\n")
	fmt.Printf("Mitarbeiter bereinigt: %d\n", cleanedCount)
	fmt.Printf("Duplikate entfernt: %d\n", totalDuplicatesRemoved)
	fmt.Printf("=== ENDE ===\n\n")

	return cleanedCount, nil
}
