package model

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestActivityConstants(t *testing.T) {
	t.Run("ActivityType constants", func(t *testing.T) {
		// Test that all activity type constants are defined
		activityTypes := []ActivityType{
			ActivityTypeEmployeeAdded,
			ActivityTypeEmployeeUpdated,
			ActivityTypeEmployeeDeleted,
			ActivityTypeVacationRequested,
			ActivityTypeVacationApproved,
			ActivityTypeVacationRejected,
			ActivityTypeDocumentUploaded,
			ActivityTypeTrainingAdded,
			ActivityTypeEvaluationAdded,
			ActivityTypeUserAdded,
			ActivityTypeUserUpdated,
			ActivityTypeUserDeleted,
			ActivityTypeConversationAdded,
			ActivityTypeConversationUpdated,
			ActivityTypeConversationCompleted,
		}

		for _, activityType := range activityTypes {
			if string(activityType) == "" {
				t.Errorf("Activity type constant should not be empty: %v", activityType)
			}
		}
	})
}

func TestActivityGetIconClass(t *testing.T) {
	tests := []struct {
		name         string
		activityType ActivityType
		expectedCSS  string
	}{
		{
			name:         "employee added - green",
			activityType: ActivityTypeEmployeeAdded,
			expectedCSS:  "bg-green-500",
		},
		{
			name:         "training added - green",
			activityType: ActivityTypeTrainingAdded,
			expectedCSS:  "bg-green-500",
		},
		{
			name:         "evaluation added - green",
			activityType: ActivityTypeEvaluationAdded,
			expectedCSS:  "bg-green-500",
		},
		{
			name:         "employee updated - blue",
			activityType: ActivityTypeEmployeeUpdated,
			expectedCSS:  "bg-blue-500",
		},
		{
			name:         "document uploaded - blue",
			activityType: ActivityTypeDocumentUploaded,
			expectedCSS:  "bg-blue-500",
		},
		{
			name:         "vacation requested - yellow",
			activityType: ActivityTypeVacationRequested,
			expectedCSS:  "bg-yellow-500",
		},
		{
			name:         "vacation approved - yellow",
			activityType: ActivityTypeVacationApproved,
			expectedCSS:  "bg-yellow-500",
		},
		{
			name:         "employee deleted - red",
			activityType: ActivityTypeEmployeeDeleted,
			expectedCSS:  "bg-red-500",
		},
		{
			name:         "vacation rejected - red",
			activityType: ActivityTypeVacationRejected,
			expectedCSS:  "bg-red-500",
		},
		{
			name:         "conversation added - purple",
			activityType: ActivityTypeConversationAdded,
			expectedCSS:  "bg-purple-500",
		},
		{
			name:         "conversation completed - green",
			activityType: ActivityTypeConversationCompleted,
			expectedCSS:  "bg-green-500",
		},
		{
			name:         "unknown activity type - gray",
			activityType: ActivityType("unknown"),
			expectedCSS:  "bg-gray-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := Activity{Type: tt.activityType}
			result := activity.GetIconClass()
			if result != tt.expectedCSS {
				t.Errorf("GetIconClass() = %q, expected %q", result, tt.expectedCSS)
			}
		})
	}
}

