package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"PeoplePilot/backend"
	"PeoplePilot/backend/db"
	"PeoplePilot/backend/repository"
	"PeoplePilot/backend/utils"
)

func main() {
	// Set Gin to release mode in production
	// gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)

	// Datenbankverbindung herstellen
	if err := db.ConnectDB(); err != nil {
		log.Fatalf("Fehler beim Verbinden zur Datenbank: %v", err)
	}
	defer db.DisconnectDB()

	// Admin-Benutzer erstellen, falls keiner existiert
	userRepo := repository.NewUserRepository()
	if err := userRepo.CreateAdminUserIfNotExists(); err != nil {
		log.Printf("Warnung: Admin-Benutzer konnte nicht erstellt werden: %v", err)
	} else {
		log.Println("Admin-Benutzer wurde überprüft/erstellt")
	}

	// Upload-Verzeichnis erstellen, falls es nicht existiert
	if err := utils.EnsureUploadDirExists(); err != nil {
		log.Printf("Warnung: Upload-Verzeichnis konnte nicht erstellt werden: %v", err)
	}

	// Initialize router
	router := setupRouter()

	// Create and configure the server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start the server
	log.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter() *gin.Engine {
	// Create a default gin router with Logger and Recovery middleware
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files
	router.Static("/static", "./frontend/static")

	// Funktion für HTML-Template-Sicherheit hinzufügen
	router.SetFuncMap(utils.TemplateHelpers())

	// Load HTML templates - update to include subdirectories
	router.SetHTMLTemplate(loadTemplates())

	// Import routes from router.go
	backend.InitializeRoutes(router)

	return router
}

func loadTemplates() *template.Template {
	templ := template.New("").Funcs(utils.TemplateHelpers())
	templ = template.Must(templ.ParseGlob("frontend/templates/*.html"))
	templ = template.Must(templ.ParseGlob("frontend/templates/components/*.html"))

	return templ
}
