// backend/handler/integration_handler.go
package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// IntegrationHandler aktualisieren, um erfasst123Service hinzuzufügen
type IntegrationHandler struct {
	timebutlerService *service.TimebutlerService
	erfasst123Service *service.Erfasst123Service
}

// NewIntegrationHandler anpassen
func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{
		timebutlerService: service.NewTimebutlerService(),
		erfasst123Service: service.NewErfasst123Service(),
	}
}

// SaveTimebutlerApiKey speichert den API-Schlüssel für Timebutler
func (h *IntegrationHandler) SaveTimebutlerApiKey(c *gin.Context) {
	apiKey := c.PostForm("timebutler-api")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "API-Schlüssel ist erforderlich",
		})
		return
	}

	err := h.timebutlerService.SaveApiKey(apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler beim Speichern des API-Schlüssels: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Timebutler-Integration erfolgreich konfiguriert",
	})
}

// GetIntegrationStatus gibt den Status aller Integrationen zurück
func (h *IntegrationHandler) GetIntegrationStatus(c *gin.Context) {
	timebutlerConnected := h.timebutlerService.IsConnected()
	erfasst123Connected := h.erfasst123Service.IsConnected() // Neu hinzugefügt

	// Prüfen, ob ein API-Schlüssel für Timebutler vorhanden ist
	hasApiKey := false
	if apiKey, err := h.timebutlerService.GetApiKey(); err == nil && apiKey != "" {
		hasApiKey = true
	}

	// Prüfen, ob Anmeldedaten für 123erfasst vorhanden sind
	hasErfasst123Credentials := false
	if email, password, err := h.erfasst123Service.GetCredentials(); err == nil && email != "" && password != "" {
		hasErfasst123Credentials = true
	}

	c.JSON(http.StatusOK, gin.H{
		"timebutler": gin.H{
			"connected": timebutlerConnected,
			"name":      "Timebutler",
			"hasApiKey": hasApiKey,
		},
		"123erfasst": gin.H{ // Neu hinzugefügt
			"connected": erfasst123Connected,
			"name":      "123erfasst",
			"hasApiKey": hasErfasst123Credentials,
		},
		"awork": gin.H{
			"connected": false,
			"name":      "AWork",
			"hasApiKey": false,
		},
	})
}

// TestTimebutlerConnection testet die Verbindung zu Timebutler
func (h *IntegrationHandler) TestTimebutlerConnection(c *gin.Context) {
	isConnected := h.timebutlerService.IsConnected()

	c.JSON(http.StatusOK, gin.H{
		"connected": isConnected,
	})
}

// GetIntegrationSettings gibt die Settings für die Integrationsseite zurück
func (h *IntegrationHandler) GetIntegrationSettings(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole, _ := c.Get("userRole")

	timebutlerConnected := h.timebutlerService.IsConnected()

	c.HTML(http.StatusOK, "integration_settings.html", gin.H{
		"title":               "Integration Einstellungen",
		"active":              "settings",
		"user":                userModel.FirstName + " " + userModel.LastName,
		"email":               userModel.Email,
		"year":                time.Now().Year(),
		"userRole":            userRole,
		"timebutlerConnected": timebutlerConnected,
	})
}

// SyncTimebutlerUsers synchronisiert Timebutler-Benutzer mit PeopleFlow-Mitarbeitern
func (h *IntegrationHandler) SyncTimebutlerUsers(c *gin.Context) {
	// Prüfen, ob Timebutler verbunden ist
	if !h.timebutlerService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Timebutler ist nicht verbunden",
		})
		return
	}

	// Synchronisierung durchführen
	updatedCount, err := h.timebutlerService.SyncTimebutlerUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler bei der Synchronisierung: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d Mitarbeiter wurden synchronisiert", updatedCount),
		"updatedCount": updatedCount,
	})
}

// SyncTimebutlerAbsences synchronisiert Timebutler-Abwesenheiten mit PeopleFlow-Mitarbeitern
func (h *IntegrationHandler) SyncTimebutlerAbsences(c *gin.Context) {
	// Prüfen, ob Timebutler verbunden ist
	if !h.timebutlerService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Timebutler ist nicht verbunden",
		})
		return
	}

	// Jahr aus der Anfrage holen oder aktuelles Jahr verwenden
	year := c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year()))

	// Synchronisierung durchführen
	updatedCount, err := h.timebutlerService.SyncTimebutlerAbsences(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler bei der Synchronisierung: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d Mitarbeiter mit Abwesenheiten wurden synchronisiert", updatedCount),
		"updatedCount": updatedCount,
	})
}

////////////////////        123Erfasst Integration /////////////////////

// SaveErfasst123Credentials speichert die Anmeldedaten für 123erfasst
func (h *IntegrationHandler) SaveErfasst123Credentials(c *gin.Context) {
	email := c.PostForm("erfasst123-email")
	password := c.PostForm("erfasst123-password")

	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "E-Mail und Passwort sind erforderlich",
		})
		return
	}

	err := h.erfasst123Service.SaveCredentials(email, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler beim Speichern der Anmeldedaten: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "123erfasst-Integration erfolgreich konfiguriert",
	})
}

// TestErfasst123Connection testet die Verbindung zu 123erfasst
func (h *IntegrationHandler) TestErfasst123Connection(c *gin.Context) {
	isConnected := h.erfasst123Service.IsConnected()

	c.JSON(http.StatusOK, gin.H{
		"connected": isConnected,
	})
}

// SyncErfasst123Employees synchronisiert 123erfasst-Mitarbeiter mit PeopleFlow-Mitarbeitern
func (h *IntegrationHandler) SyncErfasst123Employees(c *gin.Context) {
	// Prüfen, ob 123erfasst verbunden ist
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst ist nicht verbunden",
		})
		return
	}

	// Synchronisierung durchführen
	updatedCount, err := h.erfasst123Service.SyncErfasst123Employees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler bei der Synchronisierung: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d Mitarbeiter wurden synchronisiert", updatedCount),
		"updatedCount": updatedCount,
	})
}

// RemoveErfasst123Integration entfernt die 123erfasst-Integration
func (h *IntegrationHandler) RemoveErfasst123Integration(c *gin.Context) {
	integrationRepo := repository.NewIntegrationRepository()

	// Integration als inaktiv markieren
	err := integrationRepo.SetIntegrationStatus("123erfasst", false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler beim Entfernen der Integration: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "123erfasst-Integration erfolgreich entfernt",
	})
}
