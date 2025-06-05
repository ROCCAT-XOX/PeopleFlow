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
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

// Hilfsfunktion für Überstunden-Status
func getOvertimeStatus(balance float64) string {
	if balance > 20 {
		return "high"
	} else if balance > 0 {
		return "positive"
	} else if balance < -10 {
		return "critical"
	} else if balance < 0 {
		return "negative"
	}
	return "neutral"
}

// Hilfsfunktion für Aktivitäten-Formatierung
func formatActivities(activities []*model.Activity) []gin.H {
	var formatted []gin.H
	for i, activity := range activities {
		isLast := i == len(activities)-1

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

		formatted = append(formatted, gin.H{
			"IconBgClass": activity.GetIconClass(),
			"IconSVG":     activity.GetIconSVG(),
			"Message":     message,
			"Time":        activity.FormatTimeAgo(),
			"IsLast":      isLast,
		})
	}
	return formatted
}

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
		// Handler erstellen
		userHandler := handler.NewUserHandler()
		systemSettingsHandler := handler.NewSystemSettingsHandler()
		holidayHandler := handler.NewHolidayHandler()

		// Root-Pfad zum Dashboard umleiten
		router.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/dashboard")
		})

		// Dashboard Route
		authorized.GET("/dashboard", func(c *gin.Context) {
			user, _ := c.Get("user")
			userModel := user.(*model.User)
			userRole, _ := c.Get("userRole")

			// Repository für Daten
			employeeRepo := repository.NewEmployeeRepository()
			activityRepo := repository.NewActivityRepository()
			overtimeAdjustmentRepo := repository.NewOvertimeAdjustmentRepository()

			// Gemeinsame Daten für alle Rollen
			currentDate := time.Now().Format("Monday, 02. January 2006")
			commonData := gin.H{
				"title":       "Dashboard",
				"active":      "dashboard",
				"user":        userModel.FirstName + " " + userModel.LastName,
				"email":       userModel.Email,
				"userRole":    userRole,
				"year":        time.Now().Year(),
				"currentDate": currentDate,
			}

			// User-spezifisches Dashboard
			if userRole == string(model.RoleUser) {
				// Finde den Mitarbeiter-Datensatz des Users
				employee, err := employeeRepo.FindByEmail(userModel.Email)
				if err != nil {
					// Wenn kein Mitarbeiter-Datensatz gefunden wird
					c.HTML(http.StatusOK, "dashboard.html", commonData)
					return
				}

				// Lade Überstunden-Anpassungen
				adjustments, _ := overtimeAdjustmentRepo.FindByEmployeeID(employee.ID.Hex())

				// Berechne finale Überstunden
				totalAdjustments := 0.0
				for _, adj := range adjustments {
					if adj.Status == "approved" {
						totalAdjustments += adj.Hours
					}
				}
				finalOvertimeBalance := employee.OvertimeBalance + totalAdjustments

				// Berechne Urlaubstage
				currentYear := time.Now().Year()
				var usedVacationDays float64 = 0
				for _, absence := range employee.Absences {
					if absence.Type == "vacation" &&
						absence.Status == "approved" &&
						absence.StartDate.Year() == currentYear {
						usedVacationDays += absence.Days
					}
				}

				// If VacationDays is not set, provide a default
				if employee.VacationDays == 0 {
					employee.VacationDays = 30 // Default value if not set
				}

				// Calculate remaining vacation days if not already set
				if employee.RemainingVacation == 0 {
					employee.RemainingVacation = employee.VacationDays - int(usedVacationDays)
				}

				// User-spezifische Daten
				userData := gin.H{
					"employee":          employee,
					"overtimeBalance":   finalOvertimeBalance,
					"vacationDays":      employee.VacationDays,
					"usedVacationDays":  usedVacationDays,
					"remainingVacation": employee.VacationDays - int(usedVacationDays),
					"pendingAbsences":   0,
				}

				// Zähle ausstehende Abwesenheiten
				for _, absence := range employee.Absences {
					if absence.Status == "requested" {
						userData["pendingAbsences"] = userData["pendingAbsences"].(int) + 1
					}
				}

				// Daten zusammenführen
				for k, v := range userData {
					commonData[k] = v
				}

				c.HTML(http.StatusOK, "dashboard.html", commonData)
				return
			}

			// Alle Mitarbeiter abrufen für Admin/Manager/HR
			allEmployees, err := employeeRepo.FindAll()
			if err != nil {
				allEmployees = []*model.Employee{}
			}

			totalEmployees := len(allEmployees)
			commonData["totalEmployees"] = totalEmployees

			// HR-spezifisches Dashboard
			if userRole == string(model.RoleHR) {
				// HR-Service für demografische Daten
				hrService := service.NewHRService()
				hrData := hrService.CalculateHRDashboardData(allEmployees)

				// Aktivitäten
				recentActivitiesData, _ := activityRepo.FindRecent(10)

				// Abwesenheitsanträge sammeln
				pendingAbsences := []gin.H{}
				for _, emp := range allEmployees {
					for _, absence := range emp.Absences {
						if absence.Status == "requested" {
							pendingAbsences = append(pendingAbsences, gin.H{
								"EmployeeID":   emp.ID.Hex(),
								"EmployeeName": emp.FirstName + " " + emp.LastName,
								"Type":         absence.Type,
								"StartDate":    absence.StartDate.Format("02.01.2006"),
								"EndDate":      absence.EndDate.Format("02.01.2006"),
								"Days":         absence.Days,
								"Reason":       absence.Reason,
								"Department":   emp.Department,
							})
						}
					}
				}

				// Ausstehende Überstunden-Anpassungen für HR
				pendingOvertimeAdjustments := []gin.H{}
				if userRole == string(model.RoleAdmin) || userRole == string(model.RoleManager) || userRole == string(model.RoleHR) {
					adjustments, err := overtimeAdjustmentRepo.FindPending()
					if err == nil {
						for _, adj := range adjustments {
							// Mitarbeiter-Namen abrufen
							employee, err := employeeRepo.FindByID(adj.EmployeeID.Hex())
							if err == nil {
								pendingOvertimeAdjustments = append(pendingOvertimeAdjustments, gin.H{
									"ID":           adj.ID.Hex(),
									"EmployeeID":   adj.EmployeeID.Hex(),
									"EmployeeName": employee.FirstName + " " + employee.LastName,
									"Department":   employee.Department,
									"Type":         adj.GetTypeDisplayName(),
									"Hours":        adj.FormatHours(),
									"Reason":       adj.Reason,
									"Description":  adj.Description,
									"AdjusterName": adj.AdjusterName,
									"CreatedAt":    adj.CreatedAt.Format("02.01.2006"),
								})
							}
						}
					}
				}

				// NEU: Anstehende Weiterbildungen und Gespräche sammeln
				upcomingTrainings := []gin.H{}
				upcomingConversations := []gin.H{}
				now := time.Now()

				for _, emp := range allEmployees {
					// Weiterbildungen
					for _, training := range emp.Trainings {
						if (training.Status == "planned" || training.Status == "ongoing") &&
							training.StartDate.After(now.AddDate(0, 0, -7)) {
							upcomingTrainings = append(upcomingTrainings, gin.H{
								"EmployeeID":   emp.ID.Hex(),
								"EmployeeName": emp.FirstName + " " + emp.LastName,
								"Department":   string(emp.Department),
								"Title":        training.Title,
								"Provider":     training.Provider,
								"StartDate":    training.StartDate.Format("02.01.2006"),
								"EndDate":      training.EndDate.Format("02.01.2006"),
								"Status":       training.Status,
								"Description":  training.Description,
							})
						}
					}

					// Gespräche
					for _, conv := range emp.Conversations {
						if conv.Status == "planned" &&
							conv.Date.After(now) &&
							conv.Date.Before(now.AddDate(0, 0, 30)) {
							upcomingConversations = append(upcomingConversations, gin.H{
								"EmployeeID":   emp.ID.Hex(),
								"EmployeeName": emp.FirstName + " " + emp.LastName,
								"Department":   string(emp.Department),
								"Title":        conv.Title,
								"Date":         conv.Date.Format("02.01.2006"),
								"Description":  conv.Description,
								"DaysUntil":    int(conv.Date.Sub(now).Hours() / 24),
							})
						}
					}
				}

				// Sortierung
				sort.Slice(upcomingTrainings, func(i, j int) bool {
					date1, _ := time.Parse("02.01.2006", upcomingTrainings[i]["StartDate"].(string))
					date2, _ := time.Parse("02.01.2006", upcomingTrainings[j]["StartDate"].(string))
					return date1.Before(date2)
				})

				sort.Slice(upcomingConversations, func(i, j int) bool {
					date1, _ := time.Parse("02.01.2006", upcomingConversations[i]["Date"].(string))
					date2, _ := time.Parse("02.01.2006", upcomingConversations[j]["Date"].(string))
					return date1.Before(date2)
				})

				// Chart-Daten
				departmentLabels, departmentData := hrService.GetDepartmentLabelsAndData(hrData.DepartmentCounts)
				statusLabels, statusData := hrService.GetStatusLabelsAndData(hrData.StatusDistribution)
				ageLabels, ageData := hrService.GetAgeLabelsAndData(hrData.AgeDistribution)

				hrDashboardData := gin.H{
					"activeEmployees":            hrData.ActiveEmployees,
					"onLeaveEmployees":           hrData.OnLeaveEmployees,
					"sickEmployees":              hrData.SickEmployees,
					"inactiveEmployees":          hrData.InactiveEmployees,
					"currentAbsences":            hrData.CurrentAbsences,
					"pendingAbsences":            pendingAbsences,
					"pendingAbsencesCount":       len(pendingAbsences),
					"pendingOvertimeAdjustments": pendingOvertimeAdjustments,
					"pendingOvertimeCount":       len(pendingOvertimeAdjustments),
					"upcomingTrainings":          upcomingTrainings,
					"upcomingTrainingsCount":     len(upcomingTrainings),
					"upcomingConversations":      upcomingConversations,
					"upcomingConversationsCount": len(upcomingConversations),
					"departmentLabels":           departmentLabels,
					"departmentData":             departmentData,
					"statusLabels":               statusLabels,
					"statusData":                 statusData,
					"ageLabels":                  ageLabels,
					"ageData":                    ageData,
					"recentActivities":           formatActivities(recentActivitiesData),
					"absenceByMonthData":         hrData.AbsenceByMonth,
					"sicknessByMonthData":        hrData.SicknessByMonth,
				}

				for k, v := range hrDashboardData {
					commonData[k] = v
				}

				c.HTML(http.StatusOK, "dashboard.html", commonData)
				return
			}

			// Admin/Manager Dashboard
			// Services für Berechnungen
			costService := service.NewCostService()

			// Personalkosten
			monthlyLaborCosts := costService.CalculateMonthlyLaborCosts(allEmployees)
			monthlyCostsData := costService.GenerateMonthlyLaborCostsTrend(monthlyLaborCosts)

			// Überstunden-Daten sammeln
			var totalOvertime float64 = 0
			var overtimeEmployees []gin.H
			for _, emp := range allEmployees {
				if len(emp.TimeEntries) > 0 {
					// Anpassungen laden
					adjustments, _ := overtimeAdjustmentRepo.FindByEmployeeID(emp.ID.Hex())
					totalAdjustments := 0.0
					for _, adj := range adjustments {
						if adj.Status == "approved" {
							totalAdjustments += adj.Hours
						}
					}

					finalBalance := emp.OvertimeBalance + totalAdjustments
					totalOvertime += finalBalance

					if finalBalance != 0 {
						overtimeEmployees = append(overtimeEmployees, gin.H{
							"Name":    emp.FirstName + " " + emp.LastName,
							"Balance": finalBalance,
							"Status":  getOvertimeStatus(finalBalance),
						})
					}
				}
			}

			// Sortiere nach Überstunden (höchste zuerst)
			sort.Slice(overtimeEmployees, func(i, j int) bool {
				return overtimeEmployees[i]["Balance"].(float64) > overtimeEmployees[j]["Balance"].(float64)
			})

			// Limitiere auf Top 10
			if len(overtimeEmployees) > 10 {
				overtimeEmployees = overtimeEmployees[:10]
			}

			// Abwesenheitsanträge
			pendingAbsences := []gin.H{}
			approvedAbsencesToday := 0
			sickToday := 0

			today := time.Now()
			for _, emp := range allEmployees {
				for _, absence := range emp.Absences {
					// Ausstehende Anträge
					if absence.Status == "requested" {
						pendingAbsences = append(pendingAbsences, gin.H{
							"EmployeeID":   emp.ID.Hex(),
							"EmployeeName": emp.FirstName + " " + emp.LastName,
							"Type":         absence.Type,
							"StartDate":    absence.StartDate.Format("02.01.2006"),
							"EndDate":      absence.EndDate.Format("02.01.2006"),
							"Days":         absence.Days,
							"Reason":       absence.Reason,
							"Department":   emp.Department,
						})
					}

					// Aktuelle Abwesenheiten
					if absence.Status == "approved" &&
						!absence.StartDate.After(today) &&
						!absence.EndDate.Before(today) {
						if absence.Type == "sick" {
							sickToday++
						} else {
							approvedAbsencesToday++
						}
					}
				}
			}

			// Ausstehende Überstunden-Anpassungen für Admin/Manager
			pendingOvertimeAdjustments := []gin.H{}
			if userRole == string(model.RoleAdmin) || userRole == string(model.RoleManager) {
				adjustments, err := overtimeAdjustmentRepo.FindPending()
				if err == nil {
					for _, adj := range adjustments {
						// Mitarbeiter-Namen abrufen
						employee, err := employeeRepo.FindByID(adj.EmployeeID.Hex())
						if err == nil {
							pendingOvertimeAdjustments = append(pendingOvertimeAdjustments, gin.H{
								"ID":           adj.ID.Hex(),
								"EmployeeID":   adj.EmployeeID.Hex(),
								"EmployeeName": employee.FirstName + " " + employee.LastName,
								"Department":   employee.Department,
								"Type":         adj.GetTypeDisplayName(),
								"Hours":        adj.FormatHours(),
								"Reason":       adj.Reason,
								"Description":  adj.Description,
								"AdjusterName": adj.AdjusterName,
								"CreatedAt":    adj.CreatedAt.Format("02.01.2006"),
							})
						}
					}
				}
			}

			// NEU: Anstehende Weiterbildungen und Gespräche für Admin/Manager
			upcomingTrainings := []gin.H{}
			upcomingConversations := []gin.H{}
			now := time.Now()

			for _, emp := range allEmployees {
				// Weiterbildungen
				for _, training := range emp.Trainings {
					if (training.Status == "planned" || training.Status == "ongoing") &&
						training.StartDate.After(now.AddDate(0, 0, -7)) {
						upcomingTrainings = append(upcomingTrainings, gin.H{
							"EmployeeID":   emp.ID.Hex(),
							"EmployeeName": emp.FirstName + " " + emp.LastName,
							"Department":   string(emp.Department),
							"Title":        training.Title,
							"Provider":     training.Provider,
							"StartDate":    training.StartDate.Format("02.01.2006"),
							"EndDate":      training.EndDate.Format("02.01.2006"),
							"Status":       training.Status,
							"Description":  training.Description,
						})
					}
				}

				// Gespräche
				for _, conv := range emp.Conversations {
					if conv.Status == "planned" &&
						conv.Date.After(now) &&
						conv.Date.Before(now.AddDate(0, 0, 30)) {
						upcomingConversations = append(upcomingConversations, gin.H{
							"EmployeeID":   emp.ID.Hex(),
							"EmployeeName": emp.FirstName + " " + emp.LastName,
							"Department":   string(emp.Department),
							"Title":        conv.Title,
							"Date":         conv.Date.Format("02.01.2006"),
							"Description":  conv.Description,
							"DaysUntil":    int(conv.Date.Sub(now).Hours() / 24),
						})
					}
				}
			}

			// Sortierung
			sort.Slice(upcomingTrainings, func(i, j int) bool {
				date1, _ := time.Parse("02.01.2006", upcomingTrainings[i]["StartDate"].(string))
				date2, _ := time.Parse("02.01.2006", upcomingTrainings[j]["StartDate"].(string))
				return date1.Before(date2)
			})

			sort.Slice(upcomingConversations, func(i, j int) bool {
				date1, _ := time.Parse("02.01.2006", upcomingConversations[i]["Date"].(string))
				date2, _ := time.Parse("02.01.2006", upcomingConversations[j]["Date"].(string))
				return date1.Before(date2)
			})

			// Sortiere Abwesenheitsanträge nach Datum
			sort.Slice(pendingAbsences, func(i, j int) bool {
				date1, _ := time.Parse("02.01.2006", pendingAbsences[i]["StartDate"].(string))
				date2, _ := time.Parse("02.01.2006", pendingAbsences[j]["StartDate"].(string))
				return date1.Before(date2)
			})

			// Krankheitsverlauf (letzte 12 Monate)
			sicknessTrend := make([]int, 12)
			for i := 0; i < 12; i++ {
				month := now.AddDate(0, -i, 0)
				count := 0
				for _, emp := range allEmployees {
					for _, absence := range emp.Absences {
						if absence.Type == "sick" &&
							absence.Status == "approved" &&
							absence.StartDate.Year() == month.Year() &&
							absence.StartDate.Month() == month.Month() {
							count++
						}
					}
				}
				sicknessTrend[11-i] = count
			}

			// Überstunden-Verlauf (letzte 12 Monate)
			overtimeTrend := make([]float64, 12)
			monthLabels := make([]string, 12)
			for i := 0; i < 12; i++ {
				month := now.AddDate(0, -i, 0)
				monthLabels[11-i] = month.Format("Jan")

				// Hier würden wir normalerweise historische Daten abrufen
				// Für Demo-Zwecke generieren wir Beispieldaten
				overtimeTrend[11-i] = totalOvertime * (0.8 + float64(i)*0.02)
			}

			// Aktivitäten
			recentActivitiesData, _ := activityRepo.FindRecent(5)

			// Abteilungsverteilung
			departmentCounts := make(map[string]int)
			for _, emp := range allEmployees {
				if emp.Status == model.EmployeeStatusActive {
					dept := string(emp.Department)
					if dept == "" {
						dept = "Nicht zugewiesen"
					}
					departmentCounts[dept]++
				}
			}

			var departmentLabels []string
			var departmentData []int
			for dept, count := range departmentCounts {
				departmentLabels = append(departmentLabels, dept)
				departmentData = append(departmentData, count)
			}

			// Admin/Manager spezifische Daten
			adminData := gin.H{
				"monthlyLaborCosts":          fmt.Sprintf("%.2f", monthlyLaborCosts),
				"monthlyCostsData":           monthlyCostsData,
				"totalOvertime":              totalOvertime,
				"overtimeEmployees":          overtimeEmployees,
				"pendingAbsences":            pendingAbsences,
				"pendingAbsencesCount":       len(pendingAbsences),
				"pendingOvertimeAdjustments": pendingOvertimeAdjustments,
				"pendingOvertimeCount":       len(pendingOvertimeAdjustments),
				"upcomingTrainings":          upcomingTrainings,
				"upcomingTrainingsCount":     len(upcomingTrainings),
				"upcomingConversations":      upcomingConversations,
				"upcomingConversationsCount": len(upcomingConversations),
				"approvedAbsencesToday":      approvedAbsencesToday,
				"sickToday":                  sickToday,
				"sicknessTrend":              sicknessTrend,
				"overtimeTrend":              overtimeTrend,
				"monthLabels":                monthLabels,
				"recentActivities":           formatActivities(recentActivitiesData),
				"departmentLabels":           departmentLabels,
				"departmentData":             departmentData,
			}

			// Daten zusammenführen
			for k, v := range adminData {
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
		// Abwesenheitsübersicht Handler
		absenceOverviewHandler := handler.NewAbsenceOverviewHandler()

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

		// Abwesenheitsübersicht Route
		authorized.GET("/absence-overview", absenceOverviewHandler.GetAbsenceOverview)
		// API-Endpoints für Abwesenheitsanträge
		authorized.POST("/api/absence/request", absenceOverviewHandler.AddAbsenceRequest)
		authorized.POST("/api/absence/:employeeId/:absenceId/approve", middleware.RoleMiddleware(model.RoleAdmin, model.RoleManager), absenceOverviewHandler.ApproveAbsenceRequest)

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
