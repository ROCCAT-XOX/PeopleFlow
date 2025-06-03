package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SystemSettingsHandler verwaltet alle Anfragen zu System-Einstellungen
type SystemSettingsHandler struct {
	settingsRepo *repository.SystemSettingsRepository
}

// NewSystemSettingsHandler erstellt einen neuen SystemSettingsHandler
func NewSystemSettingsHandler() *SystemSettingsHandler {
	return &SystemSettingsHandler{
		settingsRepo: repository.NewSystemSettingsRepository(),
	}
}

// GetSystemSettings ruft die aktuellen System-Einstellungen ab
func (h *SystemSettingsHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Abrufen der System-Einstellungen: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateSystemSettings aktualisiert die System-Einstellungen
func (h *SystemSettingsHandler) UpdateSystemSettings(c *gin.Context) {
	// Aktuelle Einstellungen abrufen
	currentSettings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Abrufen der aktuellen Einstellungen: " + err.Error(),
		})
		return
	}

	// Formulardaten abrufen und aktualisieren
	if companyName := c.PostForm("companyName"); companyName != "" {
		currentSettings.CompanyName = companyName
	}

	if language := c.PostForm("language"); language != "" {
		currentSettings.Language = language
	}

	if state := c.PostForm("state"); state != "" {
		currentSettings.State = state
	}

	if timezone := c.PostForm("timezone"); timezone != "" {
		currentSettings.Timezone = timezone
	}

	// Einstellungen speichern
	err = h.settingsRepo.Update(currentSettings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Speichern der Einstellungen: " + err.Error(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated, // Verwende bestehenden Typ oder erstelle neuen
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		userModel.ID,
		"system",
		"System-Einstellungen",
		"System-Einstellungen aktualisiert",
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Einstellungen erfolgreich gespeichert",
		"data":    currentSettings,
	})
}

// UpdateCompanyName aktualisiert nur den Firmennamen
func (h *SystemSettingsHandler) UpdateCompanyName(c *gin.Context) {
	companyName := c.PostForm("company-name")
	if companyName == "" {
		c.Redirect(http.StatusFound, "/settings?error=empty_company_name")
		return
	}

	// Aktuelle Einstellungen abrufen
	settings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=fetch_settings")
		return
	}

	// Firmenname aktualisieren
	settings.CompanyName = companyName
	err = h.settingsRepo.Update(settings)
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=save_settings")
		return
	}

	c.Redirect(http.StatusFound, "/settings?success=company_updated")
}

// UpdateLanguage aktualisiert nur die Sprache
func (h *SystemSettingsHandler) UpdateLanguage(c *gin.Context) {
	language := c.PostForm("language")
	if language == "" {
		c.Redirect(http.StatusFound, "/settings?error=empty_language")
		return
	}

	// Aktuelle Einstellungen abrufen
	settings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=fetch_settings")
		return
	}

	// Sprache aktualisieren
	settings.Language = language
	err = h.settingsRepo.Update(settings)
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=save_settings")
		return
	}

	c.Redirect(http.StatusFound, "/settings?success=language_updated")
}

// UpdateState aktualisiert das Bundesland
func (h *SystemSettingsHandler) UpdateState(c *gin.Context) {
	// Zusätzliche Rollenprüfung
	userRole, _ := c.Get("userRole")
	if userRole != string(model.RoleAdmin) {
		c.Redirect(http.StatusFound, "/settings?error=insufficient_permissions")
		return
	}

	state := c.PostForm("state")
	if state == "" {
		c.Redirect(http.StatusFound, "/settings?error=empty_state")
		return
	}

	// Rest der Funktion bleibt unverändert...
	settings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=fetch_settings")
		return
	}

	settings.State = state
	err = h.settingsRepo.Update(settings)
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=save_settings")
		return
	}

	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	stateName := model.GermanState(state).GetDisplayName()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		userModel.ID,
		"system",
		"System-Einstellungen",
		"Bundesland geändert zu: "+stateName,
	)

	c.Redirect(http.StatusFound, "/settings?success=state_updated")
}
