package handler

import (
	"PeoplePilot/backend/db"
	"PeoplePilot/backend/model"
	"PeoplePilot/backend/repository"
	"PeoplePilot/backend/service"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DocumentHandler verwaltet alle Anfragen zu Dokumenten
type DocumentHandler struct {
	employeeRepo *repository.EmployeeRepository
	fileService  *service.FileService
}

// NewDocumentHandler erstellt einen neuen DocumentHandler
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{
		employeeRepo: repository.NewEmployeeRepository(),
		fileService:  service.NewFileService(),
	}
}

// UploadDocument lädt ein Dokument für einen Mitarbeiter hoch
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Kategorie extrahieren (z.B. "application", "training", "evaluation", "absence", "general")
	category := c.DefaultPostForm("category", "general")

	// Dokumentenname und Beschreibung aus dem Formular extrahieren
	documentName := c.PostForm("name")
	description := c.PostForm("description")

	// Related ID, falls relevant (z.B. für Training, Evaluation, Absence)
	relatedID := c.PostForm("relatedId")

	// Benutzer aus dem Kontext abrufen
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Benutzer nicht im Kontext gefunden"})
		return
	}
	userModel := user.(*model.User)

	// Hochgeladene Datei abrufen
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Datei hochgeladen: " + err.Error()})
		return
	}

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Datei mit dem FileService hochladen
	document, err := h.fileService.UploadFile(file, documentName, description, category, userModel.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Hochladen der Datei: " + err.Error()})
		return
	}

	// Basierend auf der Kategorie das Dokument dem richtigen Bereich des Mitarbeiters hinzufügen
	switch category {
	case "application":
		// Zu den Bewerbungsunterlagen hinzufügen
		employee.ApplicationDocuments = append(employee.ApplicationDocuments, *document)
	case "training":
		// Zu einem Training hinzufügen, falls relatedID vorhanden ist
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, training := range employee.Trainings {
					if training.ID == relObjID {
						employee.Trainings[i].Documents = append(employee.Trainings[i].Documents, *document)
						break
					}
				}
			}
		}
	case "evaluation":
		// Zu einer Evaluation hinzufügen, falls relatedID vorhanden ist
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, eval := range employee.Evaluations {
					if eval.ID == relObjID {
						employee.Evaluations[i].Documents = append(employee.Evaluations[i].Documents, *document)
						break
					}
				}
			}
		}
	case "absence":
		// Zu einer Abwesenheit hinzufügen, falls relatedID vorhanden ist
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, absence := range employee.Absences {
					if absence.ID == relObjID {
						employee.Absences[i].Documents = append(employee.Absences[i].Documents, *document)
						break
					}
				}
			}
		}
	default:
		// Allgemeine Dokumente
		employee.Documents = append(employee.Documents, *document)
	}

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Dokument erfolgreich hochgeladen",
		"document": document,
	})
}

