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

	// Status der automatischen Synchronisierung für 123erfasst abrufen
	autoSync, err := h.erfasst123Service.IsAutoSyncEnabled()
	if err != nil {
		autoSync = false
	}

	// Letzte Synchronisierung für 123erfasst abrufen
	lastSync, err := h.erfasst123Service.GetLastSyncTime()
	var lastSyncFormatted string
	if err != nil || lastSync.IsZero() {
		lastSyncFormatted = "Nie"
	} else {
		lastSyncFormatted = lastSync.Format("02.01.2006 15:04:05")
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
			"autoSync":  autoSync,
			"lastSync":  lastSyncFormatted,
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
	erfasst123Connected := h.erfasst123Service.IsConnected()

	c.HTML(http.StatusOK, "integration_settings.html", gin.H{
		"title":               "Integration Einstellungen",
		"active":              "settings",
		"user":                userModel.FirstName + " " + userModel.LastName,
		"email":               userModel.Email,
		"year":                time.Now().Year(),
		"userRole":            userRole,
		"timebutlerConnected": timebutlerConnected,
		"erfasst123Connected": erfasst123Connected,
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

// SyncTimebutlerHolidayEntitlements synchronizes Timebutler holiday entitlements with PeopleFlow employees
func (h *IntegrationHandler) SyncTimebutlerHolidayEntitlements(c *gin.Context) {
	// Check if Timebutler is connected
	if !h.timebutlerService.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Timebutler is not connected",
		})
		return
	}

	// Year from request or use current year
	year := c.DefaultQuery("year", fmt.Sprintf("%d", time.Now().Year()))

	// Perform synchronization
	updatedCount, err := h.timebutlerService.SyncHolidayEntitlements(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing holiday entitlements: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d employees were updated with holiday entitlements", updatedCount),
		"updatedCount": updatedCount,
	})
}

////////////////////        123Erfasst Integration /////////////////////

// SaveErfasst123Credentials speichert die Anmeldedaten für 123erfasst
func (h *IntegrationHandler) SaveErfasst123Credentials(c *gin.Context) {
	email := c.PostForm("erfasst123-email")
	password := c.PostForm("erfasst123-password")
	syncStartDate := c.PostForm("erfasst123-sync-start-date")

	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "E-Mail und Passwort sind erforderlich",
		})
		return
	}

	err := h.erfasst123Service.SaveCredentials(email, password, syncStartDate)
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

// SyncErfasst123Projects synchronizes 123erfasst project data with PeopleFlow employees
func (h *IntegrationHandler) SyncErfasst123Projects(c *gin.Context) {
	// Check if 123erfasst is connected
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst is not connected",
		})
		return
	}

	// Get date range from request
	startDate := c.DefaultQuery("startDate", "")

	// If no startDate provided, use saved sync start date
	if startDate == "" {
		savedStartDate, err := h.erfasst123Service.GetSyncStartDate()
		if err != nil {
			// Fallback to current month
			now := time.Now()
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			startDate = startOfMonth.Format("2006-01-02")
		} else {
			startDate = savedStartDate
		}
		fmt.Printf("Verwende gespeichertes Startdatum für Projekte: %s\n", startDate)
	}

	// End date
	endDate := c.DefaultQuery("endDate", "")
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Perform synchronization
	updatedCount, err := h.erfasst123Service.SyncErfasst123Projects(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing projects: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d employees were updated with project assignments", updatedCount),
		"updatedCount": updatedCount,
		"dateRange": gin.H{
			"startDate": startDate,
			"endDate":   endDate,
		},
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

// SyncErfasst123TimeEntries synchronizes 123erfasst time entries with PeopleFlow employees
func (h *IntegrationHandler) SyncErfasst123TimeEntries(c *gin.Context) {
	// Check if 123erfasst is connected
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst is not connected",
		})
		return
	}

	// HIER DIE ÄNDERUNG: Prüfe ob startDate übergeben wurde
	startDate := c.DefaultQuery("startDate", "")

	// NEUE ZEILEN: Wenn kein startDate übergeben wurde, hole es aus den Einstellungen
	if startDate == "" {
		savedStartDate, err := h.erfasst123Service.GetSyncStartDate()
		if err != nil {
			// Fallback auf Beginn des aktuellen Jahres
			startDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		} else {
			startDate = savedStartDate
		}
		fmt.Printf("Verwende gespeichertes Startdatum: %s\n", startDate)
	}

	// ÄNDERUNG: EndDate auch prüfen
	endDate := c.DefaultQuery("endDate", "")
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Rest der Funktion bleibt UNVERÄNDERT
	// Perform synchronization
	updatedCount, err := h.erfasst123Service.SyncErfasst123TimeEntries(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing time entries: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d employees were updated with time entries", updatedCount),
		"updatedCount": updatedCount,
		"dateRange": gin.H{
			"startDate": startDate,
			"endDate":   endDate,
		},
	})
}

