package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

	// Timezone field is not currently supported in SystemSettings model
	// if timezone := c.PostForm("timezone"); timezone != "" {
	//     currentSettings.Timezone = timezone
	// }

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

	// Validiere Sprache
	validLanguages := map[string]bool{
		"de": true,
		"en": true,
		"fr": true,
	}
	if !validLanguages[language] {
		c.Redirect(http.StatusFound, "/settings?error=invalid_language")
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
	stateName := model.GermanState(state).GetLabel()
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

// UpdateEmailSettings aktualisiert die E-Mail-Konfiguration
func (h *SystemSettingsHandler) UpdateEmailSettings(c *gin.Context) {
	// Nur Admins können E-Mail-Einstellungen ändern
	userRole, _ := c.Get("userRole")
	if userRole != string(model.RoleAdmin) {
		c.Redirect(http.StatusFound, "/settings?error=insufficient_permissions")
		return
	}

	// Aktuelle Einstellungen abrufen
	settings, err := h.settingsRepo.GetSettings()
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=fetch_settings")
		return
	}

	// E-Mail-Einstellungen aus Formular lesen
	smtpHost := c.PostForm("smtp-host")
	smtpPortStr := c.PostForm("smtp-port")
	smtpUser := c.PostForm("smtp-user")
	smtpPass := c.PostForm("smtp-pass")
	fromEmail := c.PostForm("from-email")
	fromName := c.PostForm("from-name")
	useTLS := c.PostForm("use-tls") == "on"
	enabled := c.PostForm("email-enabled") == "on"

	// Port validieren
	smtpPort := 587 // Default
	if smtpPortStr != "" {
		if port, err := strconv.Atoi(smtpPortStr); err == nil {
			smtpPort = port
		}
	}

	// E-Mail-Einstellungen erstellen oder aktualisieren
	if settings.EmailNotifications == nil {
		settings.EmailNotifications = &model.EmailNotificationSettings{}
	}

	settings.EmailNotifications.SMTPHost = smtpHost
	settings.EmailNotifications.SMTPPort = smtpPort
	settings.EmailNotifications.SMTPUser = smtpUser
	settings.EmailNotifications.SMTPPass = smtpPass
	settings.EmailNotifications.FromEmail = fromEmail
	settings.EmailNotifications.FromName = fromName
	settings.EmailNotifications.UseTLS = useTLS
	settings.EmailNotifications.Enabled = enabled

	// Einstellungen speichern
	err = h.settingsRepo.Update(settings)
	if err != nil {
		c.Redirect(http.StatusFound, "/settings?error=save_email_settings")
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		userModel.ID,
		"system",
		"E-Mail-Einstellungen",
		"E-Mail-Konfiguration aktualisiert",
	)

	c.Redirect(http.StatusFound, "/settings?success=email_updated")
}

// TestEmailConfiguration testet die E-Mail-Konfiguration
func (h *SystemSettingsHandler) TestEmailConfiguration(c *gin.Context) {
	// Nur Admins können E-Mail-Tests durchführen
	userRole, _ := c.Get("userRole")
	if userRole != string(model.RoleAdmin) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Unzureichende Berechtigung",
		})
		return
	}

	// Test-E-Mail-Adresse aus Request lesen
	testEmail := c.Query("email")
	if testEmail == "" {
		// Verwende die E-Mail des aktuellen Benutzers
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Keine Test-E-Mail-Adresse angegeben",
			})
			return
		}
		userModel := user.(*model.User)
		testEmail = userModel.Email
	}

	// E-Mail-Service erstellen und Test-E-Mail senden
	emailService := service.NewEmailService()
	err := emailService.SendTestEmail(testEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Senden der Test-E-Mail: " + err.Error(),
		})
		return
	}

	// Aktivität loggen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		userModel.ID,
		userModel.FirstName+" "+userModel.LastName,
		userModel.ID,
		"system",
		"E-Mail-Test",
		"Test-E-Mail gesendet an: "+testEmail,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test-E-Mail erfolgreich gesendet",
	})
}