// DeleteDocument löscht ein Dokument eines Mitarbeiters
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	// Mitarbeiter-ID und Dokument-ID aus den URL-Parametern extrahieren
	employeeID := c.Param("id")
	documentID := c.Param("documentId")

	// Kategorie aus dem Query-Parameter extrahieren
	category := c.DefaultQuery("category", "general")

	// Related ID, falls relevant
	relatedID := c.Query("relatedId")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Dokument-ID in ObjectID umwandeln
	docObjID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Dokument-ID: " + err.Error()})
		return
	}

	// Dokument finden und löschen
	var documentToDelete *model.Document
	var documentIndex int = -1

	switch category {
	case "application":
		// In Bewerbungsunterlagen suchen
		for i, doc := range employee.ApplicationDocuments {
			if doc.ID == docObjID {
				documentToDelete = &employee.ApplicationDocuments[i]
				documentIndex = i
				break
			}
		}
		if documentIndex >= 0 {
			employee.ApplicationDocuments = append(employee.ApplicationDocuments[:documentIndex], employee.ApplicationDocuments[documentIndex+1:]...)
		}
	case "training":
		// In Trainings suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, training := range employee.Trainings {
					if training.ID == relObjID {
						for j, doc := range training.Documents {
							if doc.ID == docObjID {
								documentToDelete = &training.Documents[j]
								employee.Trainings[i].Documents = append(training.Documents[:j], training.Documents[j+1:]...)
								break
							}
						}
						break
					}
				}
			}
		}
	case "evaluation":
		// In Evaluations suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, eval := range employee.Evaluations {
					if eval.ID == relObjID {
						for j, doc := range eval.Documents {
							if doc.ID == docObjID {
								documentToDelete = &eval.Documents[j]
								employee.Evaluations[i].Documents = append(eval.Documents[:j], eval.Documents[j+1:]...)
								break
							}
						}
						break
					}
				}
			}
		}
	case "absence":
		// In Absences suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for i, absence := range employee.Absences {
					if absence.ID == relObjID {
						for j, doc := range absence.Documents {
							if doc.ID == docObjID {
								documentToDelete = &absence.Documents[j]
								employee.Absences[i].Documents = append(absence.Documents[:j], absence.Documents[j+1:]...)
								break
							}
						}
						break
					}
				}
			}
		}
	default:
		// In allgemeinen Dokumenten suchen
		for i, doc := range employee.Documents {
			if doc.ID == docObjID {
				documentToDelete = &employee.Documents[i]
				documentIndex = i
				break
			}
		}
		if documentIndex >= 0 {
			employee.Documents = append(employee.Documents[:documentIndex], employee.Documents[documentIndex+1:]...)
		}
	}

	if documentToDelete == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokument nicht gefunden"})
		return
	}

	// Datei aus dem Dateisystem löschen
	err = h.fileService.DeleteFile(documentToDelete.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Löschen der Datei: " + err.Error()})
		return
	}

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dokument erfolgreich gelöscht",
	})
}

// DownloadDocument lädt ein Dokument eines Mitarbeiters herunter
// DownloadDocument lädt ein Dokument eines Mitarbeiters herunter
func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	// Mitarbeiter-ID und Dokument-ID aus den URL-Parametern extrahieren
	employeeID := c.Param("id")
	documentID := c.Param("documentId")

	// Kategorie aus dem Query-Parameter extrahieren
	category := c.DefaultQuery("category", "general")

	// Related ID, falls relevant
	relatedID := c.Query("relatedId")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Dokument-ID in ObjectID umwandeln
	docObjID, err := primitive.ObjectIDFromHex(documentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Dokument-ID: " + err.Error()})
		return
	}

	// Dokument finden
	var documentToDownload *model.Document

	switch category {
	case "application":
		// In Bewerbungsunterlagen suchen
		for i, doc := range employee.ApplicationDocuments {
			if doc.ID == docObjID {
				documentToDownload = &employee.ApplicationDocuments[i]
				break
			}
		}
	case "training":
		// In Trainings suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for _, training := range employee.Trainings {
					if training.ID == relObjID {
						for i, doc := range training.Documents {
							if doc.ID == docObjID {
								documentToDownload = &training.Documents[i]
								break
							}
						}
						break
					}
				}
			}
		}
	case "evaluation":
		// In Evaluations suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for _, eval := range employee.Evaluations {
					if eval.ID == relObjID {
						for i, doc := range eval.Documents {
							if doc.ID == docObjID {
								documentToDownload = &eval.Documents[i]
								break
							}
						}
						break
					}
				}
			}
		}
	case "absence":
		// In Absences suchen
		if relatedID != "" {
			relObjID, err := primitive.ObjectIDFromHex(relatedID)
			if err == nil {
				for _, absence := range employee.Absences {
					if absence.ID == relObjID {
						for i, doc := range absence.Documents {
							if doc.ID == docObjID {
								documentToDownload = &absence.Documents[i]
								break
							}
						}
						break
					}
				}
			}
		}
	default:
		// In allgemeinen Dokumenten suchen
		for i, doc := range employee.Documents {
			if doc.ID == docObjID {
				documentToDownload = &employee.Documents[i]
				break
			}
		}
	}

	if documentToDownload == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dokument nicht gefunden"})
		return
	}

	// Datei bereitstellen
	c.FileAttachment(documentToDownload.FilePath, documentToDownload.FileName)
}