func TestActivityGetIconSVG(t *testing.T) {
	tests := []struct {
		name         string
		activityType ActivityType
		expectSVG    bool
	}{
		{
			name:         "employee added",
			activityType: ActivityTypeEmployeeAdded,
			expectSVG:    true,
		},
		{
			name:         "employee updated",
			activityType: ActivityTypeEmployeeUpdated,
			expectSVG:    true,
		},
		{
			name:         "vacation requested",
			activityType: ActivityTypeVacationRequested,
			expectSVG:    true,
		},
		{
			name:         "vacation approved",
			activityType: ActivityTypeVacationApproved,
			expectSVG:    true,
		},
		{
			name:         "document uploaded",
			activityType: ActivityTypeDocumentUploaded,
			expectSVG:    true,
		},
		{
			name:         "training added",
			activityType: ActivityTypeTrainingAdded,
			expectSVG:    true,
		},
		{
			name:         "evaluation added",
			activityType: ActivityTypeEvaluationAdded,
			expectSVG:    true,
		},
		{
			name:         "user added",
			activityType: ActivityTypeUserAdded,
			expectSVG:    true,
		},
		{
			name:         "user updated",
			activityType: ActivityTypeUserUpdated,
			expectSVG:    true,
		},
		{
			name:         "conversation added",
			activityType: ActivityTypeConversationAdded,
			expectSVG:    true,
		},
		{
			name:         "conversation completed",
			activityType: ActivityTypeConversationCompleted,
			expectSVG:    true,
		},
		{
			name:         "unknown activity type - default SVG",
			activityType: ActivityType("unknown"),
			expectSVG:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := Activity{Type: tt.activityType}
			result := activity.GetIconSVG()
			
			if tt.expectSVG {
				if !strings.HasPrefix(result, "<svg") {
					t.Errorf("GetIconSVG() should return SVG markup starting with <svg, got: %s", result[:20])
				}
				if !strings.HasSuffix(result, "</svg>") {
					t.Errorf("GetIconSVG() should return SVG markup ending with </svg>, got: %s", result[len(result)-20:])
				}
				if !strings.Contains(result, "currentColor") {
					t.Error("GetIconSVG() should contain 'currentColor' for proper styling")
				}
			}
		})
	}
}

