package utils

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"time"
)

// EnsureUploadDirExists prüft, ob das Upload-Verzeichnis existiert und erstellt es, falls es nicht existiert
func EnsureUploadDirExists() error {
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return os.MkdirAll(uploadDir, 0755)
	}
	return nil
}

// TemplateHelpers gibt eine Map mit Hilfsfunktionen für Templates zurück
func TemplateHelpers() template.FuncMap {
	return template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatDate": func(date time.Time) string {
			return date.Format("02.01.2006")
		},
		"formatDateTime": func(date time.Time) string {
			return date.Format("02.01.2006 15:04")
		},
		"formatFileSize": func(size int64) string {
			const unit = 1024
			if size < unit {
				return fmt.Sprintf("%d B", size)
			}
			div, exp := int64(unit), 0
			for n := size / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
		},
		"iterate": func(count int) []int {
			var i []int
			for j := 0; j < count; j++ {
				i = append(i, j)
			}
			return i
		},
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"divide": func(a, b int) float64 {
			if b == 0 {
				return 0
			}
			return float64(a) / float64(b)
		},
		"round": func(num float64) int {
			return int(math.Round(num))
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"neq": func(a, b interface{}) bool {
			return a != b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		"lte": func(a, b int) bool {
			return a <= b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"gte": func(a, b int) bool {
			return a >= b
		},
		"now": func() time.Time {
			return time.Now()
		},
	}
}
