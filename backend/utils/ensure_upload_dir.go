package utils

import (
	"os"
)

// EnsureUploadDirExists prüft, ob das Upload-Verzeichnis existiert und erstellt es, falls es nicht existiert
func EnsureUploadDirExists() error {
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(uploadDir, 0755)
	}
	return nil
}
