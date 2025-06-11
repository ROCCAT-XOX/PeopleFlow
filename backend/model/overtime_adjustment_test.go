package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOvertimeAdjustmentType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		adjType  OvertimeAdjustmentType
		expected bool
	}{
		{
			name:     "Valid - Manual",
			adjType:  OvertimeAdjustmentTypeManual,
			expected: true,
		},
		{
			name:     "Valid - Correction",
			adjType:  OvertimeAdjustmentTypeCorrection,
			expected: true,
		},
		{
			name:     "Valid - CarryOver",
			adjType:  OvertimeAdjustmentTypeCarryOver,
			expected: true,
		},
		{
			name:     "Valid - Payout",
			adjType:  OvertimeAdjustmentTypePayout,
			expected: true,
		},
		{
			name:     "Invalid - Unknown",
			adjType:  OvertimeAdjustmentType("unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjType.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustmentType_GetLabel(t *testing.T) {
	tests := []struct {
		name     string
		adjType  OvertimeAdjustmentType
		expected string
	}{
		{
			name:     "Manual",
			adjType:  OvertimeAdjustmentTypeManual,
			expected: "Manuelle Anpassung",
		},
		{
			name:     "Correction",
			adjType:  OvertimeAdjustmentTypeCorrection,
			expected: "Korrektur",
		},
		{
			name:     "CarryOver",
			adjType:  OvertimeAdjustmentTypeCarryOver,
			expected: "Übertrag",
		},
		{
			name:     "Payout",
			adjType:  OvertimeAdjustmentTypePayout,
			expected: "Auszahlung",
		},
		{
			name:     "Unknown",
			adjType:  OvertimeAdjustmentType("unknown"),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjType.GetLabel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_IsApproved(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   bool
	}{
		{
			name: "Approved adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "approved",
			},
			expected: true,
		},
		{
			name: "Pending adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "pending",
			},
			expected: false,
		},
		{
			name: "Rejected adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "rejected",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.IsApproved()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_IsPending(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   bool
	}{
		{
			name: "Pending adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "pending",
			},
			expected: true,
		},
		{
			name: "Approved adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "approved",
			},
			expected: false,
		},
		{
			name: "Rejected adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "rejected",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.IsPending()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_IsRejected(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   bool
	}{
		{
			name: "Rejected adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "rejected",
			},
			expected: true,
		},
		{
			name: "Approved adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "approved",
			},
			expected: false,
		},
		{
			name: "Pending adjustment",
			adjustment: &OvertimeAdjustment{
				Status: "pending",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.IsRejected()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_GetTypeDisplayName(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   string
	}{
		{
			name: "Manual adjustment",
			adjustment: &OvertimeAdjustment{
				Type: OvertimeAdjustmentTypeManual,
			},
			expected: "Manuelle Anpassung",
		},
		{
			name: "Correction adjustment",
			adjustment: &OvertimeAdjustment{
				Type: OvertimeAdjustmentTypeCorrection,
			},
			expected: "Korrektur",
		},
		{
			name: "CarryOver adjustment",
			adjustment: &OvertimeAdjustment{
				Type: OvertimeAdjustmentTypeCarryOver,
			},
			expected: "Übertrag",
		},
		{
			name: "Payout adjustment",
			adjustment: &OvertimeAdjustment{
				Type: OvertimeAdjustmentTypePayout,
			},
			expected: "Auszahlung",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.GetTypeDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_FormatHours(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   string
	}{
		{
			name: "Positive hours",
			adjustment: &OvertimeAdjustment{
				Hours: 5.5,
			},
			expected: "+5.50 Stunden",
		},
		{
			name: "Negative hours",
			adjustment: &OvertimeAdjustment{
				Hours: -2.25,
			},
			expected: "-2.25 Stunden",
		},
		{
			name: "Zero hours",
			adjustment: &OvertimeAdjustment{
				Hours: 0,
			},
			expected: "+0.00 Stunden",
		},
		{
			name: "Large positive hours",
			adjustment: &OvertimeAdjustment{
				Hours: 40.75,
			},
			expected: "+40.75 Stunden",
		},
		{
			name: "Large negative hours",
			adjustment: &OvertimeAdjustment{
				Hours: -15.33,
			},
			expected: "-15.33 Stunden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.FormatHours()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_Description(t *testing.T) {
	tests := []struct {
		name       string
		adjustment *OvertimeAdjustment
		expected   string
	}{
		{
			name: "Has reason",
			adjustment: &OvertimeAdjustment{
				Reason: "Overtime for project deadline",
			},
			expected: "Overtime for project deadline",
		},
		{
			name: "Empty reason",
			adjustment: &OvertimeAdjustment{
				Reason: "",
			},
			expected: "",
		},
		{
			name: "Long reason",
			adjustment: &OvertimeAdjustment{
				Reason: "This is a very long reason explaining why the overtime adjustment was made including detailed context and background information",
			},
			expected: "This is a very long reason explaining why the overtime adjustment was made including detailed context and background information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.adjustment.Description()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOvertimeAdjustment_Validation(t *testing.T) {
	tests := []struct {
		name        string
		adjustment  *OvertimeAdjustment
		validate    func(*OvertimeAdjustment) error
		expectError bool
	}{
		{
			name: "Valid adjustment",
			adjustment: &OvertimeAdjustment{
				ID:           primitive.NewObjectID(),
				EmployeeID:   primitive.NewObjectID(),
				Type:         OvertimeAdjustmentTypeManual,
				Hours:        5.5,
				Reason:       "Extra work for project",
				Status:       "pending",
				AdjustedBy:   primitive.NewObjectID(),
				AdjusterName: "Manager Name",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			validate: func(oa *OvertimeAdjustment) error {
				if !oa.Type.IsValid() {
					return assert.AnError
				}
				if oa.EmployeeID.IsZero() {
					return assert.AnError
				}
				if oa.AdjustedBy.IsZero() {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Invalid type",
			adjustment: &OvertimeAdjustment{
				Type: OvertimeAdjustmentType("invalid"),
			},
			validate: func(oa *OvertimeAdjustment) error {
				if !oa.Type.IsValid() {
					return assert.AnError
				}
				return nil
			},
			expectError: true,
		},
		{
			name: "Missing employee ID",
			adjustment: &OvertimeAdjustment{
				Type:       OvertimeAdjustmentTypeManual,
				EmployeeID: primitive.ObjectID{},
			},
			validate: func(oa *OvertimeAdjustment) error {
				if oa.EmployeeID.IsZero() {
					return assert.AnError
				}
				return nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate(tt.adjustment)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}