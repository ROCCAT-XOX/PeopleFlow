package service

import (
	"PeopleFlow/backend/repository"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TimebutlerService verwaltet die Integration mit Timebutler
type TimebutlerService struct {
	integrationRepo *repository.IntegrationRepository
}

// GetApiKey ruft den gespeicherten API-Schlüssel ab
func (s *TimebutlerService) GetApiKey() (string, error) {
	return s.integrationRepo.GetApiKey("timebutler")
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
