package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OvertimeAdjustmentType repräsentiert den Typ einer Überstunden-Anpassung
type OvertimeAdjustmentType string

const (
	OvertimeAdjustmentTypeCorrection OvertimeAdjustmentType = "correction" // Korrektur
	OvertimeAdjustmentTypeManual     OvertimeAdjustmentType = "manual"     // Manuelle Anpassung
	OvertimeAdjustmentTypeBonus      OvertimeAdjustmentType = "bonus"      // Bonus/Ausgleich
	OvertimeAdjustmentTypePenalty    OvertimeAdjustmentType = "penalty"    // Abzug
)

// OvertimeAdjustment repräsentiert eine manuelle Überstunden-Anpassung
type OvertimeAdjustment struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	EmployeeID   primitive.ObjectID     `bson:"employeeId" json:"employeeId"`
	Type         OvertimeAdjustmentType `bson:"type" json:"type"`
	Hours        float64                `bson:"hours" json:"hours"`             // Kann positiv oder negativ sein
	Reason       string                 `bson:"reason" json:"reason"`           // Kurze Begründung
	Description  string                 `bson:"description" json:"description"` // Detaillierte Beschreibung
	Status       string                 `bson:"status" json:"status"`           // pending, approved, rejected
	AdjustedBy   primitive.ObjectID     `bson:"adjustedBy" json:"adjustedBy"`
	AdjusterName string                 `bson:"adjusterName" json:"adjusterName"`
	ApprovedBy   primitive.ObjectID     `bson:"approvedBy,omitempty" json:"approvedBy,omitempty"`
	ApproverName string                 `bson:"approverName,omitempty" json:"approverName,omitempty"`
	ApprovedAt   time.Time              `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`
	CreatedAt    time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time              `bson:"updatedAt" json:"updatedAt"`
}

// FormatHours formatiert die Stunden zur Anzeige
func (oa *OvertimeAdjustment) FormatHours() string {
	if oa.Hours >= 0 {
		return fmt.Sprintf("+%.1f Std", oa.Hours)
	}
	return fmt.Sprintf("%.1f Std", oa.Hours)
}

// GetTypeDisplayName gibt den deutschen Anzeigenamen für den Anpassungstyp zurück
func (oa *OvertimeAdjustment) GetTypeDisplayName() string {
	switch oa.Type {
	case OvertimeAdjustmentTypeCorrection:
		return "Korrektur"
	case OvertimeAdjustmentTypeManual:
		return "Manuelle Anpassung"
	case OvertimeAdjustmentTypeBonus:
		return "Bonus/Ausgleich"
	case OvertimeAdjustmentTypePenalty:
		return "Abzug"
	default:
		return string(oa.Type)
	}
}

// GetStatusDisplayName gibt den deutschen Anzeigenamen für den Status zurück
func (oa *OvertimeAdjustment) GetStatusDisplayName() string {
	switch oa.Status {
	case "pending":
		return "Ausstehend"
	case "approved":
		return "Genehmigt"
	case "rejected":
		return "Abgelehnt"
	default:
		return oa.Status
	}
}
