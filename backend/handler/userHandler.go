// backend/handler/userHandler.go
package handler

import (
	"net/http"
	"time"

	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"

	"github.com/gin-gonic/gin"
)

// UserHandler verwaltet alle Anfragen zu Benutzern
type UserHandler struct {
	userRepo *repository.UserRepository
}

// NewUserHandler erstellt einen neuen UserHandler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userRepo: repository.NewUserRepository(),
	}
}

// ListUsers zeigt die Liste aller Benutzer an (nur für Admins)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Alle Benutzer abrufen
	users, err := h.userRepo.FindAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Abrufen der Benutzer: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Daten an das Template übergeben
	c.HTML(http.StatusOK, "users.html", gin.H{
		"title":      "Benutzer",
		"active":     "users",
		"user":       userModel.FirstName + " " + userModel.LastName,
		"email":      userModel.Email,
		"year":       time.Now().Year(),
		"users":      users,
		"totalUsers": len(users),
		"userRole":   c.GetString("userRole"),
	})
}

// ShowAddUserForm zeigt das Formular zum Hinzufügen eines Benutzers an
func (h *UserHandler) ShowAddUserForm(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	c.HTML(http.StatusOK, "user_add.html", gin.H{
		"title":    "Benutzer hinzufügen",
		"active":   "users",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"userRole": c.GetString("userRole"),
	})
}

// AddUser fügt einen neuen Benutzer hinzu
func (h *UserHandler) AddUser(c *gin.Context) {
	// Formulardaten abrufen
	firstName := c.PostForm("firstName")
	lastName := c.PostForm("lastName")
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")

	// Neuen Benutzer erstellen
	newUser := &model.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  password,
		Role:      model.UserRole(role),
		Status:    model.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Benutzer in der Datenbank speichern
	err := h.userRepo.Create(newUser)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Erstellen des Benutzers: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserAdded,
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		newUser.ID,
		"user",
		newUser.FirstName+" "+newUser.LastName,
		"Neuer Benutzer hinzugefügt",
	)

	// Hier ändert sich die Umleitung - zur Einstellungsseite statt zur Benutzerliste
	c.Redirect(http.StatusFound, "/settings?success=added")
}

// ShowEditUserForm zeigt das Formular zum Bearbeiten eines Benutzers an
func (h *UserHandler) ShowEditUserForm(c *gin.Context) {
	id := c.Param("id")

	// Benutzer anhand der ID abrufen
	userToEdit, err := h.userRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	c.HTML(http.StatusOK, "user_edit.html", gin.H{
		"title":    "Benutzer bearbeiten",
		"active":   "users",
		"user":     currentUserModel.FirstName + " " + currentUserModel.LastName,
		"email":    currentUserModel.Email,
		"year":     time.Now().Year(),
		"editUser": userToEdit,
		"userRole": c.GetString("userRole"),
	})
}

// UpdateUser aktualisiert einen bestehenden Benutzer
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	// Benutzer anhand der ID abrufen
	userToUpdate, err := h.userRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Formulardaten abrufen und Benutzer aktualisieren
	userToUpdate.FirstName = c.PostForm("firstName")
	userToUpdate.LastName = c.PostForm("lastName")
	userToUpdate.Email = c.PostForm("email")

	// Passwort nur aktualisieren, wenn ein neues angegeben wurde
	newPassword := c.PostForm("password")
	if newPassword != "" {
		userToUpdate.Password = newPassword
		userToUpdate.HashPassword()
	}

	// Rolle nur aktualisieren, wenn der aktuelle Benutzer ein Admin ist
	currentUserRole, _ := c.Get("userRole")
	if currentUserRole == string(model.RoleAdmin) {
		userToUpdate.Role = model.UserRole(c.PostForm("role"))
		userToUpdate.Status = model.UserStatus(c.PostForm("status"))
	}

	userToUpdate.UpdatedAt = time.Now()

	// Benutzer in der Datenbank aktualisieren
	err = h.userRepo.Update(userToUpdate)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Benutzers: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserUpdated,
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		userToUpdate.ID,
		"user",
		userToUpdate.FirstName+" "+userToUpdate.LastName,
		"Benutzer aktualisiert",
	)

	// Zurück zur Benutzerliste oder zum Profil
	if currentUserRole == string(model.RoleAdmin) {
		c.Redirect(http.StatusFound, "/settings?success=updated")
	} else {
		c.Redirect(http.StatusFound, "/profile?success=updated")
	}
}

