package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ActivityType repräsentiert den Typ einer Aktivität
type ActivityType string

const (
	ActivityTypeEmployeeAdded     ActivityType = "employee_added"
	ActivityTypeEmployeeUpdated   ActivityType = "employee_updated"
	ActivityTypeEmployeeDeleted   ActivityType = "employee_deleted"
	ActivityTypeVacationRequested ActivityType = "vacation_requested"
	ActivityTypeVacationApproved  ActivityType = "vacation_approved"
	ActivityTypeVacationRejected  ActivityType = "vacation_rejected"
	ActivityTypeDocumentUploaded  ActivityType = "document_uploaded"
	ActivityTypeTrainingAdded     ActivityType = "training_added"
	ActivityTypeEvaluationAdded   ActivityType = "evaluation_added"
	ActivityTypeUserAdded         ActivityType = "user_added"
	ActivityTypeUserUpdated       ActivityType = "user_updated"
	ActivityTypeUserDeleted       ActivityType = "user_deleted"
)

// Activity repräsentiert eine Aktivität im System
type Activity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        ActivityType       `bson:"type" json:"type"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserName    string             `bson:"userName" json:"userName"`
	TargetID    primitive.ObjectID `bson:"targetId,omitempty" json:"targetId,omitempty"`
	TargetType  string             `bson:"targetType" json:"targetType"`
	TargetName  string             `bson:"targetName" json:"targetName"`
	Description string             `bson:"description" json:"description"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}

// GetIconClass gibt die CSS-Klasse für das Icon basierend auf dem Aktivitätstyp zurück
func (a *Activity) GetIconClass() string {
	switch a.Type {
	case ActivityTypeEmployeeAdded, ActivityTypeTrainingAdded, ActivityTypeEvaluationAdded:
		return "bg-green-500"
	case ActivityTypeEmployeeUpdated, ActivityTypeDocumentUploaded:
		return "bg-blue-500"
	case ActivityTypeVacationRequested, ActivityTypeVacationApproved:
		return "bg-yellow-500"
	case ActivityTypeEmployeeDeleted, ActivityTypeVacationRejected:
		return "bg-red-500"
	default:
		return "bg-gray-500"
	}
}

// GetIconSVG gibt das SVG-Icon basierend auf dem Aktivitätstyp zurück
func (a *Activity) GetIconSVG() string {
	switch a.Type {
	case ActivityTypeEmployeeAdded:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z\" /></svg>"
	case ActivityTypeEmployeeUpdated:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z\" /></svg>"
	case ActivityTypeVacationRequested, ActivityTypeVacationApproved:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z\" clip-rule=\"evenodd\" /></svg>"
	case ActivityTypeDocumentUploaded:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z\" clip-rule=\"evenodd\" /><path fill-rule=\"evenodd\" d=\"M8 11a1 1 0 10-2 0v2a1 1 0 102 0v-2zm2-3a1 1 0 00-1 1v4a1 1 0 102 0V9a1 1 0 00-1-1z\" clip-rule=\"evenodd\" /></svg>"
	case ActivityTypeTrainingAdded, ActivityTypeEvaluationAdded:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M10.394 2.08a1 1 0 00-.788 0l-7 3a1 1 0 000 1.84L5.25 8.051a.999.999 0 01.356-.257l4-1.714a1 1 0 11.788 1.838L7.667 9.088l1.94.831a1 1 0 00.787 0l7-3a1 1 0 000-1.838l-7-3zM3.31 9.397L5 10.12v4.102a8.969 8.969 0 00-1.05-.174 1 1 0 01-.89-.89 11.115 11.115 0 01.25-3.762zM9.3 16.573A9.026 9.026 0 007 14.935v-3.957l1.818.78a3 3 0 002.364 0l5.508-2.361a11.026 11.026 0 01.25 3.762 1 1 0 01-.89.89 8.968 8.968 0 00-5.35 2.524 1 1 0 01-1.4 0zM6 18a1 1 0 001-1v-2.065a8.935 8.935 0 00-2-.712V17a1 1 0 001 1z\" /></svg>"
	case ActivityTypeUserAdded:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z\" /></svg>"
	case ActivityTypeUserUpdated:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path d=\"M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z\" /></svg>"
	default:
		return "<svg class=\"h-5 w-5 text-white\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z\" clip-rule=\"evenodd\" /></svg>"
	}
}

// FormatTimeAgo formatiert den Zeitstempel als "vor X Zeit" (z.B. "vor 5 Minuten")
func (a *Activity) FormatTimeAgo() string {
	now := time.Now()
	diff := now.Sub(a.Timestamp)

	if diff < time.Minute {
		return "gerade eben"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("vor %d Minute%s", minutes, pluralS(minutes))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("vor %d Stunde%s", hours, pluralS(hours))
	} else if diff < 48*time.Hour {
		return "gestern"
	} else {
		return a.Timestamp.Format("02.01.2006 15:04")
	}
}

// Hilfsfunktion für Plural-Endungen
func pluralS(count int) string {
	if count == 1 {
		return ""
	}
	return "n"
}