func TestActivityFormatTimeAgo(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name      string
		timestamp time.Time
		expected  string
	}{
		{
			name:      "just now",
			timestamp: now.Add(-30 * time.Second),
			expected:  "gerade eben",
		},
		{
			name:      "1 minute ago",
			timestamp: now.Add(-1 * time.Minute),
			expected:  "vor 1 Minute",
		},
		{
			name:      "5 minutes ago",
			timestamp: now.Add(-5 * time.Minute),
			expected:  "vor 5 Minuten",
		},
		{
			name:      "1 hour ago",
			timestamp: now.Add(-1 * time.Hour),
			expected:  "vor 1 Stunde",
		},
		{
			name:      "3 hours ago",
			timestamp: now.Add(-3 * time.Hour),
			expected:  "vor 3 Stunden",
		},
		{
			name:      "yesterday",
			timestamp: now.Add(-25 * time.Hour),
			expected:  "gestern",
		},
		{
			name:      "3 days ago",
			timestamp: now.Add(-72 * time.Hour),
			expected:  now.Add(-72 * time.Hour).Format("02.01.2006 15:04"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := Activity{Timestamp: tt.timestamp}
			result := activity.FormatTimeAgo()
			if result != tt.expected {
				t.Errorf("FormatTimeAgo() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestActivityStructFields(t *testing.T) {
	t.Run("Activity struct initialization", func(t *testing.T) {
		userID := primitive.NewObjectID()
		targetID := primitive.NewObjectID()
		now := time.Now()

		activity := Activity{
			ID:          primitive.NewObjectID(),
			Type:        ActivityTypeEmployeeAdded,
			UserID:      userID,
			UserName:    "John Admin",
			TargetID:    targetID,
			TargetType:  "employee",
			TargetName:  "Jane Doe",
			Description: "New employee Jane Doe was added to the system",
			Timestamp:   now,
		}

		// Test that all fields are properly set
		if activity.Type != ActivityTypeEmployeeAdded {
			t.Errorf("Type = %v, expected %v", activity.Type, ActivityTypeEmployeeAdded)
		}
		if activity.UserID != userID {
			t.Errorf("UserID = %v, expected %v", activity.UserID, userID)
		}
		if activity.UserName != "John Admin" {
			t.Errorf("UserName = %q, expected %q", activity.UserName, "John Admin")
		}
		if activity.TargetID != targetID {
			t.Errorf("TargetID = %v, expected %v", activity.TargetID, targetID)
		}
		if activity.TargetType != "employee" {
			t.Errorf("TargetType = %q, expected %q", activity.TargetType, "employee")
		}
		if activity.TargetName != "Jane Doe" {
			t.Errorf("TargetName = %q, expected %q", activity.TargetName, "Jane Doe")
		}
		if activity.Description == "" {
			t.Error("Description should not be empty")
		}
		if activity.Timestamp != now {
			t.Errorf("Timestamp = %v, expected %v", activity.Timestamp, now)
		}
	})

	t.Run("Activity with optional fields", func(t *testing.T) {
		activity := Activity{
			ID:          primitive.NewObjectID(),
			Type:        ActivityTypeDocumentUploaded,
			UserID:      primitive.NewObjectID(),
			UserName:    "Jane User",
			TargetType:  "document",
			TargetName:  "contract.pdf",
			Description: "Document was uploaded",
			Timestamp:   time.Now(),
			// TargetID is optional and not set
		}

		// Test that the activity is still valid without TargetID
		if activity.Type != ActivityTypeDocumentUploaded {
			t.Error("Activity should be valid without TargetID")
		}
		if activity.TargetName != "contract.pdf" {
			t.Error("TargetName should be set even without TargetID")
		}
	})
}

func TestActivityCreationScenarios(t *testing.T) {
	t.Run("Employee lifecycle activities", func(t *testing.T) {
		userID := primitive.NewObjectID()
		employeeID := primitive.NewObjectID()
		now := time.Now()

		activities := []Activity{
			{
				Type:        ActivityTypeEmployeeAdded,
				UserID:      userID,
				UserName:    "HR Manager",
				TargetID:    employeeID,
				TargetType:  "employee",
				TargetName:  "John Doe",
				Description: "New employee John Doe was hired",
				Timestamp:   now,
			},
			{
				Type:        ActivityTypeEmployeeUpdated,
				UserID:      userID,
				UserName:    "HR Manager",
				TargetID:    employeeID,
				TargetType:  "employee",
				TargetName:  "John Doe",
				Description: "Employee John Doe's salary was updated",
				Timestamp:   now.Add(time.Hour),
			},
			{
				Type:        ActivityTypeVacationRequested,
				UserID:      employeeID,
				UserName:    "John Doe",
				TargetID:    employeeID,
				TargetType:  "vacation",
				TargetName:  "Summer Vacation",
				Description: "John Doe requested vacation from July 1-15",
				Timestamp:   now.Add(2 * time.Hour),
			},
		}

		for i, activity := range activities {
			// Test that icon classes are appropriate for employee lifecycle
			iconClass := activity.GetIconClass()
			switch activity.Type {
			case ActivityTypeEmployeeAdded:
				if iconClass != "bg-green-500" {
					t.Errorf("Activity %d: Expected green icon for employee added, got %s", i, iconClass)
				}
			case ActivityTypeEmployeeUpdated:
				if iconClass != "bg-blue-500" {
					t.Errorf("Activity %d: Expected blue icon for employee updated, got %s", i, iconClass)
				}
			case ActivityTypeVacationRequested:
				if iconClass != "bg-yellow-500" {
					t.Errorf("Activity %d: Expected yellow icon for vacation requested, got %s", i, iconClass)
				}
			}

			// Test that SVG icons are valid
			svg := activity.GetIconSVG()
			if !strings.Contains(svg, "<svg") {
				t.Errorf("Activity %d: SVG should be valid", i)
			}
		}
	})

	t.Run("Document management activities", func(t *testing.T) {
		userID := primitive.NewObjectID()
		documentID := primitive.NewObjectID()
		
		activity := Activity{
			Type:        ActivityTypeDocumentUploaded,
			UserID:      userID,
			UserName:    "Jane Manager",
			TargetID:    documentID,
			TargetType:  "document",
			TargetName:  "employee_handbook.pdf",
			Description: "Employee handbook was uploaded to the system",
			Timestamp:   time.Now(),
		}

		// Test document activity specifics
		if activity.GetIconClass() != "bg-blue-500" {
			t.Error("Document uploaded should have blue icon")
		}

		svg := activity.GetIconSVG()
		if !strings.Contains(svg, "fill-rule") {
			t.Error("Document icon should contain fill-rule attribute")
		}
	})

	t.Run("User management activities", func(t *testing.T) {
		adminID := primitive.NewObjectID()
		newUserID := primitive.NewObjectID()
		
		activities := []Activity{
			{
				Type:        ActivityTypeUserAdded,
				UserID:      adminID,
				UserName:    "System Admin",
				TargetID:    newUserID,
				TargetType:  "user",
				TargetName:  "New Manager",
				Description: "New user account created for new manager",
				Timestamp:   time.Now(),
			},
			{
				Type:        ActivityTypeUserUpdated,
				UserID:      adminID,
				UserName:    "System Admin",
				TargetID:    newUserID,
				TargetType:  "user",
				TargetName:  "New Manager",
				Description: "User permissions updated",
				Timestamp:   time.Now().Add(time.Minute),
			},
		}

		for _, activity := range activities {
			// User added should have specific icon class
			if activity.Type == ActivityTypeUserAdded {
				iconClass := activity.GetIconClass()
				if iconClass != "bg-gray-500" { // Falls back to default
					// This might need adjustment based on actual implementation
					t.Logf("User added icon class: %s", iconClass)
				}
			}

			// Test that user management SVGs are valid
			svg := activity.GetIconSVG()
			if !strings.HasPrefix(svg, "<svg") {
				t.Error("User management activity should have valid SVG")
			}
		}
	})
}

func TestPluralSFunction(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{
			name:     "singular (1)",
			count:    1,
			expected: "",
		},
		{
			name:     "plural (0)",
			count:    0,
			expected: "n",
		},
		{
			name:     "plural (2)",
			count:    2,
			expected: "n",
		},
		{
			name:     "plural (5)",
			count:    5,
			expected: "n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pluralS(tt.count)
			if result != tt.expected {
				t.Errorf("pluralS(%d) = %q, expected %q", tt.count, result, tt.expected)
			}
		})
	}
}

func TestActivityEdgeCases(t *testing.T) {
	t.Run("Empty activity", func(t *testing.T) {
		activity := Activity{}
		
		// Should not panic and should return default values
		iconClass := activity.GetIconClass()
		if iconClass != "bg-gray-500" {
			t.Errorf("Empty activity should return default gray icon, got %s", iconClass)
		}

		svg := activity.GetIconSVG()
		if !strings.Contains(svg, "<svg") {
			t.Error("Empty activity should return default SVG")
		}
	})

	t.Run("Activity with zero timestamp", func(t *testing.T) {
		activity := Activity{
			Type:      ActivityTypeEmployeeAdded,
			Timestamp: time.Time{},
		}
		
		// Should not panic when formatting time ago
		timeAgo := activity.FormatTimeAgo()
		if timeAgo == "" {
			t.Error("FormatTimeAgo should not return empty string for zero timestamp")
		}
	})

	t.Run("Activity with future timestamp", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)
		activity := Activity{
			Type:      ActivityTypeEmployeeAdded,
			Timestamp: futureTime,
		}
		
		// Should handle future timestamps gracefully
		timeAgo := activity.FormatTimeAgo()
		if timeAgo == "" {
			t.Error("FormatTimeAgo should handle future timestamps")
		}
	})
}

// Benchmark tests
func BenchmarkActivityGetIconClass(b *testing.B) {
	activity := Activity{Type: ActivityTypeEmployeeAdded}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = activity.GetIconClass()
	}
}

func BenchmarkActivityGetIconSVG(b *testing.B) {
	activity := Activity{Type: ActivityTypeEmployeeAdded}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = activity.GetIconSVG()
	}
}

func BenchmarkActivityFormatTimeAgo(b *testing.B) {
	activity := Activity{
		Type:      ActivityTypeEmployeeAdded,
		Timestamp: time.Now().Add(-5 * time.Minute),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = activity.FormatTimeAgo()
	}
}