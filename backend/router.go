package backend

import (
	"PeoplePilot/backend/db"
	"PeoplePilot/backend/handler"
	"PeoplePilot/backend/middleware"
	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"
	"PeoplePilot/backend/service"
	"PeoplePilot/backend/utils"
	"fmt"
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
		//currentYear := time.Now().Year()
		// Root-Pfad zum Dashboard umleiten
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})

		// Dashboard
		authorized.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)

			// Repository für Mitarbeiterdaten
			employeeRepo := repository.NewEmployeeRepository()

			// Service für Kostenberechnungen initialisieren
			costService := service.NewCostService()

			// Alle Mitarbeiter abrufen
			allEmployees, err := employeeRepo.FindAll()
			if err != nil {
				allEmployees = []*model.Employee{} // Leere Liste im Fehlerfall
			}

			totalEmployees := len(allEmployees)

			// Monatliche Personalkosten berechnen
			monthlyLaborCosts := costService.CalculateMonthlyLaborCosts(allEmployees)

			// Monatliche Kostendaten für das Diagramm generieren
			monthlyCostsData := costService.GenerateMonthlyLaborCostsTrend(monthlyLaborCosts)

			// Durchschnittskosten pro Mitarbeiter berechnen
			avgCostsPerEmployee := costService.CalculateAvgCostPerEmployee(monthlyLaborCosts, totalEmployees)

			// Durchschnittliche Kosten pro Mitarbeiter über Zeit generieren
			avgCostsPerEmployeeData := costService.GenerateMonthlyLaborCostsTrend(avgCostsPerEmployee)

			// Abteilungsverteilung berechnen
			departmentLabels, departmentData := costService.CountEmployeesByDepartment(allEmployees)

			// Anstehende Bewertungen generieren
			upcomingReviewsList := costService.GenerateExpectedReviews(allEmployees)

			// Personalkostenverteilung nach Abteilung berechnen
			deptCostsLabels, deptCostsData := costService.CalculateCostsByDepartment(allEmployees)

			// Altersstruktur berechnen
			ageGroups, ageCounts := costService.CalculateAgeDistribution(allEmployees)

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

			// Wenn wir tatsächliche Mitarbeiterdaten haben, diese verwenden
			if len(allEmployees) > 0 {
				recentEmployees = []gin.H{}
				maxToShow := 4
				if len(allEmployees) < maxToShow {
					maxToShow = len(allEmployees)
				}

				for i := 0; i < maxToShow; i++ {
					emp := allEmployees[i]
					status := "Aktiv"
					switch emp.Status {
					case model.EmployeeStatusInactive:
						status = "Inaktiv"
					case model.EmployeeStatusOnLeave:
						status = "Im Urlaub"
					case model.EmployeeStatusRemote:
						status = "Remote"
					}

					profileImg := emp.ProfileImage
					if profileImg == "" {
						profileImg = "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80"
					}

					recentEmployees = append(recentEmployees, gin.H{
						"ID":           emp.ID.Hex(),
						"Name":         emp.FirstName + " " + emp.LastName,
						"Position":     emp.Position,
						"Status":       status,
						"ProfileImage": profileImg,
					})
				}
			}

			// Beispielhafte Aktivitäten
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

			// Formatieren der monatlichen Personalkosten
			formattedLaborCosts := fmt.Sprintf("%.2f", monthlyLaborCosts)

			// Daten an das Template übergeben
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title":                   "Dashboard",
				"active":                  "dashboard",
				"user":                    userModel.FirstName + " " + userModel.LastName,
				"email":                   userModel.Email,
				"year":                    time.Now().Year(),
				"totalEmployees":          totalEmployees,
				"monthlyLaborCosts":       formattedLaborCosts,
				"upcomingReviews":         len(upcomingReviewsList),
				"expiredDocuments":        2,
				"recentEmployees":         recentEmployees,
				"upcomingReviewsList":     upcomingReviewsList,
				"recentActivities":        recentActivities,
				"monthlyCostsData":        monthlyCostsData,
				"avgCostsPerEmployeeData": avgCostsPerEmployeeData,
				"departmentLabels":        departmentLabels,
				"departmentData":          departmentData,
				"deptCostsLabels":         deptCostsLabels,
				"deptCostsData":           deptCostsData,
				"ageGroups":               ageGroups,
				"ageCounts":               ageCounts,
			})
		})

		employeeHandler := handler.NewEmployeeHandler()
		documentHandler := handler.NewDocumentHandler()

		// Mitarbeiter-Routen zum autorisierten Bereich hinzufügen
		authorized.GET("/employees", employeeHandler.ListEmployees)
		authorized.GET("/employees/view/:id", employeeHandler.GetEmployeeDetails)
		authorized.GET("/employees/edit/:id", employeeHandler.ShowEditEmployeeForm)
		authorized.POST("/employees/add", employeeHandler.AddEmployee)
		authorized.POST("/employees/edit/:id", employeeHandler.UpdateEmployee)
		authorized.DELETE("/employees/delete/:id", employeeHandler.DeleteEmployee)

		// Profilbil hinzufügen
		// Im router.go, innerhalb des authorized-Bereichs
		authorized.POST("/employees/:id/profile-image", employeeHandler.UploadProfileImage)

		// Dokument-Routen
		authorized.POST("/employees/:id/documents", documentHandler.UploadDocument)
		authorized.DELETE("/employees/:id/documents/:documentId", documentHandler.DeleteDocument)
		authorized.GET("/employees/:id/documents/:documentId/download", documentHandler.DownloadDocument)

		// Training-Routen
		authorized.POST("/employees/:id/trainings", documentHandler.AddTraining)
		authorized.DELETE("/employees/:id/trainings/:trainingId", documentHandler.DeleteTraining)

		// Evaluation-Routen
		authorized.POST("/employees/:id/evaluations", documentHandler.AddEvaluation)
		authorized.DELETE("/employees/:id/evaluations/:evaluationId", documentHandler.DeleteEvaluation)

		// Absence-Routen
		authorized.POST("/employees/:id/absences", documentHandler.AddAbsence)
		authorized.DELETE("/employees/:id/absences/:absenceId", documentHandler.DeleteAbsence)
		authorized.POST("/employees/:id/absences/:absenceId/approve", documentHandler.ApproveAbsence)

		// Development-Routen
		authorized.POST("/employees/:id/development", documentHandler.AddDevelopmentItem)
		authorized.DELETE("/employees/:id/development/:itemId", documentHandler.DeleteDevelopmentItem)

		// Optionale API-Endpoints für AJAX-Anfragen
		api := router.Group("/api")
		api.Use(middleware.AuthMiddleware())
		{
			api.DELETE("/employees/:id", employeeHandler.DeleteEmployee)
		}
	}
}