// AddTraining fügt ein neues Training für einen Mitarbeiter hinzu
func (h *DocumentHandler) AddTraining(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Formulardaten abrufen
	title := c.PostForm("title")
	description := c.PostForm("description")
	provider := c.PostForm("provider")
	certificate := c.PostForm("certificate")
	status := c.PostForm("status")
	notes := c.PostForm("notes")

	// Datumsangaben parsen
	startDateStr := c.PostForm("startDate")
	endDateStr := c.PostForm("endDate")

	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	// Neues Training erstellen
	training := model.Training{
		ID:          primitive.NewObjectID(),
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Provider:    provider,
		Certificate: certificate,
		Status:      status,
		Notes:       notes,
		Documents:   []model.Document{},
	}

	// Training zum Mitarbeiter hinzufügen
	employee.Trainings = append(employee.Trainings, training)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Training erfolgreich hinzugefügt",
		"training": training,
	})
}

// AddEvaluation fügt eine neue Leistungsbeurteilung für einen Mitarbeiter hinzu
// AddEvaluation fügt eine neue Leistungsbeurteilung für einen Mitarbeiter hinzu
func (h *DocumentHandler) AddEvaluation(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Formulardaten abrufen
	title := c.PostForm("title")
	dateStr := c.PostForm("date")
	overallRatingStr := c.PostForm("overallRating")
	strengths := c.PostForm("strengths")
	areasToImprove := c.PostForm("areasToImprove")
	comments := c.PostForm("comments")

	// Datum und Bewertung parsen
	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	overallRating := 3 // Standardwert
	if overallRatingStr != "" {
		rating, err := strconv.Atoi(overallRatingStr)
		if err == nil && rating >= 1 && rating <= 5 {
			overallRating = rating
		}
	}

	// Neue Evaluation erstellen
	evaluation := model.Evaluation{
		ID:               primitive.NewObjectID(),
		Title:            title,
		Date:             date,
		EvaluatorID:      userModel.ID,
		EvaluatorName:    userModel.FirstName + " " + userModel.LastName,
		OverallRating:    overallRating,
		Strengths:        strengths,
		AreasToImprove:   areasToImprove,
		Comments:         comments,
		EmployeeComments: "",
		Documents:        []model.Document{},
	}

	// Evaluation zum Mitarbeiter hinzufügen
	employee.Evaluations = append(employee.Evaluations, evaluation)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Leistungsbeurteilung erfolgreich hinzugefügt",
		"evaluation": evaluation,
	})
}

