package backend

import (
	"PeopleFlow/backend/db"
	"PeopleFlow/backend/handler"
	"PeopleFlow/backend/middleware"
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/service"
	"PeopleFlow/backend/utils"
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

	router.Use(middleware.CORSMiddleware())

	// Auth-Handler erstellen
	authHandler := handler.NewAuthHandler()
	router.POST("/auth", authHandler.Login)
	router.GET("/logout", authHandler.Logout)

	// Auth middleware für geschützte Routen
	authorized := router.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		// In der InitializeRoutes-Funktion nach der Deklaration der Auth-Middleware
		// Handler erstellen
		userHandler := handler.NewUserHandler()
		systemSettingsHandler := handler.NewSystemSettingsHandler()
		holidayHandler := handler.NewHolidayHandler()

		// Root-Pfad zum Dashboard umleiten
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})

		// Dashboard Route in backend/router.go - ersetze die bestehende Dashboard Route
		authorized.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)
			userRole, _ := c.Get("userRole")

			// Repository für Mitarbeiterdaten
			employeeRepo := repository.NewEmployeeRepository()
			activityRepo := repository.NewActivityRepository()

			// Alle Mitarbeiter abrufen
			allEmployees, err := employeeRepo.FindAll()
			if err != nil {
				allEmployees = []*model.Employee{} // Leere Liste im Fehlerfall
			}

			totalEmployees := len(allEmployees)
			currentDate := time.Now().Format("Monday, 02. January 2006")

			// Gemeinsame Daten für alle Rollen
			commonData := gin.H{
				"title":          "Dashboard",
				"active":         "dashboard",
				"user":           userModel.FirstName + " " + userModel.LastName,
				"email":          userModel.Email,
				"userRole":       userRole,
				"year":           time.Now().Year(),
				"currentDate":    currentDate,
				"totalEmployees": totalEmployees,
			}

			// HR-spezifisches Dashboard
			if userRole == string(model.RoleHR) {
				// HR-Service initialisieren und echte Daten berechnen
				hrService := service.NewHRService()
				hrData := hrService.CalculateHRDashboardData(allEmployees)

				// Chart-Daten generieren
				departmentLabels, departmentData := hrService.GetDepartmentLabelsAndData(hrData.DepartmentCounts)
				statusLabels, statusData := hrService.GetStatusLabelsAndData(hrData.StatusDistribution)
				ageLabels, ageData := hrService.GetAgeLabelsAndData(hrData.AgeDistribution)
				tenureLabels, tenureData := hrService.GetTenureLabelsAndData(hrData.TenureDistribution)

				// Aktivitäten für HR
				recentActivitiesData, err := activityRepo.FindRecent(5)
				if err != nil {
					recentActivitiesData = []*model.Activity{}
				}

				var recentActivities []gin.H
				for i, activity := range recentActivitiesData {
					isLast := i == len(recentActivitiesData)-1

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

				// Mitarbeiterübersicht für HR (zeige mehr Details)
				recentEmployees := []gin.H{}
				maxToShow := 8
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
						status = "Krank" // Remote wird als Krank angezeigt
					}

					recentEmployees = append(recentEmployees, gin.H{
						"ID":           emp.ID.Hex(),
						"Name":         emp.FirstName + " " + emp.LastName,
						"Position":     emp.Position,
						"Status":       status,
						"Department":   emp.Department,
						"ProfileImage": emp.ProfileImage,
						"HireDate":     emp.HireDate.Format("02.01.2006"),
						"Tenure":       hrService.CalculateTenure(emp.HireDate),
					})
				}

				// HR-spezifische Daten zu commonData hinzufügen
				for k, v := range commonData {
					commonData[k] = v
				}

				hrDataForTemplate := gin.H{
					// Grundstatistiken
					"activeEmployees":       hrData.ActiveEmployees,
					"onLeaveEmployees":      hrData.OnLeaveEmployees,
					"sickEmployees":         hrData.SickEmployees, // Geändert von remoteEmployees
					"inactiveEmployees":     hrData.InactiveEmployees,
					"newEmployeesThisMonth": hrData.NewEmployeesThisMonth,
					"currentAbsences":       hrData.CurrentAbsences,
					"upcomingAbsences":      hrData.UpcomingAbsences,
					"absenceRate":           fmt.Sprintf("%.1f", hrData.AbsenceRate),
					"sickRate":              fmt.Sprintf("%.1f", hrData.SickRate),

					// Review-Statistiken
					"upcomingReviews": hrData.UpcomingReviews,
					"overdueReviews":  hrData.OverdueReviews,

					// Fluktuationsstatistiken
					"turnoverRate":      fmt.Sprintf("%.1f", hrData.TurnoverRate),
					"averageEmployment": fmt.Sprintf("%.1f", hrData.AverageEmployment),
					"retentionRate":     fmt.Sprintf("%.1f", hrData.RetentionRate),

					// Chart-Daten
					"departmentLabels":      departmentLabels,
					"departmentData":        departmentData,
					"statusLabels":          statusLabels,
					"statusData":            statusData,
					"ageLabels":             ageLabels,
					"ageData":               ageData,
					"tenureLabels":          tenureLabels,
					"tenureData":            tenureData,
					"monthlyHiresData":      hrData.MonthlyHires,
					"monthlyDeparturesData": hrData.MonthlyDepartures,
					"absenceByMonthData":    hrData.AbsenceByMonth,
					"sicknessByMonthData":   hrData.SicknessByMonth,

					// Mitarbeiterübersicht
					"recentEmployees":  recentEmployees,
					"recentActivities": recentActivities,
				}

				// Daten zusammenführen
				for k, v := range hrDataForTemplate {
					commonData[k] = v
				}

				c.HTML(http.StatusOK, "dashboard.html", commonData)
				return
			}

			// Hier folgt der bestehende Code für Admin/Manager/User Rollen
			// Service für Kostenberechnungen initialisieren (nur für Admin/Manager)
			costService := service.NewCostService()

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
						profileImg = "" // Leer lassen, damit Platzhalter mit Initialen angezeigt wird
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
						"ID":       "",
						"Name":     "Keine Mitarbeiter",
						"Position": "",
						"Status":   "",
					},
				}
			}

			// Anzahl abgelaufener Dokumente (in einer echten Anwendung würden wir dies berechnen)
			expiredDocuments := 2

			// Formatieren der monatlichen Personalkosten
			formattedLaborCosts := fmt.Sprintf("%.2f", monthlyLaborCosts)

			// Vollständige Daten für Admin/Manager
			for k, v := range commonData {
				commonData[k] = v
			}

			adminManagerData := gin.H{
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
			}

			// Daten zusammenführen
			for k, v := range adminManagerData {
				commonData[k] = v
			}

			c.HTML(http.StatusOK, "dashboard.html", commonData)
		})

		// Kalender-Handler und andere Handler
		calendarHandler := handler.NewCalendarHandler()
		planningHandler := handler.NewPlanningHandler()
		timeTrackingHandler := handler.NewTimeTrackingHandler()
		statisticsHandler := handler.NewStatisticsHandler()
		statisticsAPIHandler := handler.NewStatisticsAPIHandler()
		overtimeHandler := handler.NewOvertimeHandler()

		// Hauptrouten
		authorized.GET("/absence", calendarHandler.GetAbsenceCalendar)
		authorized.GET("/planning", planningHandler.GetProjectPlanningView)
		authorized.GET("/timetracking", timeTrackingHandler.GetTimeTrackingView)
		authorized.GET("/api/timetracking/employee/:id", timeTrackingHandler.GetEmployeeTimeEntries)
		authorized.GET("/timetracking/export", timeTrackingHandler.ExportTimeTracking)
		authorized.GET("/statistics", statisticsHandler.GetStatisticsView)
		authorized.POST("/api/statistics/filter", statisticsAPIHandler.GetFilteredStatistics)
		authorized.POST("/api/statistics/extended", statisticsAPIHandler.GetExtendedStatistics)

		// Benutzerprofilrouten
		authorized.GET("/profile", userHandler.ShowUserProfile)

		// Einstellungsrouten (für alle Benutzer)
		authorized.GET("/settings", userHandler.ShowSettings)

		// System-Einstellungen Routen (nur für Admins)
		authorized.POST("/api/settings/company-name", middleware.RoleMiddleware(model.RoleAdmin), systemSettingsHandler.UpdateCompanyName)
		authorized.POST("/api/settings/language", middleware.RoleMiddleware(model.RoleAdmin), systemSettingsHandler.UpdateLanguage)
		authorized.POST("/api/settings/state", middleware.RoleMiddleware(model.RoleAdmin), systemSettingsHandler.UpdateState)
		authorized.GET("/api/settings", systemSettingsHandler.GetSystemSettings)
		authorized.POST("/api/settings", middleware.RoleMiddleware(model.RoleAdmin), systemSettingsHandler.UpdateSystemSettings)

		// Feiertags-API Routen
		authorized.GET("/api/holidays", holidayHandler.GetHolidays)
		authorized.GET("/api/holidays/check", holidayHandler.CheckHoliday)
		authorized.GET("/api/holidays/working-days", holidayHandler.GetWorkingDays)
		authorized.GET("/api/holidays/current-year", holidayHandler.GetCurrentYearHolidays)

		// Benutzerverwaltungsrouten (mit rollenbasierter Zugriffssteuerung)
		authorized.GET("/users", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), userHandler.ListUsers)
		authorized.GET("/users/add", middleware.RoleMiddleware(model.RoleAdmin), userHandler.ShowAddUserForm)
		authorized.POST("/users/add", middleware.RoleMiddleware(model.RoleAdmin), userHandler.AddUser)
		authorized.GET("/users/edit/:id", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), middleware.HRMiddleware(), userHandler.ShowEditUserForm)
		authorized.POST("/users/edit/:id", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), middleware.HRMiddleware(), userHandler.UpdateUser)
		authorized.DELETE("/users/delete/:id", middleware.RoleMiddleware(model.RoleAdmin), middleware.HRMiddleware(), userHandler.DeleteUser)

		// Passwortänderungsroute
		authorized.POST("/users/change-password", middleware.SelfOrAdminMiddleware(), userHandler.ChangePassword)

		// Mitarbeiter-Handler und Routen
		employeeHandler := handler.NewEmployeeHandler()
		documentHandler := handler.NewDocumentHandler()

		// Mitarbeiter-Routen
		authorized.GET("/employees", middleware.SalaryViewMiddleware(), employeeHandler.ListEmployees)
		authorized.GET("/employees/view/:id", middleware.SalaryViewMiddleware(), middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.GetEmployeeDetails)
		authorized.GET("/employees/edit/:id", middleware.SalaryViewMiddleware(), middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.ShowEditEmployeeForm)
		authorized.POST("/employees/add", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.AddEmployee)
		authorized.POST("/employees/edit/:id", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.UpdateEmployee)
		authorized.DELETE("/employees/delete/:id", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.DeleteEmployee)
		authorized.GET("/employees/:id/profile-image", employeeHandler.GetProfileImage)
		authorized.POST("/employees/:id/profile-image", employeeHandler.UploadProfileImage)

		// Überstunden Routen
		authorized.POST("/api/timetracking/recalculate-overtime", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), timeTrackingHandler.RecalculateOvertime)
		authorized.GET("/api/timetracking/employee/:id/overtime", timeTrackingHandler.GetEmployeeOvertimeDetails)
		authorized.POST("/api/timetracking/employee/:id/overtime", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), employeeHandler.RecalculateEmployeeOvertime)
		authorized.GET("/overtime", overtimeHandler.GetOvertimeView)
		authorized.POST("/api/overtime/recalculate", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), overtimeHandler.RecalculateAllOvertime)
		authorized.GET("/api/overtime/export", overtimeHandler.ExportOvertimeData)
		authorized.GET("/api/overtime/employee/:id", overtimeHandler.GetEmployeeOvertimeDetails)

		// Überstunden-Anpassungen Routen
		authorized.POST("/api/overtime/employee/:id/adjustment", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR), overtimeHandler.AddOvertimeAdjustment)
		authorized.GET("/api/overtime/employee/:id/adjustments", overtimeHandler.GetEmployeeAdjustments)
		authorized.POST("/api/overtime/adjustments/:adjustmentId/approve", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), overtimeHandler.ApproveAdjustment)
		authorized.GET("/api/overtime/adjustments/pending", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), overtimeHandler.GetPendingAdjustments)
		authorized.DELETE("/api/overtime/adjustments/:adjustmentId", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), overtimeHandler.DeleteAdjustment)

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

		// Conversation-Routen
		authorized.POST("/employees/:id/conversations", documentHandler.AddConversation)
		authorized.DELETE("/employees/:id/conversations/:conversationId", documentHandler.DeleteConversation)
		authorized.POST("/employees/:id/conversations/:conversationId/complete", documentHandler.CompleteConversation)
		authorized.PUT("/employees/:id/conversations/:conversationId", documentHandler.UpdateConversation)
		authorized.GET("/upcoming-conversations", employeeHandler.ListUpcomingConversations)

		// Integrations-Handler
		integrationHandler := handler.NewIntegrationHandler()

		// API-Endpunkte für Integrationen
		authorized.POST("/api/integrations/timebutler/save", middleware.RoleMiddleware(model.RoleAdmin), integrationHandler.SaveTimebutlerApiKey)
		authorized.GET("/api/integrations/status", integrationHandler.GetIntegrationStatus)
		authorized.GET("/api/integrations/timebutler/test", integrationHandler.TestTimebutlerConnection)
		authorized.POST("/api/integrations/timebutler/sync/users", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncTimebutlerUsers)
		authorized.POST("/api/integrations/timebutler/sync/absences", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncTimebutlerAbsences)
		authorized.POST("/api/integrations/timebutler/sync/holidayentitlements", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncTimebutlerHolidayEntitlements)

		// API-Endpunkte für 123Erfasst
		authorized.POST("/api/integrations/123erfasst/save", middleware.RoleMiddleware(model.RoleAdmin), integrationHandler.SaveErfasst123Credentials)
		authorized.GET("/api/integrations/123erfasst/test", integrationHandler.TestErfasst123Connection)
		authorized.POST("/api/integrations/123erfasst/sync/projects", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncErfasst123Projects)
		authorized.POST("/api/integrations/123erfasst/remove", middleware.RoleMiddleware(model.RoleAdmin), integrationHandler.RemoveErfasst123Integration)
		authorized.POST("/api/integrations/123erfasst/sync/times", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncErfasst123TimeEntries)
		authorized.GET("/api/integrations/123erfasst/sync-status", integrationHandler.GetErfasst123SyncStatus)
		authorized.POST("/api/integrations/123erfasst/set-auto-sync", middleware.RoleMiddleware(model.RoleAdmin), integrationHandler.SetErfasst123AutoSync)
		authorized.POST("/api/integrations/123erfasst/set-sync-start-date", middleware.RoleMiddleware(model.RoleAdmin), integrationHandler.SetErfasst123SyncStartDate)
		authorized.POST("/api/integrations/123erfasst/full-sync", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.TriggerErfasst123FullSync)
		authorized.POST("/api/integrations/123erfasst/sync/employees", middleware.RoleMiddleware(model.RoleAdmin, model.RoleHR), integrationHandler.SyncErfasst123Employees)

		// Optionale API-Endpoints für AJAX-Anfragen
		api := router.Group("/api")
		api.Use(middleware.AuthMiddleware())
		{

			api.DELETE("/employees/:id", employeeHandler.DeleteEmployee)
			api.GET("/employees/:id/name", handler.GetEmployeeName)
		}
	}
}
