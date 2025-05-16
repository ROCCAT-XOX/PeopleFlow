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
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	// Datum konvertieren
	for i, person := range response.Data.Persons.Nodes {
		// HireDate
		if person.Employee.HireDate != "" {
			hireDate, err := time.Parse("2006-01-02", person.Employee.HireDate)
			if err != nil {
				fmt.Printf("Warnung: Fehler beim Parsen des Eintrittsdatums für %s %s: %v\n",
					person.Firstname, person.Lastname, err)
			} else {
				response.Data.Persons.Nodes[i].Employee.HireDateParsed = hireDate
			}
		}

		// ExitDate (kann null sein)
		if person.Employee.ExitDate != nil && *person.Employee.ExitDate != "" {
			exitDate, err := time.Parse("2006-01-02", *person.Employee.ExitDate)
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

		// Wenn Änderungen vorgenommen wurden, Mitarbeiter aktualisieren
		if updated {
			employee.UpdatedAt = time.Now()
			err := employeeRepo.Update(employee)
			if err != nil {
				fmt.Printf("Fehler: Konnte %s %s nicht aktualisieren: %v\n",
					employee.FirstName, employee.LastName, err)
				return updatedCount, err
			}
			updatedCount++
			fmt.Printf("Erfolg: %s %s wurde aktualisiert\n", employee.FirstName, employee.LastName)
		}
	}

	return updatedCount, nil
}

// GetProjectPlannings retrieves project planning data from 123erfasst
func (s *Erfasst123Service) GetProjectPlannings(startDate, endDate string) ([]model.Erfasst123Planning, error) {
	email, password, err := s.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Basic Auth Token
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// GraphQL query for project planning data
	query := fmt.Sprintf(`{
		"query": "query GetPlanningsByDateRange($filter: PlanningFilter) { plannings(filter: $filter) { nodes { project { id name } persons { ident firstname lastname } dateStart dateEnd } totalCount } }",
		"variables": {
			"filter": {
				"dateFrom": {
					"_gte": "%sT00:00:00Z",
					"_lte": "%sT23:59:59Z"
				}
			}
		}
	}`, startDate, endDate)

	// HTTP request
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
		return nil, fmt.Errorf("123erfasst API error: %s", res.Status)
	}

	// Parse response
	var response model.Erfasst123PlanningResponse
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	// Parse dates and mark the integration as active
	for i, planning := range response.Data.Plannings.Nodes {
		// Parse start date
		if planning.DateStart != "" {
			startDate, err := time.Parse("2006-01-02", planning.DateStart)
			if err == nil {
				response.Data.Plannings.Nodes[i].DateStartParsed = startDate
			} else {
				fmt.Printf("Warning: Error parsing start date %s: %v\n", planning.DateStart, err)
			}
		}

		// Parse end date
		if planning.DateEnd != "" {
			endDate, err := time.Parse("2006-01-02", planning.DateEnd)
			if err == nil {
				response.Data.Plannings.Nodes[i].DateEndParsed = endDate
			} else {
				fmt.Printf("Warning: Error parsing end date %s: %v\n", planning.DateEnd, err)
			}
		}
	}

	// Mark integration as active and update last sync time
	s.integrationRepo.SetIntegrationStatus("123erfasst", true)
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return response.Data.Plannings.Nodes, nil
}