// AddAbsence fügt eine neue Abwesenheit für einen Mitarbeiter hinzu
func (h *DocumentHandler) AddAbsence(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Formulardaten abrufen
	absenceType := c.PostForm("type")
	startDateStr := c.PostForm("startDate")
	endDateStr := c.PostForm("endDate")
	reason := c.PostForm("reason")
	notes := c.PostForm("notes")

	// Datumsangaben parsen
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	} else {
		startDate = time.Now()
	}

	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	} else {
		endDate = startDate
	}

	// Anzahl der Tage berechnen
	days := endDate.Sub(startDate).Hours()/24 + 1

	// Neue Abwesenheit erstellen
	absence := model.Absence{
		ID:        primitive.NewObjectID(),
		Type:      absenceType,
		StartDate: startDate,
		EndDate:   endDate,
		Days:      days,
		Status:    "requested", // Standardmäßig beantragt
		Reason:    reason,
		Notes:     notes,
		Documents: []model.Document{},
	}

	// Falls der aktuelle Benutzer ein Administrator oder HR ist, automatisch genehmigen
	if userModel.Role == model.RoleAdmin || userModel.Role == model.RoleHR {
		absence.Status = "approved"
		absence.ApprovedBy = userModel.ID
		absence.ApproverName = userModel.FirstName + " " + userModel.LastName

		// Urlaubstage abziehen, falls es sich um Urlaub handelt
		if absenceType == "vacation" && absence.Status == "approved" {
			employee.RemainingVacation -= int(days)
			if employee.RemainingVacation < 0 {
				employee.RemainingVacation = 0
			}
		}
	}

	// Abwesenheit zum Mitarbeiter hinzufügen
	employee.Absences = append(employee.Absences, absence)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Abwesenheit erfolgreich hinzugefügt",
		"absence": absence,
	})
}

// AddDevelopmentItem fügt einen neuen Entwicklungspunkt für einen Mitarbeiter hinzu
func (h *DocumentHandler) AddDevelopmentItem(c *gin.Context) {
	// Mitarbeiter-ID aus dem URL-Parameter extrahieren
	employeeID := c.Param("id")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Formulardaten abrufen
	title := c.PostForm("title")
	description := c.PostForm("description")
	itemType := c.PostForm("type")
	targetDateStr := c.PostForm("targetDate")
	status := c.PostForm("status")
	notes := c.PostForm("notes")

	// Datum parsen
	var targetDate time.Time
	if targetDateStr != "" {
		targetDate, _ = time.Parse("2006-01-02", targetDateStr)
	}

	// Neuen Entwicklungspunkt erstellen
	developmentItem := model.DevelopmentItem{
		ID:          primitive.NewObjectID(),
		Title:       title,
		Description: description,
		Type:        itemType,
		TargetDate:  targetDate,
		Status:      status,
		Notes:       notes,
	}

	// Entwicklungspunkt zum Mitarbeiter hinzufügen
	employee.DevelopmentPlan = append(employee.DevelopmentPlan, developmentItem)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Entwicklungspunkt erfolgreich hinzugefügt",
		"developmentItem": developmentItem,
	})
}

// ApproveAbsence genehmigt oder lehnt eine Abwesenheitsanfrage ab
func (h *DocumentHandler) ApproveAbsence(c *gin.Context) {
	// Mitarbeiter-ID und Abwesenheits-ID aus den URL-Parametern extrahieren
	employeeID := c.Param("id")
	absenceID := c.Param("absenceId")

	// Aktion (approve/reject) aus dem Request-Body abrufen
	action := c.PostForm("action")
	if action != "approve" && action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Aktion. Verwenden Sie 'approve' oder 'reject'"})
		return
	}

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Aktuellen Benutzer aus dem Context abrufen
	user, _ := c.Get("user")
	userModel := user.(*model.User)

	// Abwesenheits-ID in ObjectID umwandeln
	absObjID, err := primitive.ObjectIDFromHex(absenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Abwesenheits-ID: " + err.Error()})
		return
	}

	// Abwesenheit finden und aktualisieren
	var absenceFound bool
	for i, absence := range employee.Absences {
		if absence.ID == absObjID {
			// Status aktualisieren
			if action == "approve" {
				employee.Absences[i].Status = "approved"

				// Urlaubstage abziehen, falls es sich um Urlaub handelt
				if absence.Type == "vacation" {
					employee.RemainingVacation -= int(absence.Days)
					if employee.RemainingVacation < 0 {
						employee.RemainingVacation = 0
					}
				}
			} else {
				employee.Absences[i].Status = "rejected"
			}

			// Genehmiger setzen
			employee.Absences[i].ApprovedBy = userModel.ID
			employee.Absences[i].ApproverName = userModel.FirstName + " " + userModel.LastName

			absenceFound = true
			break
		}
	}

	if !absenceFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Abwesenheit nicht gefunden"})
		return
	}

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Abwesenheitsanfrage erfolgreich " + (map[string]string{"approve": "genehmigt", "reject": "abgelehnt"})[action],
	})
}

