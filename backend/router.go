package backend

import (
	"PeoplePilot/backend/db"
	"PeoplePilot/backend/handler"
	"PeoplePilot/backend/middleware"
	"PeoplePilot/backend/model"
	"PeoplePilot/backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// InitializeRoutes setzt alle Routen für die Anwendung auf
func InitializeRoutes(router *gin.Engine) {
	// Stelle sicher, dass die Datenbankverbindung hergestellt ist
	if err := db.ConnectDB(); err != nil {
		panic("Fehler beim Verbinden zur Datenbank")
	}

	// Public routes (keine Authentifizierung erforderlich)
	router.GET("/login", func(c *gin.Context) {
		// Token aus dem Cookie extrahieren
		tokenString, err := c.Cookie("token")
		if err == nil && tokenString != "" {
			// Token validieren
			_, err := utils.ValidateJWT(tokenString)
			if err == nil {
				// Gültiges Token, zum Dashboard umleiten
				c.Redirect(http.StatusFound, "/dashboard")
				return
			}
		}

		// Kein Token oder ungültiges Token, Login-Seite anzeigen
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
			"year":  time.Now().Year(),
		})
	})

	// Auth-Handler erstellen
	authHandler := handler.NewAuthHandler()
	router.POST("/auth", authHandler.Login)
	router.GET("/logout", authHandler.Logout)

	// Auth middleware für geschützte Routen
	authorized := router.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		currentYear := time.Now().Year()
		// Root-Pfad zum Dashboard umleiten
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})

		// Dashboard
		authorized.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)

			// Beispielhafte Daten für das Dashboard
			recentEmployees := []gin.H{
				{
					"ID":           "1",
					"Name":         "Max Mustermann",
					"Position":     "Software Developer",
					"Status":       "Aktiv",
					"ProfileImage": "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
				},
				{
					"ID":           "2",
					"Name":         "Erika Musterfrau",
					"Position":     "HR Manager",
					"Status":       "Im Urlaub",
					"ProfileImage": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
				},
				{
					"ID":           "3",
					"Name":         "John Doe",
					"Position":     "Marketing Specialist",
					"Status":       "Remote",
					"ProfileImage": "https://images.unsplash.com/photo-1570295999919-56ceb5ecca61?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
				},
				{
					"ID":           "4",
					"Name":         "Jane Smith",
					"Position":     "Finance Director",
					"Status":       "Aktiv",
					"ProfileImage": "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
				},
			}

			upcomingReviews := []gin.H{
				{
					"EmployeeName": "Max Mustermann",
					"ReviewType":   "Leistungsbeurteilung",
					"Date":         "18.04.2025",
				},
				{
					"EmployeeName": "Erika Musterfrau",
					"ReviewType":   "Beförderungsgespräch",
					"Date":         "22.04.2025",
				},
				{
					"EmployeeName": "John Doe",
					"ReviewType":   "Einarbeitung",
					"Date":         "25.04.2025",
				},
			}

			recentActivities := []gin.H{
				{
					"IconBgClass": "bg-green-500",
					"IconSVG":     "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M10 18a8 8 0 100-16 8 8 0 000 16zm.75-13a.75.75 0 00-1.5 0v5.5a.75.75 0 001.5 0V5z\" clip-rule=\"evenodd\" /></svg>",
					"Message":     "<a href=\"#\" class=\"font-medium text-gray-900\">Max Mustermann</a> wurde als neuer Mitarbeiter hinzugefügt",
					"Time":        "Heute 08:30 Uhr",
					"IsLast":      false,
				},
				{
					"IconBgClass": "bg-blue-500",
					"IconSVG":     "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z\" clip-rule=\"evenodd\" /></svg>",
					"Message":     "<a href=\"#\" class=\"font-medium text-gray-900\">Erika Musterfrau</a> hat einen Urlaubsantrag eingereicht",
					"Time":        "Gestern 17:45 Uhr",
					"IsLast":      false,
				},
				{
					"IconBgClass": "bg-yellow-500",
					"IconSVG":     "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z\" clip-rule=\"evenodd\" /></svg>",
					"Message":     "<a href=\"#\" class=\"font-medium text-gray-900\">John Doe</a> benötigt Hilfe bei der Einrichtung seines Accounts",
					"Time":        "14.04.2025 10:23 Uhr",
					"IsLast":      true,
				},
			}

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title":               "Dashboard",
				"active":              "dashboard",
				"user":                userModel.FirstName + " " + userModel.LastName,
				"email":               userModel.Email,
				"year":                currentYear,
				"totalEmployees":      44,
				"pendingRequests":     5,
				"upcomingReviews":     3,
				"expiredDocuments":    2,
				"recentEmployees":     recentEmployees,
				"upcomingReviewsList": upcomingReviews,
				"recentActivities":    recentActivities,
			})
		})

		// Weitere Routen hier hinzufügen, z.B. für Mitarbeiter, Einstellungen, etc.
		authorized.GET("/employees", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)

			c.HTML(http.StatusOK, "employees.html", gin.H{
				"title":  "Mitarbeiter",
				"active": "employees",
				"user":   userModel.FirstName + " " + userModel.LastName,
				"email":  userModel.Email,
				"year":   currentYear,
			})
		})
	}
}
