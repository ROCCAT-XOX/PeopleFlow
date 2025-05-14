// backend/handler/integration_handler.go
package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// IntegrationHandler verwaltet alle Anfragen zu Integrationen
type IntegrationHandler struct {
	timebutlerService *service.TimebutlerService
}

// NewIntegrationHandler erstellt einen neuen IntegrationHandler
func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{
		timebutlerService: service.NewTimebutlerService(),
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

	// Prüfen, ob ein API-Schlüssel für Timebutler vorhanden ist
	hasApiKey := false
	if apiKey, err := h.timebutlerService.GetApiKey(); err == nil && apiKey != "" {
		hasApiKey = true
	}

	c.JSON(http.StatusOK, gin.H{
		"timebutler": gin.H{
			"connected": timebutlerConnected,
			"name":      "Timebutler",
			"hasApiKey": hasApiKey, // Hinzugefügt
		},
		"awork": gin.H{
			"connected": false,
			"name":      "AWork",
			"hasApiKey": false, // Hinzugefügt
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
