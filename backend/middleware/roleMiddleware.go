// backend/middleware/roleMiddleware.go
package middleware

import (
	"PeoplePilot/backend/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// RoleMiddleware prüft, ob der Benutzer die erforderliche Rolle hat
func RoleMiddleware(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rolle aus dem Kontext abrufen
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Keine Benutzerrolle gefunden"})
			c.Abort()
			return
		}

		// Prüfen, ob die Rolle des Benutzers in den erlaubten Rollen enthalten ist
		hasPermission := false
		for _, role := range allowedRoles {
			if userRole == string(role) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"title":   "Zugriff verweigert",
				"message": "Sie haben keine Berechtigung, auf diese Ressource zuzugreifen.",
				"year":    time.Now().Year(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SelfOrAdminMiddleware erlaubt Zugriff, wenn der Benutzer auf seine eigenen Daten zugreift oder ein Admin ist
func SelfOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Benutzer-ID aus dem Kontext abrufen
		userId, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Keine Benutzer-ID gefunden"})
			c.Abort()
			return
		}

		// Rolle aus dem Kontext abrufen
		userRole, _ := c.Get("userRole")

		// Angeforderte ID aus dem Parameter abrufen
		requestedID := c.Param("id")

		// Wenn der Benutzer ein Admin ist oder auf seine eigenen Daten zugreift, hat er Zugriff
		if userRole == string(model.RoleAdmin) || userId == requestedID {
			c.Next()
			return
		}

		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Zugriff verweigert",
			"message": "Sie haben keine Berechtigung, auf diese Ressource zuzugreifen.",
			"year":    time.Now().Year(),
		})
		c.Abort()
	}
}
