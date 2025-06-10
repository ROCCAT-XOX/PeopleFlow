// backend/model/overtime_adjustment_type.go
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OvertimeAdjustmentType definiert die Art der Überstunden-Anpassung
type OvertimeAdjustmentType string

const (
	OvertimeAdjustmentTypeManual     OvertimeAdjustmentType = "manual"
	OvertimeAdjustmentTypeCorrection OvertimeAdjustmentType = "correction"
	OvertimeAdjustmentTypeCarryOver  OvertimeAdjustmentType = "carryover"
	OvertimeAdjustmentTypePayout     OvertimeAdjustmentType = "payout"
)

// OvertimeAdjustment repräsentiert eine manuelle Anpassung der Überstunden
type OvertimeAdjustment struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	EmployeeID   primitive.ObjectID     `bson:"employeeId" json:"employeeId"`
	Type         OvertimeAdjustmentType `bson:"type" json:"type"`
	Hours        float64                `bson:"hours" json:"hours"`
	Reason       string                 `bson:"reason" json:"reason"`
	Status       string                 `bson:"status" json:"status"` // pending, approved, rejected
	AdjustedBy   primitive.ObjectID     `bson:"adjustedBy" json:"adjustedBy"`
	AdjusterName string                 `bson:"adjusterName" json:"adjusterName"`
	ApprovedBy   primitive.ObjectID     `bson:"approvedBy,omitempty" json:"approvedBy,omitempty"`
	ApproverName string                 `bson:"approverName,omitempty" json:"approverName,omitempty"`
	ApprovedAt   time.Time              `bson:"approvedAt,omitempty" json:"approvedAt,omitempty"`
	CreatedAt    time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time              `bson:"updatedAt" json:"updatedAt"`
}

// IsValid prüft, ob der OvertimeAdjustmentType gültig ist
func (oat OvertimeAdjustmentType) IsValid() bool {
	switch oat {
	case OvertimeAdjustmentTypeManual, OvertimeAdjustmentTypeCorrection,
		OvertimeAdjustmentTypeCarryOver, OvertimeAdjustmentTypePayout:
		return true
	default:
		return false
	}
}

// GetLabel gibt ein benutzerfreundliches Label zurück
func (oat OvertimeAdjustmentType) GetLabel() string {
	switch oat {
	case OvertimeAdjustmentTypeManual:
		return "Manuelle Anpassung"
	case OvertimeAdjustmentTypeCorrection:
		return "Korrektur"
	case OvertimeAdjustmentTypeCarryOver:
		return "Übertrag"
	case OvertimeAdjustmentTypePayout:
		return "Auszahlung"
	default:
		return string(oat)
	}
}

// IsApproved prüft, ob die Anpassung genehmigt wurde
func (oa *OvertimeAdjustment) IsApproved() bool {
	return oa.Status == "approved"
}

// IsPending prüft, ob die Anpassung noch aussteht
func (oa *OvertimeAdjustment) IsPending() bool {
	return oa.Status == "pending"
}

// IsRejected prüft, ob die Anpassung abgelehnt wurde
func (oa *OvertimeAdjustment) IsRejected() bool {
	return oa.Status == "rejected"
}
