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
	"net/http"
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
func (s *Erfasst123Service) SaveCredentials(email, password string) error {
	// Testen, ob die Anmeldedaten funktionieren
	if err := s.testConnection(email, password); err != nil {
		return err
	}

	// Anmeldedaten zusammen speichern (werden im Repository verschlüsselt)
	credentials := fmt.Sprintf("%s:%s", email, password)
	if err := s.integrationRepo.SaveApiKey("123erfasst", credentials); err != nil {
		return err
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