// DeleteTraining löscht ein Training eines Mitarbeiters
func (h *DocumentHandler) DeleteTraining(c *gin.Context) {
	// 1) Parameter extrahieren
	empIDhex := c.Param("id")
	trainingIDhex := c.Param("trainingId")
	log.Printf("DeleteTraining aufgerufen für Mitarbeiter %s, Training %s", empIDhex, trainingIDhex)

	// 2) Objekt‑IDs parsen
	empObjID, err := primitive.ObjectIDFromHex(empIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Mitarbeiter‑ID"})
		return
	}
	trainObjID, err := primitive.ObjectIDFromHex(trainingIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Training‑ID"})
		return
	}

	// 3) Dokumente des Trainings aus dem Dateisystem löschen
	if employee, err := h.employeeRepo.FindByID(empIDhex); err == nil {
		for _, t := range employee.Trainings {
			if t.ID == trainObjID {
				for _, doc := range t.Documents {
					if err := h.fileService.DeleteFile(doc.FilePath); err != nil {
						log.Printf("Warnung: Konnte Trainings‑Dokument %s nicht löschen: %v", doc.FilePath, err)
					}
				}
				break
			}
		}
	}

	// 4) Trainings‑Eintrag in MongoDB direkt via $pull entfernen
	coll := db.GetCollection("employees")
	filter := bson.M{"_id": empObjID}
	update := bson.M{"$pull": bson.M{"trainings": bson.M{"_id": trainObjID}}}

	res, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("❌ MongoDB UpdateOne fehlgeschlagen: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB‑Fehler beim Löschen des Trainings"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Training nicht gefunden oder bereits gelöscht"})
		return
	}

	log.Printf("✅ Training %s erfolgreich gelöscht (ModifiedCount=%d)", trainingIDhex, res.ModifiedCount)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Training erfolgreich gelöscht",
	})
}

// DeleteEvaluation löscht eine Leistungsbeurteilung eines Mitarbeiters
func (h *DocumentHandler) DeleteEvaluation(c *gin.Context) {
	// 1) Parameter extrahieren
	empIDhex := c.Param("id")
	evalIDhex := c.Param("evaluationId")
	log.Printf("DeleteEvaluation aufgerufen für Mitarbeiter %s, Evaluation %s", empIDhex, evalIDhex)

	// 2) Objekt‑IDs parsen
	empObjID, err := primitive.ObjectIDFromHex(empIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Mitarbeiter‑ID"})
		return
	}
	evalObjID, err := primitive.ObjectIDFromHex(evalIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Evaluations‑ID"})
		return
	}

	// 3) Zuerst noch die zugehörigen Dokumente aus dem Dateisystem löschen,
	//    damit sie nicht übrig bleiben.
	if employee, err := h.employeeRepo.FindByID(empIDhex); err == nil {
		for _, ev := range employee.Evaluations {
			if ev.ID == evalObjID {
				for _, doc := range ev.Documents {
					if err := h.fileService.DeleteFile(doc.FilePath); err != nil {
						log.Printf("Warnung: Dokument %s konnte nicht gelöscht werden: %v", doc.FilePath, err)
					}
				}
				break
			}
		}
	}

	// 4) Jetzt das Array‑Element in MongoDB direkt mit $pull löschen
	coll := db.GetCollection("employees")
	filter := bson.M{"_id": empObjID}
	update := bson.M{"$pull": bson.M{"evaluations": bson.M{"_id": evalObjID}}}

	res, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("❌ MongoDB UpdateOne fehlgeschlagen: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB‑Fehler beim Löschen der Evaluierung"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evaluierung nicht gefunden oder bereits gelöscht"})
		return
	}

	log.Printf("✅ Evaluierung %s erfolgreich gelöscht (ModifiedCount=%d)", evalIDhex, res.ModifiedCount)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leistungsbeurteilung erfolgreich gelöscht",
	})
}

