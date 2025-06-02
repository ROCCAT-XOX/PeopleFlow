package utils

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

// GetInitials extrahiert die Initialen aus einem Vor- und Nachnamen
func GetInitials(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "?"
	}

	if len(parts) == 1 {
		if len(parts[0]) > 0 {
			return strings.ToUpper(string(parts[0][0]))
		}
		return "?"
	}

	// Erste Buchstaben von Vor- und Nachname
	firstInitial := string(parts[0][0])
	lastInitial := string(parts[len(parts)-1][0])

	return strings.ToUpper(firstInitial + lastInitial)
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
		"add": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va + vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va + vb
				}
				if vb, ok := b.(int); ok {
					return va + float64(vb)
				}
			}
			return 0
		},
		"subtract": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va - vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va - vb
				}
				if vb, ok := b.(int); ok {
					return va - float64(vb)
				}
			}
			return 0
		},
		"multiply": func(a, b interface{}) interface{} {
			switch va := a.(type) {
			case int:
				if vb, ok := b.(int); ok {
					return va * vb
				}
			case float64:
				if vb, ok := b.(float64); ok {
					return va * vb
				}
				if vb, ok := b.(int); ok {
					return va * float64(vb)
				}
			}
			return 0
		},
		"divide": func(a, b interface{}) float64 {
			var fa, fb float64

			switch va := a.(type) {
			case int:
				fa = float64(va)
			case float64:
				fa = va
			default:
				return 0
			}

			switch vb := b.(type) {
			case int:
				fb = float64(vb)
			case float64:
				fb = vb
			default:
				return 0
			}

			if fb == 0 {
				return 0
			}
			return fa / fb
		},
		"round": func(num float64) int {
			return int(math.Round(num))
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"neq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) != fmt.Sprintf("%v", b)
		},
		// Erweiterte Vergleichsfunktionen für verschiedene Typen
		"lt": func(a, b interface{}) bool {
			return compareValues(a, b) < 0
		},
		"lte": func(a, b interface{}) bool {
			return compareValues(a, b) <= 0
		},
		"gt": func(a, b interface{}) bool {
			return compareValues(a, b) > 0
		},
		"gte": func(a, b interface{}) bool {
			return compareValues(a, b) >= 0
		},
		"ge": func(a, b interface{}) bool { // Alias für gte
			return compareValues(a, b) >= 0
		},
		"le": func(a, b interface{}) bool { // Alias für lte
			return compareValues(a, b) <= 0
		},
		"now": func() time.Time {
			return time.Now()
		},
		"isoWeek": func(t time.Time) int {
			_, week := t.ISOWeek()
			return week
		},
		"abs": func(x interface{}) interface{} {
			switch v := x.(type) {
			case int:
				if v < 0 {
					return -v
				}
				return v
			case float64:
				if v < 0 {
					return -v
				}
				return v
			default:
				return x
			}
		},
		"getInitials": GetInitials,
	}
}

// compareValues vergleicht zwei Werte verschiedener numerischer Typen
func compareValues(a, b interface{}) int {
	var fa, fb float64

	// Konvertiere a zu float64
	switch va := a.(type) {
	case int:
		fa = float64(va)
	case int32:
		fa = float64(va)
	case int64:
		fa = float64(va)
	case float32:
		fa = float64(va)
	case float64:
		fa = va
	default:
		return 0 // Unbekannter Typ
	}

	// Konvertiere b zu float64
	switch vb := b.(type) {
	case int:
		fb = float64(vb)
	case int32:
		fb = float64(vb)
	case int64:
		fb = float64(vb)
	case float32:
		fb = float64(vb)
	case float64:
		fb = vb
	default:
		return 0 // Unbekannter Typ
	}

	// Vergleiche
	if fa < fb {
		return -1
	} else if fa > fb {
		return 1
	}
	return 0
}