// SyncErfasst123Projects synchronizes 123erfasst project planning data with PeopleFlow employees
func (s *Erfasst123Service) SyncErfasst123Projects(startDate, endDate string) (int, error) {
	// Get project planning data
	plannings, err := s.GetProjectPlannings(startDate, endDate)
	if err != nil {
		return 0, err
	}

	// Repository for employees
	employeeRepo := repository.NewEmployeeRepository()

	// Counter for updated employees
	updatedCount := 0

	// Process each planning
	for _, planning := range plannings {
		// Skip if dates not parsed correctly
		if planning.DateStartParsed.IsZero() || planning.DateEndParsed.IsZero() {
			continue
		}

		// Process each person in the planning
		for _, person := range planning.Persons {
			// Find the employee by 123erfasst ID
			employee, err := employeeRepo.FindByErfasst123ID(person.Ident)
			if err != nil {
				// Employee not found, try by email or name
				employees, err := employeeRepo.FindAll()
				if err != nil {
					continue
				}

				// Match by email or name
				for _, emp := range employees {
					if strings.EqualFold(emp.Email, person.Mail) ||
						(strings.EqualFold(emp.FirstName, person.Firstname) &&
							strings.EqualFold(emp.LastName, person.Lastname)) {
						employee = emp
						break
					}
				}

				// If still not found, skip this person
				if employee == nil {
					continue
				}
			}

			// Check if this project assignment already exists
			projectExists := false
			for _, assignment := range employee.ProjectAssignments {
				// Check if the same project with overlapping dates exists
				if assignment.ProjectID == planning.Project.ID &&
					assignment.Source == "123erfasst" &&
					assignment.StartDate.Equal(planning.DateStartParsed) &&
					assignment.EndDate.Equal(planning.DateEndParsed) {
					projectExists = true
					break
				}
			}

			// If project doesn't exist, add it
			if !projectExists {
				// Create a new project assignment
				newAssignment := model.ProjectAssignment{
					ID:          primitive.NewObjectID(),
					ProjectID:   planning.Project.ID,
					ProjectName: planning.Project.Name,
					StartDate:   planning.DateStartParsed,
					EndDate:     planning.DateEndParsed,
					Source:      "123erfasst",
				}

				// Add to employee's project assignments
				employee.ProjectAssignments = append(employee.ProjectAssignments, newAssignment)
				employee.UpdatedAt = time.Now()

				// Update employee in database
				err = employeeRepo.Update(employee)
				if err != nil {
					fmt.Printf("Error updating employee %s %s: %v\n",
						employee.FirstName, employee.LastName, err)
					continue
				}

				updatedCount++
				fmt.Printf("Updated employee %s %s with project assignment: %s\n",
					employee.FirstName, employee.LastName, planning.Project.Name)
			}
		}
	}

	return updatedCount, nil
}