// DeleteAbsence löscht eine Abwesenheit eines Mitarbeiters
func (h *DocumentHandler) DeleteAbsence(c *gin.Context) {
	// 1) Parameter extrahieren
	empIDhex := c.Param("id")
	absIDhex := c.Param("absenceId")
	log.Printf("DeleteAbsence aufgerufen für Mitarbeiter %s, Abwesenheit %s", empIDhex, absIDhex)

	// 2) Objekt‑IDs parsen
	empObjID, err := primitive.ObjectIDFromHex(empIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Mitarbeiter‑ID"})
		return
	}
	absObjID, err := primitive.ObjectIDFromHex(absIDhex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Abwesenheits‑ID"})
		return
	}

	// 3) Dokumente aus dem Dateisystem löschen (falls vorhanden)
	if employee, err := h.employeeRepo.FindByID(empIDhex); err == nil {
		for _, abs := range employee.Absences {
			if abs.ID == absObjID {
				// Urlaubstage zurückgeben, sofern genehmigt
				if abs.Type == "vacation" && abs.Status == "approved" {
					employee.RemainingVacation += int(abs.Days)
				}
				for _, doc := range abs.Documents {
					if err := h.fileService.DeleteFile(doc.FilePath); err != nil {
						log.Printf("Warnung: Konnte Abwesenheits‑Dokument %s nicht löschen: %v", doc.FilePath, err)
					}
				}
				break
			}
		}
	}

	// 4) Array‑Eintrag direkt in MongoDB entfernen
	coll := db.GetCollection("employees")
	filter := bson.M{"_id": empObjID}
	update := bson.M{"$pull": bson.M{"absences": bson.M{"_id": absObjID}}}

	res, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("❌ MongoDB UpdateOne fehlgeschlagen: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB‑Fehler beim Löschen der Abwesenheit"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden"})
		return
	}
	if res.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Abwesenheit nicht gefunden oder bereits gelöscht"})
		return
	}

	log.Printf("✅ Abwesenheit %s erfolgreich gelöscht (ModifiedCount=%d)", absIDhex, res.ModifiedCount)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Abwesenheit erfolgreich gelöscht",
	})
}

// DeleteDevelopmentItem löscht einen Entwicklungspunkt eines Mitarbeiters
func (h *DocumentHandler) DeleteDevelopmentItem(c *gin.Context) {
	// Mitarbeiter-ID und DevelopmentItem-ID aus den URL-Parametern extrahieren
	employeeID := c.Param("id")
	itemID := c.Param("itemId")

	// Mitarbeiter abrufen
	employee, err := h.employeeRepo.FindByID(employeeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mitarbeiter nicht gefunden: " + err.Error()})
		return
	}

	// Item-ID in ObjectID umwandeln
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültige Item-ID: " + err.Error()})
		return
	}

	// Item finden und löschen
	var itemIndex int = -1

	for i, item := range employee.DevelopmentPlan {
		if item.ID == itemObjID {
			itemIndex = i
			break
		}
	}

	if itemIndex < 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entwicklungspunkt nicht gefunden"})
		return
	}

	// Item aus der Liste entfernen
	employee.DevelopmentPlan = append(employee.DevelopmentPlan[:itemIndex], employee.DevelopmentPlan[itemIndex+1:]...)

	// Mitarbeiter aktualisieren
	err = h.employeeRepo.Update(employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Aktualisieren des Mitarbeiters: " + err.Error()})
		return
	}

	// Erfolg zurückmelden
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Entwicklungspunkt erfolgreich gelöscht",
	})
}
