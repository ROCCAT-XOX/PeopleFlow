package service

import (
	"PeoplePilot/backend/model"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileService verwaltet Dateioperationen wie Upload und Löschung
type FileService struct {
	uploadDir string
}

// NewFileService erstellt einen neuen FileService
func NewFileService() *FileService {
	// Stellt sicher, dass das Upload-Verzeichnis existiert
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	return &FileService{
		uploadDir: uploadDir,
	}
}

// UploadFile lädt eine Datei hoch und erstellt ein Document-Objekt
func (s *FileService) UploadFile(file *multipart.FileHeader, name, description, category string, uploaderID primitive.ObjectID) (*model.Document, error) {
	// Generiere eine eindeutige ID für die Datei
	documentID := primitive.NewObjectID()

	// Erstelle einen eindeutigen Dateinamen
	originalFilename := filepath.Base(file.Filename)
	extension := filepath.Ext(originalFilename)
	uniqueFilename := fmt.Sprintf("%s%s", documentID.Hex(), extension)

	// Definiere den vollständigen Pfad, unter dem die Datei gespeichert wird
	filePath := filepath.Join(s.uploadDir, uniqueFilename)

	// Erstelle die Zieldatei
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen der Zieldatei: %v", err)
	}
	defer dst.Close()

	// Öffne die hochgeladene Datei
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Öffnen der hochgeladenen Datei: %v", err)
	}
	defer src.Close()

	// Kopiere den Inhalt der hochgeladenen Datei in die Zieldatei
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("Fehler beim Kopieren der Datei: %v", err)
	}

	// Erstelle das Document-Objekt
	document := &model.Document{
		ID:          documentID,
		Name:        name,
		FileName:    originalFilename,
		FileType:    file.Header.Get("Content-Type"),
		Description: description,
		Category:    category,
		FilePath:    filePath,
		FileSize:    file.Size,
		UploadDate:  time.Now(),
		UploadedBy:  uploaderID,
	}

	return document, nil
}

// DeleteFile löscht eine Datei aus dem Dateisystem
func (s *FileService) DeleteFile(filePath string) error {
	if filePath == "" {
		return errors.New("Leerer Dateipfad")
	}

	// Prüfen, ob die Datei existiert
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.New("Datei nicht gefunden")
	}

	// Datei löschen
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("Fehler beim Löschen der Datei: %v", err)
	}

	return nil
}

// GetFilePath gibt den vollen Dateipfad zurück
func (s *FileService) GetFilePath(fileName string) string {
	return filepath.Join(s.uploadDir, fileName)
}

// UploadProfileImage lädt ein Profilbild hoch und gibt den Dateipfad zurück
func (s *FileService) UploadProfileImage(file *multipart.FileHeader, employeeID string) (string, error) {
	// Generiere einen eindeutigen Dateinamen
	originalFilename := filepath.Base(file.Filename)
	extension := filepath.Ext(originalFilename)
	uniqueFilename := fmt.Sprintf("profile_%s%s", employeeID, extension)

	// Definiere den relativen Pfad für die Datenbank (dieser wird im HTML angezeigt)
	relativePath := fmt.Sprintf("/static/uploads/%s", uniqueFilename)

	// Definiere den vollständigen Pfad, unter dem die Datei gespeichert wird
	// Geändert von "." zu "./frontend", um mit der router.Static Konfiguration übereinzustimmen
	filePath := filepath.Join(".", "frontend", "static", "uploads", uniqueFilename)

	// Stelle sicher, dass das Verzeichnis existiert
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Erstellen des Verzeichnisses: %v", err)
	}

	// Erstelle die Zieldatei
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Erstellen der Zieldatei: %v", err)
	}
	defer dst.Close()

	// Öffne die hochgeladene Datei
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("Fehler beim Öffnen der hochgeladenen Datei: %v", err)
	}
	defer src.Close()

	// Kopiere den Inhalt der hochgeladenen Datei in die Zieldatei
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("Fehler beim Kopieren der Datei: %v", err)
	}

	return relativePath, nil
}
