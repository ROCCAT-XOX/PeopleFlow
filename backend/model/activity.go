package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityType definiert die Art der Aktivität
type ActivityType string

const (
	ActivityTypeEmployeeAdded        ActivityType = "employee_added"
	ActivityTypeEmployeeUpdated      ActivityType = "employee_updated"
	ActivityTypeEmployeeDeleted      ActivityType = "employee_deleted"
	ActivityTypeVacationRequested    ActivityType = "vacation_requested"
	ActivityTypeVacationApproved     ActivityType = "vacation_approved"
	ActivityTypeVacationRejected     ActivityType = "vacation_rejected"
	ActivityTypeOvertimeAdjusted     ActivityType = "overtime_adjusted"
	ActivityTypeDocumentUploaded     ActivityType = "document_uploaded"
	ActivityTypeSystemSettingChanged ActivityType = "system_setting_changed"
)

// Activity repräsentiert eine System-Aktivität oder Aktion
type Activity struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        ActivityType       `bson:"type" json:"type"`
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserName    string             `bson:"userName" json:"userName"`
	TargetID    primitive.ObjectID `bson:"targetId,omitempty" json:"targetId,omitempty"`
	TargetType  string             `bson:"targetType,omitempty" json:"targetType,omitempty"`
	TargetName  string             `bson:"targetName,omitempty" json:"targetName,omitempty"`
	Description string             `bson:"description" json:"description"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	Metadata    map[string]string  `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// IsValid prüft, ob der ActivityType gültig ist
func (at ActivityType) IsValid() bool {
	switch at {
	case ActivityTypeEmployeeAdded, ActivityTypeEmployeeUpdated, ActivityTypeEmployeeDeleted,
		ActivityTypeVacationRequested, ActivityTypeVacationApproved, ActivityTypeVacationRejected,
		ActivityTypeOvertimeAdjusted, ActivityTypeDocumentUploaded, ActivityTypeSystemSettingChanged:
		return true
	default:
		return false
	}
}

// RequiresTarget prüft, ob dieser ActivityType ein Target benötigt
func (at ActivityType) RequiresTarget() bool {
	switch at {
	case ActivityTypeSystemSettingChanged:
		return false
	default:
		return true
	}
}

// GetLabel gibt ein benutzerfreundliches Label für den ActivityType zurück
func (at ActivityType) GetLabel() string {
	switch at {
	case ActivityTypeEmployeeAdded:
		return "Mitarbeiter hinzugefügt"
	case ActivityTypeEmployeeUpdated:
		return "Mitarbeiter aktualisiert"
	case ActivityTypeEmployeeDeleted:
		return "Mitarbeiter gelöscht"
	case ActivityTypeVacationRequested:
		return "Urlaub beantragt"
	case ActivityTypeVacationApproved:
		return "Urlaub genehmigt"
	case ActivityTypeVacationRejected:
		return "Urlaub abgelehnt"
	case ActivityTypeOvertimeAdjusted:
		return "Überstunden angepasst"
	case ActivityTypeDocumentUploaded:
		return "Dokument hochgeladen"
	case ActivityTypeSystemSettingChanged:
		return "Systemeinstellung geändert"
	default:
		return "Unbekannte Aktivität"
	}
}

// GetIcon gibt ein passendes Icon für den ActivityType zurück (für UI)
func (at ActivityType) GetIcon() string {
	switch at {
	case ActivityTypeEmployeeAdded:
		return "user-plus"
	case ActivityTypeEmployeeUpdated:
		return "user-edit"
	case ActivityTypeEmployeeDeleted:
		return "user-minus"
	case ActivityTypeVacationRequested, ActivityTypeVacationApproved, ActivityTypeVacationRejected:
		return "calendar"
	case ActivityTypeOvertimeAdjusted:
		return "clock"
	case ActivityTypeDocumentUploaded:
		return "file"
	case ActivityTypeSystemSettingChanged:
		return "settings"
	default:
		return "activity"
	}
}

// GetTimeAgo gibt eine relative Zeitangabe zurück
func (a *Activity) GetTimeAgo() string {
	duration := time.Since(a.Timestamp)

	if duration < time.Minute {
		return "gerade eben"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "vor 1 Minute"
		}
		return fmt.Sprintf("vor %d Minuten", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "vor 1 Stunde"
		}
		return fmt.Sprintf("vor %d Stunden", hours)
	} else if duration < 30*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "vor 1 Tag"
		}
		return fmt.Sprintf("vor %d Tagen", days)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "vor 1 Monat"
		}
		return fmt.Sprintf("vor %d Monaten", months)
	} else {
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "vor 1 Jahr"
		}
		return fmt.Sprintf("vor %d Jahren", years)
	}
}
