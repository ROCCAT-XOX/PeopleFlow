package handler

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

// PasswordResetHandler verwaltet Passwort-Reset-Anfragen
type PasswordResetHandler struct {
	userRepo  *repository.UserRepository
	emailService *service.EmailService
}

// PasswordResetToken speichert Passwort-Reset-Tokens
type PasswordResetToken struct {
	Token     string    `bson:"token" json:"token"`
	UserEmail string    `bson:"userEmail" json:"userEmail"`
	ExpiresAt time.Time `bson:"expiresAt" json:"expiresAt"`
	Used      bool      `bson:"used" json:"used"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

// In-Memory-Speicher für Reset-Tokens (in Produktion sollte dies in der Datenbank gespeichert werden)
var resetTokens = make(map[string]*PasswordResetToken)

// NewPasswordResetHandler erstellt einen neuen PasswordResetHandler
func NewPasswordResetHandler() *PasswordResetHandler {
	return &PasswordResetHandler{
		userRepo:     repository.NewUserRepository(),
		emailService: service.NewEmailService(),
	}
}

// RequestPasswordReset verarbeitet Passwort-Reset-Anfragen
func (h *PasswordResetHandler) RequestPasswordReset(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "E-Mail-Adresse ist erforderlich",
		})
		return
	}

	// Prüfen, ob E-Mail-Service konfiguriert ist
	if !h.emailService.IsEmailConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "E-Mail-Service ist nicht konfiguriert",
		})
		return
	}

	// Benutzer suchen
	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		// Aus Sicherheitsgründen geben wir auch bei nicht existierenden Benutzern eine Erfolgsantwort
		log.Printf("Passwort-Reset-Anfrage für nicht existierende E-Mail: %s", email)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Falls ein Konto mit dieser E-Mail-Adresse existiert, wurde eine Reset-E-Mail gesendet",
		})
		return
	}

	// Reset-Token generieren
	token, err := generateResetToken()
	if err != nil {
		log.Printf("Fehler beim Generieren des Reset-Tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Interner Serverfehler",
		})
		return
	}

	// Token speichern (1 Stunde gültig)
	resetToken := &PasswordResetToken{
		Token:     token,
		UserEmail: email,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}
	resetTokens[token] = resetToken

	// Reset-E-Mail senden
	err = h.emailService.SendPasswordResetEmail(email, token)
	if err != nil {
		log.Printf("Fehler beim Senden der Reset-E-Mail an %s: %v", email, err)
		// Token entfernen bei Fehler
		delete(resetTokens, token)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Senden der E-Mail",
		})
		return
	}

	// Aktivität loggen
	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		user.ID,
		user.FirstName+" "+user.LastName,
		user.ID,
		"password_reset",
		"Passwort-Reset",
		"Passwort-Reset-E-Mail angefordert",
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Falls ein Konto mit dieser E-Mail-Adresse existiert, wurde eine Reset-E-Mail gesendet",
	})
}

// ShowPasswordResetForm zeigt das Passwort-Reset-Formular an
func (h *PasswordResetHandler) ShowPasswordResetForm(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":   "Ungültiger Reset-Link",
			"message": "Der Link ist ungültig oder beschädigt.",
		})
		return
	}

	// Token validieren
	resetToken, exists := resetTokens[token]
	if !exists || resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":   "Ungültiger oder abgelaufener Reset-Link",
			"message": "Der Link ist ungültig, bereits verwendet oder abgelaufen. Bitte fordern Sie einen neuen Reset-Link an.",
		})
		return
	}

	// Reset-Formular anzeigen
	c.HTML(http.StatusOK, "password_reset.html", gin.H{
		"token": token,
	})
}

// ResetPassword verarbeitet die Passwort-Reset-Anfrage
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	newPassword := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if token == "" || newPassword == "" || confirmPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Alle Felder sind erforderlich",
		})
		return
	}

	if newPassword != confirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Passwörter stimmen nicht überein",
		})
		return
	}

	// Passwort-Stärke prüfen
	if len(newPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Passwort muss mindestens 8 Zeichen lang sein",
		})
		return
	}

	// Token validieren
	resetToken, exists := resetTokens[token]
	if !exists || resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Ungültiger oder abgelaufener Reset-Token",
		})
		return
	}

	// Benutzer suchen
	user, err := h.userRepo.FindByEmail(resetToken.UserEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Benutzer nicht gefunden",
		})
		return
	}

	// Passwort hashen
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Fehler beim Hashen des Passworts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Interner Serverfehler",
		})
		return
	}

	// Passwort aktualisieren
	user.Password = string(hashedPassword)
	err = h.userRepo.Update(user)
	if err != nil {
		log.Printf("Fehler beim Aktualisieren des Passworts für Benutzer %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Fehler beim Speichern des neuen Passworts",
		})
		return
	}

	// Token als verwendet markieren
	resetToken.Used = true
	resetTokens[token] = resetToken

	// Aktivität loggen
	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		user.ID,
		user.FirstName+" "+user.LastName,
		user.ID,
		"password_reset",
		"Passwort-Reset",
		"Passwort erfolgreich zurückgesetzt",
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Passwort erfolgreich zurückgesetzt",
	})
}

// generateResetToken generiert einen sicheren Reset-Token
func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CleanupExpiredTokens entfernt abgelaufene Tokens (sollte regelmäßig aufgerufen werden)
func CleanupExpiredTokens() {
	now := time.Now()
	for token, resetToken := range resetTokens {
		if now.After(resetToken.ExpiresAt) {
			delete(resetTokens, token)
		}
	}
}