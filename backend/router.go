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
		// In der InitializeRoutes-Funktion nach der Deklaration der Auth-Middleware
		userHandler := handler.NewUserHandler()

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

			// Repository für Aktivitätsdaten
			activityRepo := repository.NewActivityRepository()

			// Neueste Aktivitäten abrufen
			recentActivitiesData, err := activityRepo.FindRecent(5)
			if err != nil {
				recentActivitiesData = []*model.Activity{} // Leere Liste im Fehlerfall
			}

			// Aktivitäten in ein Format konvertieren, das für die Vorlage geeignet ist
			var recentActivities []gin.H
			for i, activity := range recentActivitiesData {
				isLast := i == len(recentActivitiesData)-1

				// Nachricht formatieren
				var message string
				switch activity.Type {
				case model.ActivityTypeEmployeeAdded:
					message = fmt.Sprintf("<a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde als neuer Mitarbeiter hinzugefügt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeEmployeeUpdated:
					message = fmt.Sprintf("<a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde aktualisiert",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeVacationRequested:
					message = fmt.Sprintf("<a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> hat einen Urlaubsantrag eingereicht",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeVacationApproved:
					message = fmt.Sprintf("Urlaubsantrag von <a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> wurde genehmigt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeDocumentUploaded:
					message = fmt.Sprintf("<a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> hat ein Dokument hochgeladen",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeTrainingAdded:
					message = fmt.Sprintf("Weiterbildung für <a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> hinzugefügt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeEvaluationAdded:
					message = fmt.Sprintf("Leistungsbeurteilung für <a href=\"/employees/view/%s\" class=\"font-medium text-gray-900\">%s</a> hinzugefügt",
						activity.TargetID.Hex(), activity.TargetName)
				case model.ActivityTypeEmployeeDeleted:
					message = fmt.Sprintf("Mitarbeiter <span class=\"font-medium text-gray-900\">%s</span> wurde entfernt",
						activity.TargetName)
				default:
					message = activity.Description
				}

				recentActivities = append(recentActivities, gin.H{
					"IconBgClass": activity.GetIconClass(),
					"IconSVG":     activity.GetIconSVG(),
					"Message":     message,
					"Time":        activity.FormatTimeAgo(),
					"IsLast":      isLast,
				})
			}

			// Falls keine Aktivitäten gefunden wurden, verwenden wir Beispieldaten
			if len(recentActivities) == 0 {
				recentActivities = []gin.H{
					{
						"IconBgClass": "bg-gray-500",
						"IconSVG":     "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z\" clip-rule=\"evenodd\" /></svg>",
						"Message":     "Keine Aktivitäten vorhanden",
						"Time":        "Jetzt",
						"IsLast":      true,
					},
				}
			}

			// Beispielhafte Daten für das Dashboard - Mitarbeiterübersicht
			recentEmployees := []gin.H{}

			// Wenn wir tatsächliche Mitarbeiterdaten haben, diese verwenden
			if len(allEmployees) > 0 {
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
						profileImg = "/static/img/default-avatar.png"
					}

					recentEmployees = append(recentEmployees, gin.H{
						"ID":           emp.ID.Hex(),
						"Name":         emp.FirstName + " " + emp.LastName,
						"Position":     emp.Position,
						"Status":       status,
						"ProfileImage": profileImg,
					})
				}
			} else {
				// Beispielhafte Daten, falls keine echten Daten vorhanden sind
				recentEmployees = []gin.H{
					{
						"ID":           "1",
						"Name":         "Max Mustermann",
						"Position":     "Software Developer",
						"Status":       "Aktiv",
						"ProfileImage": "/static/img/default-avatar.png",
					},
					{
						"ID":           "2",
						"Name":         "Erika Musterfrau",
						"Position":     "HR Manager",
						"Status":       "Im Urlaub",
						"ProfileImage": "/static/img/default-avatar.png",
					},
					{
						"ID":           "3",
						"Name":         "John Doe",
						"Position":     "Marketing Specialist",
						"Status":       "Remote",
						"ProfileImage": "/static/img/default-avatar.png",
					},
					{
						"ID":           "4",
						"Name":         "Jane Smith",
						"Position":     "Finance Director",
						"Status":       "Aktiv",
						"ProfileImage": "/static/img/default-avatar.png",
					},
				}
			}

			// Anzahl abgelaufener Dokumente (in einer echten Anwendung würden wir dies berechnen)
			expiredDocuments := 2

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
				"expiredDocuments":        expiredDocuments,
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

		// Benutzerprofilrouten
		authorized.GET("/profile", userHandler.ShowUserProfile)

		// Einstellungsrouten (für alle Benutzer)
		authorized.GET("/settings", userHandler.ShowSettings)

		// Benutzerverwaltungsrouten (für Administratoren)
		authorized.POST("/users/add", middleware.RoleMiddleware(model.RoleAdmin), userHandler.AddUser)
		authorized.POST("/users/edit/:id", middleware.RoleMiddleware(model.RoleAdmin), userHandler.UpdateUser)
		authorized.DELETE("/users/delete/:id", middleware.RoleMiddleware(model.RoleAdmin), userHandler.DeleteUser)

		// Passwortänderungsroute - ein Benutzer kann nur sein eigenes Passwort ändern
		authorized.POST("/users/change-password", middleware.SelfOrAdminMiddleware(), userHandler.ChangePassword)

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
