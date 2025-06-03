// backend/model/overtime_adjustment.go
package model

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// OvertimeAdjustmentType repräsentiert den Typ einer Überstunden-Anpassung
type OvertimeAdjustmentType string

const (
	AdjustmentTypeCorrection OvertimeAdjustmentType = "correction" // Korrektur wegen Fehler
	AdjustmentTypeManual     OvertimeAdjustmentType = "manual"     // Manuelle Anpassung
	AdjustmentTypeBonus      OvertimeAdjustmentType = "bonus"      // Bonus/Ausgleich
	AdjustmentTypePenalty    OvertimeAdjustmentType = "penalty"    // Abzug
)

// OvertimeAdjustment repräsentiert eine manuelle Überstunden-Anpassung
type OvertimeAdjustment struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	EmployeeID   primitive.ObjectID     `bson:"employeeId" json:"employeeId"`
	Type         OvertimeAdjustmentType `bson:"type" json:"type"`
	Hours        float64                `bson:"hours" json:"hours"`               // Kann positiv oder negativ sein
	Reason       string                 `bson:"reason" json:"reason"`             // Pflichtfeld: Begründung
	Description  string                 `bson:"description" json:"description"`   // Detaillierte Beschreibung
	AdjustedBy   primitive.ObjectID     `bson:"adjustedBy" json:"adjustedBy"`     // Wer die Anpassung vorgenommen hat
	AdjusterName string                 `bson:"adjusterName" json:"adjusterName"` // Name des Bearbeiters
	CreatedAt    time.Time              `bson:"createdAt" json:"createdAt"`
	ApprovedBy   primitive.ObjectID     `bson:"approvedBy,omitempty" json:"approvedBy,omitempty"`     // Optional: Genehmiger
	ApproverName string                 `bson:"approverName,omitempty" json:"approverName,omitempty"` // Name des Genehmigers
	ApprovedAt   time.Time              `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`     // Genehmigungsdatum
	Status       string                 `bson:"status" json:"status"`                                 // pending, approved, rejected
}

// GetDisplayType gibt den deutschen Anzeigenamen für den Anpassungstyp zurück
func (a OvertimeAdjustment) GetDisplayType() string {
	switch a.Type {
	case AdjustmentTypeCorrection:
		return "Korrektur"
	case AdjustmentTypeManual:
		return "Manuelle Anpassung"
	case AdjustmentTypeBonus:
		return "Bonus/Ausgleich"
	case AdjustmentTypePenalty:
		return "Abzug"
	default:
		return string(a.Type)
	}
}

// GetStatusDisplay gibt den deutschen Anzeigenamen für den Status zurück
func (a OvertimeAdjustment) GetStatusDisplay() string {
	switch a.Status {
	case "pending":
		return "Ausstehend"
	case "approved":
		return "Genehmigt"
	case "rejected":
		return "Abgelehnt"
	default:
		return a.Status
	}
}

// GetStatusClass gibt die CSS-Klasse für den Status zurück
func (a OvertimeAdjustment) GetStatusClass() string {
	switch a.Status {
	case "pending":
		return "bg-yellow-100 text-yellow-800"
	case "approved":
		return "bg-green-100 text-green-800"
	case "rejected":
		return "bg-red-100 text-red-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
}

// FormatHours formatiert die Stunden zur Anzeige
func (a OvertimeAdjustment) FormatHours() string {
	if a.Hours >= 0 {
		return fmt.Sprintf("+%.2f Std", a.Hours)
	}
	return fmt.Sprintf("%.2f Std", a.Hours)
}
