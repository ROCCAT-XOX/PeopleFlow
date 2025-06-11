package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestActivityType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		actType  ActivityType
		expected bool
	}{
		{
			name:     "Valid - Employee Added",
			actType:  ActivityTypeEmployeeAdded,
			expected: true,
		},
		{
			name:     "Valid - Employee Updated",
			actType:  ActivityTypeEmployeeUpdated,
			expected: true,
		},
		{
			name:     "Valid - Employee Deleted",
			actType:  ActivityTypeEmployeeDeleted,
			expected: true,
		},
		{
			name:     "Valid - Vacation Requested",
			actType:  ActivityTypeVacationRequested,
			expected: true,
		},
		{
			name:     "Valid - Vacation Approved",
			actType:  ActivityTypeVacationApproved,
			expected: true,
		},
		{
			name:     "Valid - Vacation Rejected",
			actType:  ActivityTypeVacationRejected,
			expected: true,
		},
		{
			name:     "Valid - Overtime Adjusted",
			actType:  ActivityTypeOvertimeAdjusted,
			expected: true,
		},
		{
			name:     "Valid - Document Uploaded",
			actType:  ActivityTypeDocumentUploaded,
			expected: true,
		},
		{
			name:     "Valid - System Setting Changed",
			actType:  ActivityTypeSystemSettingChanged,
			expected: true,
		},
		{
			name:     "Valid - Conversation Added",
			actType:  ActivityTypeConversationAdded,
			expected: true,
		},
		{
			name:     "Valid - Conversation Completed",
			actType:  ActivityTypeConversationCompleted,
			expected: true,
		},
		{
			name:     "Valid - Conversation Updated",
			actType:  ActivityTypeConversationUpdated,
			expected: true,
		},
		{
			name:     "Valid - User Added",
			actType:  ActivityTypeUserAdded,
			expected: true,
		},
		{
			name:     "Valid - User Updated",
			actType:  ActivityTypeUserUpdated,
			expected: true,
		},
		{
			name:     "Valid - User Deleted",
			actType:  ActivityTypeUserDeleted,
			expected: true,
		},
		{
			name:     "Invalid - Unknown Type",
			actType:  ActivityType("unknown_type"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.actType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivityType_RequiresTarget(t *testing.T) {
	tests := []struct {
		name     string
		actType  ActivityType
		expected bool
	}{
		{
			name:     "System Setting Changed - No Target",
			actType:  ActivityTypeSystemSettingChanged,
			expected: false,
		},
		{
			name:     "Employee Added - Requires Target",
			actType:  ActivityTypeEmployeeAdded,
			expected: true,
		},
		{
			name:     "Document Uploaded - Requires Target",
			actType:  ActivityTypeDocumentUploaded,
			expected: true,
		},
		{
			name:     "User Updated - Requires Target",
			actType:  ActivityTypeUserUpdated,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.actType.RequiresTarget()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivityType_GetLabel(t *testing.T) {
	tests := []struct {
		name     string
		actType  ActivityType
		expected string
	}{
		{
			name:     "Employee Added",
			actType:  ActivityTypeEmployeeAdded,
			expected: "Mitarbeiter hinzugefügt",
		},
		{
			name:     "Employee Updated",
			actType:  ActivityTypeEmployeeUpdated,
			expected: "Mitarbeiter aktualisiert",
		},
		{
			name:     "Employee Deleted",
			actType:  ActivityTypeEmployeeDeleted,
			expected: "Mitarbeiter gelöscht",
		},
		{
			name:     "Vacation Requested",
			actType:  ActivityTypeVacationRequested,
			expected: "Urlaub beantragt",
		},
		{
			name:     "Vacation Approved",
			actType:  ActivityTypeVacationApproved,
			expected: "Urlaub genehmigt",
		},
		{
			name:     "Vacation Rejected",
			actType:  ActivityTypeVacationRejected,
			expected: "Urlaub abgelehnt",
		},
		{
			name:     "Overtime Adjusted",
			actType:  ActivityTypeOvertimeAdjusted,
			expected: "Überstunden angepasst",
		},
		{
			name:     "Document Uploaded",
			actType:  ActivityTypeDocumentUploaded,
			expected: "Dokument hochgeladen",
		},
		{
			name:     "System Setting Changed",
			actType:  ActivityTypeSystemSettingChanged,
			expected: "Systemeinstellung geändert",
		},
		{
			name:     "Unknown Type",
			actType:  ActivityType("unknown"),
			expected: "Unbekannte Aktivität",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.actType.GetLabel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivityType_GetIcon(t *testing.T) {
	tests := []struct {
		name     string
		actType  ActivityType
		expected string
	}{
		{
			name:     "Employee Added",
			actType:  ActivityTypeEmployeeAdded,
			expected: "user-plus",
		},
		{
			name:     "Employee Updated",
			actType:  ActivityTypeEmployeeUpdated,
			expected: "user-edit",
		},
		{
			name:     "Employee Deleted",
			actType:  ActivityTypeEmployeeDeleted,
			expected: "user-minus",
		},
		{
			name:     "Vacation Activities",
			actType:  ActivityTypeVacationRequested,
			expected: "calendar",
		},
		{
			name:     "Overtime Adjusted",
			actType:  ActivityTypeOvertimeAdjusted,
			expected: "clock",
		},
		{
			name:     "Document Uploaded",
			actType:  ActivityTypeDocumentUploaded,
			expected: "file",
		},
		{
			name:     "System Setting Changed",
			actType:  ActivityTypeSystemSettingChanged,
			expected: "settings",
		},
		{
			name:     "Unknown Type",
			actType:  ActivityType("unknown"),
			expected: "activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.actType.GetIcon()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivity_GetTimeAgo(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "Just now",
			timestamp: time.Now().Add(-30 * time.Second),
			expected:  "gerade eben",
		},
		{
			name:      "1 minute ago",
			timestamp: time.Now().Add(-1 * time.Minute),
			expected:  "vor 1 Minute",
		},
		{
			name:      "5 minutes ago",
			timestamp: time.Now().Add(-5 * time.Minute),
			expected:  "vor 5 Minuten",
		},
		{
			name:      "1 hour ago",
			timestamp: time.Now().Add(-1 * time.Hour),
			expected:  "vor 1 Stunde",
		},
		{
			name:      "3 hours ago",
			timestamp: time.Now().Add(-3 * time.Hour),
			expected:  "vor 3 Stunden",
		},
		{
			name:      "1 day ago",
			timestamp: time.Now().Add(-24 * time.Hour),
			expected:  "vor 1 Tag",
		},
		{
			name:      "5 days ago",
			timestamp: time.Now().Add(-5 * 24 * time.Hour),
			expected:  "vor 5 Tagen",
		},
		{
			name:      "1 month ago",
			timestamp: time.Now().Add(-30 * 24 * time.Hour),
			expected:  "vor 1 Monat",
		},
		{
			name:      "6 months ago",
			timestamp: time.Now().Add(-180 * 24 * time.Hour),
			expected:  "vor 6 Monaten",
		},
		{
			name:      "1 year ago",
			timestamp: time.Now().Add(-365 * 24 * time.Hour),
			expected:  "vor 1 Jahr",
		},
		{
			name:      "2 years ago",
			timestamp: time.Now().Add(-730 * 24 * time.Hour),
			expected:  "vor 2 Jahren",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := &Activity{
				Timestamp: tt.timestamp,
			}
			result := activity.GetTimeAgo()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivity_GetIconClass(t *testing.T) {
	tests := []struct {
		name     string
		activity *Activity
		expected string
	}{
		{
			name:     "Employee Added - Green",
			activity: &Activity{Type: ActivityTypeEmployeeAdded},
			expected: "text-green-500",
		},
		{
			name:     "Employee Updated - Blue",
			activity: &Activity{Type: ActivityTypeEmployeeUpdated},
			expected: "text-blue-500",
		},
		{
			name:     "Employee Deleted - Red",
			activity: &Activity{Type: ActivityTypeEmployeeDeleted},
			expected: "text-red-500",
		},
		{
			name:     "Vacation Requested - Yellow",
			activity: &Activity{Type: ActivityTypeVacationRequested},
			expected: "text-yellow-500",
		},
		{
			name:     "Vacation Approved - Green",
			activity: &Activity{Type: ActivityTypeVacationApproved},
			expected: "text-green-500",
		},
		{
			name:     "Vacation Rejected - Red",
			activity: &Activity{Type: ActivityTypeVacationRejected},
			expected: "text-red-500",
		},
		{
			name:     "Overtime Adjusted - Purple",
			activity: &Activity{Type: ActivityTypeOvertimeAdjusted},
			expected: "text-purple-500",
		},
		{
			name:     "Document Uploaded - Blue",
			activity: &Activity{Type: ActivityTypeDocumentUploaded},
			expected: "text-blue-500",
		},
		{
			name:     "System Setting Changed - Gray",
			activity: &Activity{Type: ActivityTypeSystemSettingChanged},
			expected: "text-gray-500",
		},
		{
			name:     "Conversation Activities - Indigo",
			activity: &Activity{Type: ActivityTypeConversationAdded},
			expected: "text-indigo-500",
		},
		{
			name:     "User Activities - Orange",
			activity: &Activity{Type: ActivityTypeUserAdded},
			expected: "text-orange-500",
		},
		{
			name:     "Unknown Type - Gray",
			activity: &Activity{Type: ActivityType("unknown")},
			expected: "text-gray-400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.activity.GetIconClass()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestActivity_GetIconSVG(t *testing.T) {
	tests := []struct {
		name     string
		activity *Activity
		contains string
	}{
		{
			name:     "Employee Added",
			activity: &Activity{Type: ActivityTypeEmployeeAdded},
			contains: "M12 4v16m8-8H4",
		},
		{
			name:     "Employee Updated",
			activity: &Activity{Type: ActivityTypeEmployeeUpdated},
			contains: "M11 5H6a2 2 0 00-2",
		},
		{
			name:     "Employee Deleted",
			activity: &Activity{Type: ActivityTypeEmployeeDeleted},
			contains: "M19 7l-.867 12.142",
		},
		{
			name:     "Vacation Activities",
			activity: &Activity{Type: ActivityTypeVacationRequested},
			contains: "M8 7V3m8 4V3m-9",
		},
		{
			name:     "Overtime Adjusted",
			activity: &Activity{Type: ActivityTypeOvertimeAdjusted},
			contains: "M12 8v4l3 3m6-3",
		},
		{
			name:     "Document Uploaded",
			activity: &Activity{Type: ActivityTypeDocumentUploaded},
			contains: "M7 16a4 4 0 01-.88",
		},
		{
			name:     "System Setting Changed",
			activity: &Activity{Type: ActivityTypeSystemSettingChanged},
			contains: "M10.325 4.317c.426",
		},
		{
			name:     "User Activities",
			activity: &Activity{Type: ActivityTypeUserAdded},
			contains: "M16 7a4 4 0 11-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.activity.GetIconSVG()
			assert.Contains(t, result, tt.contains)
			assert.Contains(t, result, `<svg`)
			assert.Contains(t, result, `</svg>`)
		})
	}
}

func TestActivity_Validate(t *testing.T) {
	tests := []struct {
		name        string
		activity    *Activity
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid activity",
			activity: &Activity{
				Type:        ActivityTypeEmployeeAdded,
				UserID:      primitive.NewObjectID(),
				UserName:    "John Doe",
				TargetID:    primitive.NewObjectID(),
				TargetType:  "employee",
				TargetName:  "Jane Smith",
				Description: "Added new employee",
				Timestamp:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "Invalid activity type",
			activity: &Activity{
				Type:        ActivityType("invalid"),
				UserID:      primitive.NewObjectID(),
				UserName:    "John Doe",
				Description: "Invalid activity",
				Timestamp:   time.Now(),
			},
			expectError: true,
			errorMsg:    "invalid activity type",
		},
		{
			name: "Missing user ID",
			activity: &Activity{
				Type:        ActivityTypeEmployeeAdded,
				UserName:    "John Doe",
				Description: "Missing user ID",
				Timestamp:   time.Now(),
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name: "Missing target for activity that requires it",
			activity: &Activity{
				Type:        ActivityTypeEmployeeAdded,
				UserID:      primitive.NewObjectID(),
				UserName:    "John Doe",
				Description: "Missing target",
				Timestamp:   time.Now(),
			},
			expectError: true,
			errorMsg:    "target ID is required",
		},
		{
			name: "System setting change without target",
			activity: &Activity{
				Type:        ActivityTypeSystemSettingChanged,
				UserID:      primitive.NewObjectID(),
				UserName:    "Admin",
				Description: "Changed system settings",
				Timestamp:   time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.activity.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}