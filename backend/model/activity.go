package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityType definiert die Art der Aktivität
type ActivityType string

const (
	ActivityTypeEmployeeAdded           ActivityType = "employee_added"
	ActivityTypeEmployeeUpdated         ActivityType = "employee_updated"
	ActivityTypeEmployeeDeleted         ActivityType = "employee_deleted"
	ActivityTypeVacationRequested       ActivityType = "vacation_requested"
	ActivityTypeVacationApproved        ActivityType = "vacation_approved"
	ActivityTypeVacationRejected        ActivityType = "vacation_rejected"
	ActivityTypeOvertimeAdjusted        ActivityType = "overtime_adjusted"
	ActivityTypeDocumentUploaded        ActivityType = "document_uploaded"
	ActivityTypeSystemSettingChanged    ActivityType = "system_setting_changed"
	ActivityTypeConversationAdded       ActivityType = "conversation_added"
	ActivityTypeConversationCompleted   ActivityType = "conversation_completed"
	ActivityTypeConversationUpdated     ActivityType = "conversation_updated"
	ActivityTypeUserAdded               ActivityType = "user_added"
	ActivityTypeUserUpdated             ActivityType = "user_updated"
	ActivityTypeUserDeleted             ActivityType = "user_deleted"
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
		ActivityTypeOvertimeAdjusted, ActivityTypeDocumentUploaded, ActivityTypeSystemSettingChanged,
		ActivityTypeConversationAdded, ActivityTypeConversationCompleted, ActivityTypeConversationUpdated,
		ActivityTypeUserAdded, ActivityTypeUserUpdated, ActivityTypeUserDeleted:
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

// GetIconClass returns the CSS class for the activity icon
func (a *Activity) GetIconClass() string {
	switch a.Type {
	case ActivityTypeEmployeeAdded:
		return "text-green-500"
	case ActivityTypeEmployeeUpdated:
		return "text-blue-500"
	case ActivityTypeEmployeeDeleted:
		return "text-red-500"
	case ActivityTypeVacationRequested:
		return "text-yellow-500"
	case ActivityTypeVacationApproved:
		return "text-green-500"
	case ActivityTypeVacationRejected:
		return "text-red-500"
	case ActivityTypeOvertimeAdjusted:
		return "text-purple-500"
	case ActivityTypeDocumentUploaded:
		return "text-blue-500"
	case ActivityTypeSystemSettingChanged:
		return "text-gray-500"
	case ActivityTypeConversationAdded, ActivityTypeConversationCompleted, ActivityTypeConversationUpdated:
		return "text-indigo-500"
	case ActivityTypeUserAdded, ActivityTypeUserUpdated, ActivityTypeUserDeleted:
		return "text-orange-500"
	default:
		return "text-gray-400"
	}
}

// GetIconSVG returns the SVG icon for the activity
func (a *Activity) GetIconSVG() string {
	switch a.Type {
	case ActivityTypeEmployeeAdded:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>`
	case ActivityTypeEmployeeUpdated:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>`
	case ActivityTypeEmployeeDeleted:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>`
	case ActivityTypeVacationRequested, ActivityTypeVacationApproved, ActivityTypeVacationRejected:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>`
	case ActivityTypeOvertimeAdjusted:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`
	case ActivityTypeDocumentUploaded:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M9 19l3 3m0 0l3-3m-3 3V10"/></svg>`
	case ActivityTypeSystemSettingChanged:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>`
	case ActivityTypeConversationAdded, ActivityTypeConversationCompleted, ActivityTypeConversationUpdated:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/></svg>`
	case ActivityTypeUserAdded, ActivityTypeUserUpdated, ActivityTypeUserDeleted:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/></svg>`
	default:
		return `<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`
	}
}


// Validate checks if the activity is valid
func (a *Activity) Validate() error {
	if !a.Type.IsValid() {
		return fmt.Errorf("invalid activity type: %s", a.Type)
	}

	if a.UserID.IsZero() {
		return fmt.Errorf("user ID is required")
	}

	if a.Type.RequiresTarget() && a.TargetID.IsZero() {
		return fmt.Errorf("target ID is required for activity type: %s", a.Type)
	}

	if a.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}