// DeleteUser löscht einen Benutzer
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Benutzer anhand der ID abrufen (für den Aktivitätslog)
	userToDelete, err := h.userRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Benutzer nicht gefunden"})
		return
	}

	// Benutzer löschen
	err = h.userRepo.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen des Benutzers: " + err.Error()})
		return
	}

	// Aktivität loggen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	activityRepo := repository.NewActivityRepository()
	_, _ = activityRepo.LogActivity(
		model.ActivityTypeUserDeleted,
		currentUserModel.ID,
		currentUserModel.FirstName+" "+currentUserModel.LastName,
		userToDelete.ID,
		"user",
		userToDelete.FirstName+" "+userToDelete.LastName,
		"Benutzer gelöscht",
	)

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{"message": "Benutzer erfolgreich gelöscht"})
}

// ShowUserProfile zeigt das Profil des aktuellen Benutzers an
func (h *UserHandler) ShowUserProfile(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	currentUser, _ := c.Get("user")
	currentUserModel := currentUser.(*model.User)

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":    "Mein Profil",
		"active":   "profile",
		"user":     currentUserModel.FirstName + " " + currentUserModel.LastName,
		"email":    currentUserModel.Email,
		"year":     time.Now().Year(),
		"userRole": c.GetString("userRole"),
		"profile":  currentUserModel,
	})
}

// ChangePassword ändert das Passwort eines Benutzers
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// ID aus dem Formular abrufen
	id := c.PostForm("id")

	// Benutzer anhand der ID abrufen
	user, err := h.userRepo.FindByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Benutzer nicht gefunden",
			"year":    time.Now().Year(),
		})
		return
	}

	// Passwörter aus dem Formular abrufen
	currentPassword := c.PostForm("currentPassword")
	newPassword := c.PostForm("newPassword")
	confirmPassword := c.PostForm("confirmPassword")

	// Überprüfen, ob das aktuelle Passwort korrekt ist
	if !user.CheckPassword(currentPassword) {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Das aktuelle Passwort ist nicht korrekt",
			"year":    time.Now().Year(),
		})
		return
	}

	// Überprüfen, ob die neuen Passwörter übereinstimmen
	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Die neuen Passwörter stimmen nicht überein",
			"year":    time.Now().Year(),
		})
		return
	}

	// Überprüfen, ob das neue Passwort eine Mindestlänge hat
	if len(newPassword) < 6 {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Das neue Passwort muss mindestens 6 Zeichen lang sein",
			"year":    time.Now().Year(),
		})
		return
	}

	// Passwort aktualisieren
	user.Password = newPassword
	err = user.HashPassword()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Hashen des Passworts: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Benutzer in der Datenbank aktualisieren
	user.UpdatedAt = time.Now()
	err = h.userRepo.Update(user)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Fehler",
			"message": "Fehler beim Aktualisieren des Benutzers: " + err.Error(),
			"year":    time.Now().Year(),
		})
		return
	}

	// Zurück zum Profil mit Erfolgsmeldung
	c.Redirect(http.StatusFound, "/profile?success=password_changed")
}

// ShowSettings zeigt die Einstellungsseite an
func (h *UserHandler) ShowSettings(c *gin.Context) {
	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)
	userRole := c.GetString("userRole")

	// Erfolgsparameter aus der URL extrahieren
	success := c.Query("success")

	// Vorbereitete Daten für die Einstellungsseite
	data := gin.H{
		"title":    "Einstellungen",
		"active":   "settings",
		"user":     userModel.FirstName + " " + userModel.LastName,
		"email":    userModel.Email,
		"year":     time.Now().Year(),
		"userRole": userRole,
	}

	// Erfolgsparameter hinzufügen, wenn vorhanden
	if success != "" {
		data["success"] = success
	}

	// Wenn der Benutzer ein Administrator ist, fügen wir Benutzerdaten hinzu
	if userRole == string(model.RoleAdmin) {
		users, err := h.userRepo.FindAll()
		if err != nil {
			users = []*model.User{} // Leere Liste im Fehlerfall
		}
		data["users"] = users
		data["totalUsers"] = len(users)
	}

	c.HTML(http.StatusOK, "settings.html", data)
}