// GetTimeEntries retrieves time entry data from 123erfasst
func (s *Erfasst123Service) GetTimeEntries(startDate, endDate string) ([]model.Erfasst123Time, error) {
	email, password, err := s.GetCredentials()
	if err != nil {
		return nil, err
	}

	// Basic Auth Token
	auth := fmt.Sprintf("%s:%s", email, password)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	// Format the dates properly with timezone information
	startDateTime := fmt.Sprintf("%sT00:00:00Z", startDate)
	endDateTime := fmt.Sprintf("%sT23:59:59Z", endDate)

	fmt.Printf("Formatted date range: %s to %s\n", startDateTime, endDateTime)

	// GraphQL query for time entry data
	query := fmt.Sprintf(`{
        "query": "query GetStaffTimesWithActivityAndWage($filter: TimeCollectionFilter) { times(filter: $filter) { nodes { fid person { ident firstname lastname mail } project { id name } date timeStart timeEnd activity { ident name } wageType { ident name } } totalCount } }",
        "variables": {
            "filter": {
                "date": {
                    "_gte": "%s",
                    "_lte": "%s"
                }
            }
        }
    }`, startDateTime, endDateTime)

	// Debug: Print the query
	fmt.Printf("Query: %s\n", query)

	// HTTP request
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
		return nil, fmt.Errorf("123erfasst API error: %s", res.Status)
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Debug: Log a sample of the response
	if len(body) > 100 {
		fmt.Printf("Response sample (first 100 chars): %s...\n", string(body[:100]))
	} else {
		fmt.Printf("Response: %s\n", string(body))
	}

	// Parse response using a new reader
	var response model.Erfasst123TimeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Parse dates and calculate durations
	for i, timeEntry := range response.Data.Times.Nodes {
		// Debug: Log the raw date fields for this entry
		fmt.Printf("Entry %d raw data - date: %s, timeStart: %s, timeEnd: %s\n",
			i, timeEntry.Date, timeEntry.TimeStart, timeEntry.TimeEnd)

		// Parse date with multiple formats
		if timeEntry.Date != "" {
			var parseErr error
			var dateParsed time.Time

			// Try different date formats
			formats := []string{
				"2006-01-02",  // ISO format
				"02.01.2006",  // German format
				"2006/01/02",  // Slash format
				"Jan 2, 2006", // Text format
				"2 Jan 2006",  // Alternative text format
				time.RFC3339,  // RFC3339
			}

			for _, format := range formats {
				dateParsed, parseErr = time.Parse(format, timeEntry.Date)
				if parseErr == nil {
					response.Data.Times.Nodes[i].DateParsed = dateParsed
					break
				}
			}

			if parseErr != nil {
				fmt.Printf("Warning: Could not parse date %s with any format\n", timeEntry.Date)
			}
		} else {
			fmt.Printf("Warning: Empty date field for time entry %d\n", i)
		}

		// Parse start time with multiple formats
		if timeEntry.TimeStart != "" {
			var parseErr error
			var startTime time.Time

			// Try different time formats
			formats := []string{
				time.RFC3339,          // 2006-01-02T15:04:05Z07:00
				"2006-01-02T15:04:05", // Without timezone
				"2006-01-02 15:04:05", // Space instead of T
				"15:04:05",            // Time only
			}

			for _, format := range formats {
				startTime, parseErr = time.Parse(format, timeEntry.TimeStart)
				if parseErr == nil {
					response.Data.Times.Nodes[i].TimeStartParsed = startTime
					break
				}
			}

			if parseErr != nil {
				// Special handling for time-only formats
				if len(timeEntry.TimeStart) <= 8 && strings.Contains(timeEntry.TimeStart, ":") {
					// If it's just time (like "07:30:00"), combine with the date
					if !response.Data.Times.Nodes[i].DateParsed.IsZero() {
						timeComponents := strings.Split(timeEntry.TimeStart, ":")
						if len(timeComponents) >= 2 {
							hour, _ := strconv.Atoi(timeComponents[0])
							minute, _ := strconv.Atoi(timeComponents[1])

							// Create a new time using the date and the time components
							date := response.Data.Times.Nodes[i].DateParsed
							combinedTime := time.Date(
								date.Year(), date.Month(), date.Day(),
								hour, minute, 0, 0, time.UTC)

							response.Data.Times.Nodes[i].TimeStartParsed = combinedTime
							fmt.Printf("Combined date %s with time %s: %s\n",
								date.Format("2006-01-02"),
								timeEntry.TimeStart,
								combinedTime.Format(time.RFC3339))
						}
					}
				} else {
					fmt.Printf("Warning: Could not parse start time %s with any format\n", timeEntry.TimeStart)
				}
			}
		} else {
			fmt.Printf("Warning: Empty timeStart field for time entry %d\n", i)
		}

		// Parse end time with similar approach
		if timeEntry.TimeEnd != "" {
			var parseErr error
			var endTime time.Time

			// Try different time formats
			formats := []string{
				time.RFC3339,          // 2006-01-02T15:04:05Z07:00
				"2006-01-02T15:04:05", // Without timezone
				"2006-01-02 15:04:05", // Space instead of T
				"15:04:05",            // Time only
			}

			for _, format := range formats {
				endTime, parseErr = time.Parse(format, timeEntry.TimeEnd)
				if parseErr == nil {
					response.Data.Times.Nodes[i].TimeEndParsed = endTime
					break
				}
			}

			if parseErr != nil {
				// Special handling for time-only formats
				if len(timeEntry.TimeEnd) <= 8 && strings.Contains(timeEntry.TimeEnd, ":") {
					// If it's just time (like "12:00:00"), combine with the date
					if !response.Data.Times.Nodes[i].DateParsed.IsZero() {
						timeComponents := strings.Split(timeEntry.TimeEnd, ":")
						if len(timeComponents) >= 2 {
							hour, _ := strconv.Atoi(timeComponents[0])
							minute, _ := strconv.Atoi(timeComponents[1])

							// Create a new time using the date and the time components
							date := response.Data.Times.Nodes[i].DateParsed
							combinedTime := time.Date(
								date.Year(), date.Month(), date.Day(),
								hour, minute, 0, 0, time.UTC)

							response.Data.Times.Nodes[i].TimeEndParsed = combinedTime
							fmt.Printf("Combined date %s with time %s: %s\n",
								date.Format("2006-01-02"),
								timeEntry.TimeEnd,
								combinedTime.Format(time.RFC3339))
						}
					}
				} else {
					fmt.Printf("Warning: Could not parse end time %s with any format\n", timeEntry.TimeEnd)
				}
			}
		} else {
			fmt.Printf("Warning: Empty timeEnd field for time entry %d\n", i)
		}

		// Calculate duration if both start and end times are valid
		if !response.Data.Times.Nodes[i].TimeStartParsed.IsZero() && !response.Data.Times.Nodes[i].TimeEndParsed.IsZero() {
			duration := response.Data.Times.Nodes[i].TimeEndParsed.Sub(response.Data.Times.Nodes[i].TimeStartParsed).Hours()
			response.Data.Times.Nodes[i].Duration = duration
			fmt.Printf("Calculated duration for entry %d: %.2f hours\n", i, duration)
		} else {
			fmt.Printf("Skipping duration calculation for entry %d due to invalid times\n", i)
		}
	}

	// Log how many entries have valid dates
	validEntries := 0
	for _, entry := range response.Data.Times.Nodes {
		if !entry.DateParsed.IsZero() && !entry.TimeStartParsed.IsZero() && !entry.TimeEndParsed.IsZero() {
			validEntries++
		}
	}
	fmt.Printf("Found %d entries with valid dates out of %d total entries\n",
		validEntries, len(response.Data.Times.Nodes))

	// Mark integration as active
	s.integrationRepo.SetIntegrationStatus("123erfasst", true)
	s.integrationRepo.SetLastSync("123erfasst", time.Now())

	return response.Data.Times.Nodes, nil
}

