package handler

import (
	"net/http"
	"time"

	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"
	"PeoplePilot/backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler repräsentiert den Handler für Authentifizierungsoperationen
type AuthHandler struct {
	userRepo *repository.UserRepository
}

// NewAuthHandler erstellt einen neuen AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userRepo: repository.NewUserRepository(),
	}
}

// Login verarbeitet die Login-Anfrage
func (h *AuthHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" {
		// Wenn kein E-Mail-Parameter gesendet wurde, zurück zum Login mit Fehlermeldung
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "E-Mail ist erforderlich",
			"year":  time.Now().Year(),
		})
		return
	}

	if password == "" {
		// Wenn kein Passwort gesendet wurde, zurück zum Login mit Fehlermeldung
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Passwort ist erforderlich",
			"year":  time.Now().Year(),
		})
		return
	}

	// Benutzer anhand der E-Mail finden
	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		// Benutzer nicht gefunden, zurück zum Login mit Fehlermeldung
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Ungültige E-Mail oder Passwort",
			"year":  time.Now().Year(),
		})
		return
	}

	// Überprüfen, ob das Passwort übereinstimmt
	if !user.CheckPassword(password) {
		// Passwort stimmt nicht überein, zurück zum Login mit Fehlermeldung
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Ungültige E-Mail oder Passwort",
			"year":  time.Now().Year(),
		})
		return
	}

	// Überprüfen, ob der Benutzer aktiv ist
	if user.Status != model.StatusActive {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Ihr Konto ist inaktiv",
			"year":  time.Now().Year(),
		})
		return
	}

	// JWT-Token generieren
	token, err := utils.GenerateJWT(user.ID.Hex(), string(user.Role))
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"error": "Ein interner Fehler ist aufgetreten",
			"year":  time.Now().Year(),
		})
		return
	}

	// Cookie setzen
	c.SetCookie(
		"token",
		token,
		int(time.Hour.Seconds()*24), // 24 Stunden
		"/",
		"",
		false,
		true,
	)

	// Nach erfolgreicher Anmeldung zum Dashboard weiterleiten
	c.Redirect(http.StatusFound, "/dashboard")
}

// Logout behandelt die Logout-Anfrage
func (h *AuthHandler) Logout(c *gin.Context) {
	// Token-Cookie löschen
	c.SetCookie(
		"token",
		"",
		-1, // Sofort ablaufen lassen
		"/",
		"",
		false,
		true,
	)

	// Zur Login-Seite umleiten
	c.Redirect(http.StatusFound, "/login")
}
