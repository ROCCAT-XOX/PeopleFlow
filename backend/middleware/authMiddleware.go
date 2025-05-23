package middleware

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/utils"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware ist eine Middleware für die Benutzerauthentifizierung
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Token aus dem Cookie oder Auth-Header extrahieren
		tokenString, err := extractToken(c)
		if err != nil {
			// Kein Token gefunden, zum Login umleiten
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Token validieren
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			// Ungültiges Token, zum Login umleiten
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Benutzer aus der Datenbank abrufen
		userRepo := repository.NewUserRepository()
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			// Benutzer nicht gefunden, zum Login umleiten
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Überprüfen, ob der Benutzer aktiv ist
		if user.Status != model.StatusActive {
			// Benutzer inaktiv, zum Login umleiten
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Benutzer und Claims an den Kontext weitergeben
		c.Set("user", user)
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// AdminMiddleware ist eine Middleware für administrative Operationen
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != string(model.RoleAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Nur Administratoren dürfen diese Aktion ausführen"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractToken extrahiert das JWT-Token aus dem Cookie oder Header
func extractToken(c *gin.Context) (string, error) {
	// Zuerst nach Cookie suchen
	token, err := c.Cookie("token")
	if err == nil && token != "" {
		return token, nil
	}

	// Dann nach Authorization Header suchen
	bearerToken := c.GetHeader("Authorization")
	if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer "), nil
	}

	return "", errors.New("kein Token gefunden")
}

// CORSMiddleware aktiviert CORS für Anfragen vom Frontend
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4321")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle OPTIONS Methode
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
