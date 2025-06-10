package model

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOvertimeAdjustmentConstants(t *testing.T) {
	t.Run("OvertimeAdjustmentType constants", func(t *testing.T) {
		// Test that all overtime adjustment type constants are defined
		adjustmentTypes := []OvertimeAdjustmentType{
			OvertimeAdjustmentTypeCorrection,
			OvertimeAdjustmentTypeManual,
			OvertimeAdjustmentTypeBonus,
			OvertimeAdjustmentTypePenalty,
		}

		for _, adjustmentType := range adjustmentTypes {
			if string(adjustmentType) == "" {
				t.Errorf("Overtime adjustment type constant should not be empty: %v", adjustmentType)
			}
		}
	})
}

func TestOvertimeAdjustmentFormatHours(t *testing.T) {
	tests := []struct {
		name     string
		hours    float64
		expected string
	}{
		{
			name:     "positive hours",
			hours:    8.5,
			expected: "+8.5 Std",
		},
		{
			name:     "negative hours",
			hours:    -4.0,
			expected: "-4.0 Std",
		},
		{
			name:     "zero hours",
			hours:    0.0,
			expected: "+0.0 Std",
		},
		{
			name:     "fractional positive",
			hours:    2.75,
			expected: "+2.8 Std", // Rounded to 1 decimal place
		},
		{
			name:     "fractional negative",
			hours:    -1.33,
			expected: "-1.3 Std", // Rounded to 1 decimal place
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := OvertimeAdjustment{Hours: tt.hours}
			result := adjustment.FormatHours()
			if result != tt.expected {
				t.Errorf("FormatHours() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestOvertimeAdjustmentGetTypeDisplayName(t *testing.T) {
	tests := []struct {
		name         string
		adjustmentType OvertimeAdjustmentType
		expected     string
	}{
		{
			name:         "correction type",
			adjustmentType: OvertimeAdjustmentTypeCorrection,
			expected:     "Korrektur",
		},
		{
			name:         "manual type",
			adjustmentType: OvertimeAdjustmentTypeManual,
			expected:     "Manuelle Anpassung",
		},
		{
			name:         "bonus type",
			adjustmentType: OvertimeAdjustmentTypeBonus,
			expected:     "Bonus/Ausgleich",
		},
		{
			name:         "penalty type",
			adjustmentType: OvertimeAdjustmentTypePenalty,
			expected:     "Abzug",
		},
		{
			name:         "unknown type",
			adjustmentType: OvertimeAdjustmentType("unknown"),
			expected:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := OvertimeAdjustment{Type: tt.adjustmentType}
			result := adjustment.GetTypeDisplayName()
			if result != tt.expected {
				t.Errorf("GetTypeDisplayName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestOvertimeAdjustmentGetStatusDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "pending status",
			status:   "pending",
			expected: "Ausstehend",
		},
		{
			name:     "approved status",
			status:   "approved",
			expected: "Genehmigt",
		},
		{
			name:     "rejected status",
			status:   "rejected",
			expected: "Abgelehnt",
		},
		{
			name:     "unknown status",
			status:   "unknown",
			expected: "unknown",
		},
		{
			name:     "empty status",
			status:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := OvertimeAdjustment{Status: tt.status}
			result := adjustment.GetStatusDisplayName()
			if result != tt.expected {
				t.Errorf("GetStatusDisplayName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestOvertimeAdjustmentStructFields(t *testing.T) {
	t.Run("Complete overtime adjustment", func(t *testing.T) {
		employeeID := primitive.NewObjectID()
		adjustedByID := primitive.NewObjectID()
		approvedByID := primitive.NewObjectID()
		now := time.Now()
		approvedAt := now.Add(time.Hour)

		adjustment := OvertimeAdjustment{
			ID:           primitive.NewObjectID(),
			EmployeeID:   employeeID,
			Type:         OvertimeAdjustmentTypeCorrection,
			Hours:        5.5,
			Reason:       "Fehlerhaft erfasste Zeiten",
			Description:  "Korrektur der am 15.03. fehlerhaft erfassten Arbeitszeiten",
			Status:       "approved",
			AdjustedBy:   adjustedByID,
			AdjusterName: "HR Manager",
			ApprovedBy:   approvedByID,
			ApproverName: "Department Head",
			ApprovedAt:   approvedAt,
			CreatedAt:    now,
			UpdatedAt:    now.Add(30 * time.Minute),
		}

		// Test that all fields are properly set
		if adjustment.EmployeeID != employeeID {
			t.Errorf("EmployeeID = %v, expected %v", adjustment.EmployeeID, employeeID)
		}
		if adjustment.Type != OvertimeAdjustmentTypeCorrection {
			t.Errorf("Type = %v, expected %v", adjustment.Type, OvertimeAdjustmentTypeCorrection)
		}
		if adjustment.Hours != 5.5 {
			t.Errorf("Hours = %v, expected %v", adjustment.Hours, 5.5)
		}
		if adjustment.Status != "approved" {
			t.Errorf("Status = %q, expected %q", adjustment.Status, "approved")
		}
		if adjustment.AdjustedBy != adjustedByID {
			t.Errorf("AdjustedBy = %v, expected %v", adjustment.AdjustedBy, adjustedByID)
		}
		if adjustment.ApprovedBy != approvedByID {
			t.Errorf("ApprovedBy = %v, expected %v", adjustment.ApprovedBy, approvedByID)
		}
		if adjustment.ApprovedAt != approvedAt {
			t.Errorf("ApprovedAt = %v, expected %v", adjustment.ApprovedAt, approvedAt)
		}
	})

	t.Run("Pending overtime adjustment", func(t *testing.T) {
		employeeID := primitive.NewObjectID()
		adjustedByID := primitive.NewObjectID()
		now := time.Now()

		adjustment := OvertimeAdjustment{
			ID:           primitive.NewObjectID(),
			EmployeeID:   employeeID,
			Type:         OvertimeAdjustmentTypeManual,
			Hours:        -2.0,
			Reason:       "Abzug für Privatnutzung",
			Description:  "Abzug für private Internetnutzung während der Arbeitszeit",
			Status:       "pending",
			AdjustedBy:   adjustedByID,
			AdjusterName: "Direct Supervisor",
			CreatedAt:    now,
			UpdatedAt:    now,
			// ApprovedBy and ApprovedAt should be zero values for pending
		}

		// Test pending adjustment specifics
		if adjustment.Status != "pending" {
			t.Errorf("Status should be pending, got %q", adjustment.Status)
		}
		if !adjustment.ApprovedBy.IsZero() {
			t.Error("ApprovedBy should be zero for pending adjustment")
		}
		if !adjustment.ApprovedAt.IsZero() {
			t.Error("ApprovedAt should be zero for pending adjustment")
		}
		if adjustment.Hours >= 0 {
			t.Error("This test case should have negative hours")
		}
	})
}

func TestOvertimeAdjustmentLifecycle(t *testing.T) {
	t.Run("Adjustment lifecycle - creation to approval", func(t *testing.T) {
		employeeID := primitive.NewObjectID()
		hrManagerID := primitive.NewObjectID()
		departmentHeadID := primitive.NewObjectID()
		
		// 1. Initial creation
		adjustment := OvertimeAdjustment{
			ID:           primitive.NewObjectID(),
			EmployeeID:   employeeID,
			Type:         OvertimeAdjustmentTypeBonus,
			Hours:        8.0,
			Reason:       "Wochenendarbeit",
			Description:  "Zusätzliche Stunden für kritisches Projekt am Wochenende",
			Status:       "pending",
			AdjustedBy:   hrManagerID,
			AdjusterName: "HR Manager",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Test initial state
		if adjustment.GetStatusDisplayName() != "Ausstehend" {
			t.Error("Initial status should be 'Ausstehend'")
		}
		if adjustment.GetTypeDisplayName() != "Bonus/Ausgleich" {
			t.Error("Type should be 'Bonus/Ausgleich'")
		}
		if adjustment.FormatHours() != "+8.0 Std" {
			t.Error("Hours should format as '+8.0 Std'")
		}

		// 2. Approval
		adjustment.Status = "approved"
		adjustment.ApprovedBy = departmentHeadID
		adjustment.ApproverName = "Department Head"
		adjustment.ApprovedAt = time.Now()
		adjustment.UpdatedAt = time.Now()

		// Test approved state
		if adjustment.GetStatusDisplayName() != "Genehmigt" {
			t.Error("Approved status should be 'Genehmigt'")
		}
		if adjustment.ApprovedBy.IsZero() {
			t.Error("ApprovedBy should be set after approval")
		}
		if adjustment.ApprovedAt.IsZero() {
			t.Error("ApprovedAt should be set after approval")
		}
	})

	t.Run("Adjustment lifecycle - creation to rejection", func(t *testing.T) {
		adjustment := OvertimeAdjustment{
			Type:   OvertimeAdjustmentTypePenalty,
			Hours:  -4.0,
			Status: "rejected",
		}

		if adjustment.GetStatusDisplayName() != "Abgelehnt" {
			t.Error("Rejected status should be 'Abgelehnt'")
		}
		if adjustment.GetTypeDisplayName() != "Abzug" {
			t.Error("Penalty type should be 'Abzug'")
		}
		if adjustment.FormatHours() != "-4.0 Std" {
			t.Error("Negative hours should format as '-4.0 Std'")
		}
	})
}

func TestOvertimeAdjustmentEdgeCases(t *testing.T) {
	t.Run("Zero hours adjustment", func(t *testing.T) {
		adjustment := OvertimeAdjustment{
			Hours: 0.0,
			Type:  OvertimeAdjustmentTypeCorrection,
		}

		formatted := adjustment.FormatHours()
		if formatted != "+0.0 Std" {
			t.Errorf("Zero hours should format as '+0.0 Std', got %q", formatted)
		}
	})

	t.Run("Very large positive adjustment", func(t *testing.T) {
		adjustment := OvertimeAdjustment{
			Hours: 99.9,
		}

		formatted := adjustment.FormatHours()
		if formatted != "+99.9 Std" {
			t.Errorf("Large positive hours should format correctly, got %q", formatted)
		}
	})

	t.Run("Very large negative adjustment", func(t *testing.T) {
		adjustment := OvertimeAdjustment{
			Hours: -50.5,
		}

		formatted := adjustment.FormatHours()
		if formatted != "-50.5 Std" {
			t.Errorf("Large negative hours should format correctly, got %q", formatted)
		}
	})

	t.Run("Empty adjustment", func(t *testing.T) {
		adjustment := OvertimeAdjustment{}

		// Should not panic and should return reasonable defaults
		if adjustment.GetTypeDisplayName() != "" {
			t.Error("Empty type should return empty string")
		}
		if adjustment.GetStatusDisplayName() != "" {
			t.Error("Empty status should return empty string")
		}
		if adjustment.FormatHours() != "+0.0 Std" {
			t.Error("Zero hours should format as '+0.0 Std'")
		}
	})
}

func TestOvertimeAdjustmentScenarios(t *testing.T) {
	t.Run("Correction scenario", func(t *testing.T) {
		// Scenario: Employee forgot to clock out, leading to incorrect overtime calculation
		adjustment := OvertimeAdjustment{
			Type:        OvertimeAdjustmentTypeCorrection,
			Hours:       -3.5,
			Reason:      "Vergessen auszustempeln",
			Description: "Mitarbeiter hat vergessen auszustempeln, dadurch wurden 3.5h zu viel berechnet",
			Status:      "approved",
		}

		if adjustment.GetTypeDisplayName() != "Korrektur" {
			t.Error("Should be correction type")
		}
		if adjustment.FormatHours() != "-3.5 Std" {
			t.Error("Should format negative hours correctly")
		}
	})

	t.Run("Bonus scenario", func(t *testing.T) {
		// Scenario: Employee worked during holiday
		adjustment := OvertimeAdjustment{
			Type:        OvertimeAdjustmentTypeBonus,
			Hours:       8.0,
			Reason:      "Feiertagsarbeit",
			Description: "Zusätzliche Vergütung für Arbeit an einem Feiertag",
			Status:      "pending",
		}

		if adjustment.GetTypeDisplayName() != "Bonus/Ausgleich" {
			t.Error("Should be bonus type")
		}
		if adjustment.GetStatusDisplayName() != "Ausstehend" {
			t.Error("Should be pending status")
		}
	})

	t.Run("Penalty scenario", func(t *testing.T) {
		// Scenario: Disciplinary action for excessive breaks
		adjustment := OvertimeAdjustment{
			Type:        OvertimeAdjustmentTypePenalty,
			Hours:       -1.0,
			Reason:      "Überlange Pausenzeiten",
			Description: "Abzug für nicht eingehaltene Pausenzeiten",
			Status:      "approved",
		}

		if adjustment.GetTypeDisplayName() != "Abzug" {
			t.Error("Should be penalty type")
		}
		if adjustment.Hours >= 0 {
			t.Error("Penalty should have negative hours")
		}
	})

	t.Run("Manual adjustment scenario", func(t *testing.T) {
		// Scenario: System migration data correction
		adjustment := OvertimeAdjustment{
			Type:        OvertimeAdjustmentTypeManual,
			Hours:       12.5,
			Reason:      "Datenkorrektur nach Migration",
			Description: "Manuelle Korrektur der Überstunden nach Systemumstellung",
			Status:      "approved",
		}

		if adjustment.GetTypeDisplayName() != "Manuelle Anpassung" {
			t.Error("Should be manual adjustment type")
		}
	})
}

// Benchmark tests
func BenchmarkOvertimeAdjustmentFormatHours(b *testing.B) {
	adjustment := OvertimeAdjustment{Hours: 8.75}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adjustment.FormatHours()
	}
}

func BenchmarkOvertimeAdjustmentGetTypeDisplayName(b *testing.B) {
	adjustment := OvertimeAdjustment{Type: OvertimeAdjustmentTypeCorrection}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adjustment.GetTypeDisplayName()
	}
}

func BenchmarkOvertimeAdjustmentGetStatusDisplayName(b *testing.B) {
	adjustment := OvertimeAdjustment{Status: "approved"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adjustment.GetStatusDisplayName()
	}
}