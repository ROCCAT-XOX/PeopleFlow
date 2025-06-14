// backend/middleware/roleMiddleware.go

package middleware

import (
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"fmt"
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

// HRMiddleware prüft, ob der Benutzer ein HR-Mitarbeiter ist, aber keine höheren Rollen bearbeiten darf
func HRMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Aktuelle Benutzerrolle
		userRole, _ := c.Get("userRole")

		// Ziel-Benutzer ID aus dem Parameter
		targetID := c.Param("id")

		// Wenn kein Target-ID vorhanden ist (z.B. bei einer Liste), einfach durchlassen
		if targetID == "" {
			c.Next()
			return
		}

		// Wenn HR-Rolle, dann prüfen, ob der Zielbenutzer kein Admin oder Manager ist
		if userRole == string(model.RoleHR) {
			userRepo := repository.NewUserRepository()
			targetUser, err := userRepo.FindByID(targetID)

			// Wenn der Zielbenutzer existiert und ein Admin oder Manager ist
			if err == nil && (targetUser.Role == model.RoleAdmin || targetUser.Role == model.RoleManager) {
				c.HTML(http.StatusForbidden, "error.html", gin.H{
					"title":   "Zugriff verweigert",
					"message": "Sie haben keine Berechtigung, Administratoren oder Manager zu bearbeiten.",
					"year":    time.Now().Year(),
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// SalaryViewMiddleware beschränkt den Zugriff auf Gehaltsdaten
func SalaryViewMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get("userRole")

		// Nur Admin und Manager dürfen Gehaltsdaten sehen
		if userRole != string(model.RoleAdmin) && userRole != string(model.RoleManager) {
			c.Set("hideSalary", true)
			fmt.Println("Salary hidden")
		} else {
			c.Set("hideSalary", false) // Explizit auf false setzen

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

		// Angeforderte ID aus dem Parameter ODER aus dem Formular abrufen
		requestedID := c.Param("id")
		if requestedID == "" {
			// Versuche ID aus dem Formular zu lesen (für POST-Requests wie Passwortänderung)
			requestedID = c.PostForm("id")
		}

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