// GetErfasst123SyncStatus returns the synchronization status for 123erfasst
func (h *IntegrationHandler) GetErfasst123SyncStatus(c *gin.Context) {
	// Check if 123erfasst is connected
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst is not connected",
		})
		return
	}

	// Get sync status
	status, err := h.erfasst123Service.GetSyncStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error getting sync status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// SetErfasst123AutoSync enables or disables automatic synchronization for 123erfasst
func (h *IntegrationHandler) SetErfasst123AutoSync(c *gin.Context) {
	// Get enabled status from request
	enabledStr := c.PostForm("enabled")
	enabled := enabledStr == "true"

	// Set auto-sync status
	err := h.erfasst123Service.SetAutoSync(enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error setting auto-sync: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Auto-Sync %s", map[bool]string{true: "aktiviert", false: "deaktiviert"}[enabled]),
	})
}

// SetErfasst123SyncStartDate sets the start date for data synchronization
func (h *IntegrationHandler) SetErfasst123SyncStartDate(c *gin.Context) {
	// Get start date from request
	startDate := c.PostForm("startDate")
	if startDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Startdatum ist erforderlich",
		})
		return
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ungültiges Datumsformat. Bitte verwende YYYY-MM-DD",
		})
		return
	}

	// Set sync start date
	err = h.erfasst123Service.SetSyncStartDate(startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error setting sync start date: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Synchronisierungsstartdatum wurde aktualisiert",
	})
}

// TriggerErfasst123FullSync triggers a full synchronization of 123erfasst data
func (h *IntegrationHandler) TriggerErfasst123FullSync(c *gin.Context) {
	// Check if 123erfasst is connected
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst is not connected",
		})
		return
	}

	// Get sync start date
	startDate, err := h.erfasst123Service.GetSyncStartDate()
	if err != nil {
		startDate = time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}

	// Current date as end date
	endDate := time.Now().Format("2006-01-02")

	// Sync employees
	empCount, err := h.erfasst123Service.SyncErfasst123Employees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing employees: " + err.Error(),
		})
		return
	}

	// Sync projects
	projCount, err := h.erfasst123Service.SyncErfasst123Projects(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing projects: " + err.Error(),
		})
		return
	}

	// Sync time entries
	timeCount, err := h.erfasst123Service.SyncErfasst123TimeEntries(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing time entries: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Synchronisierung abgeschlossen: %d Mitarbeiter, %d Projekte, %d Zeiteinträge",
			empCount, projCount, timeCount),
		"data": gin.H{
			"employeeCount": empCount,
			"projectCount":  projCount,
			"timeCount":     timeCount,
			"dateRange": gin.H{
				"startDate": startDate,
				"endDate":   endDate,
			},
		},
	})
}

func (h *IntegrationHandler) SyncErfasst123Employees(c *gin.Context) {
	// Prüfen, ob 123erfasst verbunden ist
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst is not connected",
		})
		return
	}

	// Synchronisierung durchführen
	updatedCount, err := h.erfasst123Service.SyncErfasst123Employees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error synchronizing employees: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("%d employees were updated", updatedCount),
		"updatedCount": updatedCount,
	})
}

// TestErfasst123ProjectAPI testet die Projekt-API von 123erfasst
func (h *IntegrationHandler) TestErfasst123ProjectAPI(c *gin.Context) {
	// Prüfen ob 123erfasst verbunden ist
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst ist nicht verbunden",
		})
		return
	}

	// Test durchführen
	if err := h.erfasst123Service.TestProjectAPI(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler beim Testen der Projekt-API: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Projekt-API Test abgeschlossen. Siehe Server-Logs für Details.",
	})
}

// CleanupDuplicates bereinigt doppelte Zeiteinträge in der Datenbank
func (h *IntegrationHandler) CleanupDuplicates(c *gin.Context) {
	// Prüfen ob 123erfasst verbunden ist
	if !h.erfasst123Service.IsConnected() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "123erfasst ist nicht verbunden",
		})
		return
	}

	// Bereinigung durchführen
	cleanedCount, err := h.erfasst123Service.CleanupDuplicateTimeEntries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Fehler beim Bereinigen: " + err.Error(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeEmployeeUpdated, // Verwende vorhandenen Type
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		userModel.ID,
		"system",
		"123erfasst Integration",
		fmt.Sprintf("Duplikate bereinigt: %d Mitarbeiter aktualisiert", cleanedCount),
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Duplikate erfolgreich bereinigt. %d Mitarbeiter wurden aktualisiert.", cleanedCount),
		"data": gin.H{
			"cleanedCount": cleanedCount,
		},
	})
}