// SyncErfasst123TimeEntries synchronizes 123erfasst time entries with PeopleFlow employees
func (s *Erfasst123Service) SyncErfasst123TimeEntries(startDate, endDate string) (int, error) {
	// Get time entries from 123erfasst
	fmt.Printf("Fetching time entries from 123erfasst for period %s to %s\n", startDate, endDate)

	timeEntries, err := s.GetTimeEntries(startDate, endDate)
	if err != nil {
		fmt.Printf("Error fetching time entries: %v\n", err)
		return 0, err
	}

	fmt.Printf("Received %d time entries from 123erfasst\n", len(timeEntries))

	// Repository for employees
	employeeRepo := repository.NewEmployeeRepository()

	// Counter for updated employees
	updatedCount := 0

	// Process each time entry
	employeesChecked := make(map[string]bool)
	employeesFound := make(map[string]bool)

	for _, timeEntry := range timeEntries {
		// Skip if dates not parsed correctly
		if timeEntry.DateParsed.IsZero() || timeEntry.TimeStartParsed.IsZero() || timeEntry.TimeEndParsed.IsZero() {
			fmt.Printf("Skipping time entry with invalid dates for person %s %s\n",
				timeEntry.Person.Firstname, timeEntry.Person.Lastname)
			continue
		}

		personIdent := timeEntry.Person.Ident
		personName := fmt.Sprintf("%s %s", timeEntry.Person.Firstname, timeEntry.Person.Lastname)

		// Only log the lookup attempt once per employee
		if !employeesChecked[personIdent] {
			employeesChecked[personIdent] = true
			fmt.Printf("Looking up employee with 123erfasst ID: %s (%s)\n", personIdent, personName)
		}

		// Find the employee by 123erfasst ID
		employee, err := employeeRepo.FindByErfasst123ID(personIdent)
		if err != nil {
			if !employeesChecked[personIdent] {
				fmt.Printf("Could not find employee by Erfasst123ID %s: %v\n", personIdent, err)
			}

			// Employee not found by ID, try by email and name
			if !employeesFound[personIdent] {
				employees, err := employeeRepo.FindAll()
				if err != nil {
					fmt.Printf("Error fetching all employees: %v\n", err)
					continue
				}

				// Match by email or name
				employeeFound := false
				for _, emp := range employees {
					if strings.EqualFold(emp.Email, timeEntry.Person.Mail) ||
						(strings.EqualFold(emp.FirstName, timeEntry.Person.Firstname) &&
							strings.EqualFold(emp.LastName, timeEntry.Person.Lastname)) {

						// Found a match - update the employee with the 123erfasst ID
						if emp.Erfasst123ID == "" {
							fmt.Printf("Updating employee %s %s with 123erfasst ID: %s\n",
								emp.FirstName, emp.LastName, personIdent)
							emp.Erfasst123ID = personIdent
							emp.UpdatedAt = time.Now()
							err = employeeRepo.Update(emp)
							if err != nil {
								fmt.Printf("Error updating employee %s %s with 123erfasst ID: %v\n",
									emp.FirstName, emp.LastName, err)
							}
						}

						employee = emp
						employeeFound = true
						employeesFound[personIdent] = true
						fmt.Printf("Found employee %s %s by email/name match\n",
							emp.FirstName, emp.LastName)
						break
					}
				}

				// If still not found, skip this time entry
				if !employeeFound {
					if !employeesFound[personIdent] {
						fmt.Printf("Could not find matching employee for %s (%s, email: %s)\n",
							personName, personIdent, timeEntry.Person.Mail)
					}
					continue
				}
			} else {
				// Skip if we've already tried and failed to find this employee
				continue
			}
		} else {
			employeesFound[personIdent] = true
		}

		// Check if this time entry already exists
		timeEntryExists := false
		for _, existingEntry := range employee.TimeEntries {
			if existingEntry.Date.Equal(timeEntry.DateParsed) &&
				existingEntry.StartTime.Equal(timeEntry.TimeStartParsed) &&
				existingEntry.EndTime.Equal(timeEntry.TimeEndParsed) &&
				existingEntry.ProjectID == timeEntry.Project.ID {
				timeEntryExists = true
				break
			}
		}

		// If time entry doesn't exist, add it
		if !timeEntryExists {
			// Create a new time entry
			wageTypeName := ""
			if timeEntry.WageType != nil {
				wageTypeName = timeEntry.WageType.Name
			}

			activityName := ""
			if timeEntry.Activity.Name != "" {
				activityName = timeEntry.Activity.Name
			}

			newTimeEntry := model.TimeEntry{
				ID:          primitive.NewObjectID(),
				Date:        timeEntry.DateParsed,
				StartTime:   timeEntry.TimeStartParsed,
				EndTime:     timeEntry.TimeEndParsed,
				Duration:    timeEntry.Duration,
				ProjectID:   timeEntry.Project.ID,
				ProjectName: timeEntry.Project.Name,
				Activity:    activityName,
				WageType:    wageTypeName,
				Source:      "123erfasst",
			}

			// Add to employee's time entries
			employee.TimeEntries = append(employee.TimeEntries, newTimeEntry)
			employee.UpdatedAt = time.Now()

			// Update employee in database
			err = employeeRepo.Update(employee)
			if err != nil {
				fmt.Printf("Error updating employee %s %s: %v\n",
					employee.FirstName, employee.LastName, err)
				continue
			}

			updatedCount++
			fmt.Printf("Updated employee %s %s with time entry for project: %s on %s\n",
				employee.FirstName, employee.LastName, timeEntry.Project.Name,
				timeEntry.DateParsed.Format("2006-01-02"))
		}
	}

	fmt.Printf("Sync complete. Updated %d time entries across employees.\n", updatedCount)
	return updatedCount, nil

}

// GetSyncStatus returns the synchronization status and settings
func (s *Erfasst123Service) GetSyncStatus() (gin.H, error) {
	// Get auto-sync status
	autoSync, err := s.IsAutoSyncEnabled()
	if err != nil {
		autoSync = false
	}

	// Get last sync time
	lastSync, err := s.GetLastSyncTime()
	if err != nil {
		lastSync = time.Time{}
	}

	// Get sync start date
	startDate, err := s.GetSyncStartDate()
	if err != nil {
		// Default to start of current year
		startDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}

	// Format last sync time for display
	var lastSyncFormatted string
	if lastSync.IsZero() {
		lastSyncFormatted = "Nie"
	} else {
		lastSyncFormatted = lastSync.Format("02.01.2006 15:04:05")
	}

	return gin.H{
		"autoSync":  autoSync,
		"lastSync":  lastSyncFormatted,
		"startDate": startDate,
	}, nil

}
