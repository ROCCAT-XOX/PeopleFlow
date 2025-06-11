package repository

import (
	"errors"
	"fmt"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OvertimeAdjustmentRepository errors
var (
	ErrAdjustmentNotFound    = errors.New("overtime adjustment not found")
	ErrInvalidAdjustmentType = errors.New("invalid adjustment type")
	ErrInvalidAdjustmentData = errors.New("invalid adjustment data")
	ErrInvalidStatus         = errors.New("invalid adjustment status")
	ErrAlreadyProcessed      = errors.New("adjustment already processed")
)

// Valid adjustment statuses
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// OvertimeAdjustmentRepository enthält alle Datenbankoperationen für Überstunden-Anpassungen
type OvertimeAdjustmentRepository struct {
	*BaseRepository
	collection *mongo.Collection
}

// NewOvertimeAdjustmentRepository erstellt ein neues OvertimeAdjustmentRepository
func NewOvertimeAdjustmentRepository() *OvertimeAdjustmentRepository {
	collection := db.GetCollection("overtime_adjustments")
	return &OvertimeAdjustmentRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
	}
}

// ValidateAdjustment validates adjustment data
func (r *OvertimeAdjustmentRepository) ValidateAdjustment(adjustment *model.OvertimeAdjustment) error {
	// Validate adjustment type
	validTypes := map[model.OvertimeAdjustmentType]bool{
		model.OvertimeAdjustmentTypeManual:     true,
		model.OvertimeAdjustmentTypeCorrection: true,
		model.OvertimeAdjustmentTypeCarryOver:  true,
		model.OvertimeAdjustmentTypePayout:     true,
	}

	if !validTypes[adjustment.Type] {
		return fmt.Errorf("%w: %s", ErrInvalidAdjustmentType, adjustment.Type)
	}

	// Validate employee ID
	if adjustment.EmployeeID.IsZero() {
		return fmt.Errorf("%w: employee ID is required", ErrInvalidAdjustmentData)
	}

	// Validate hours
	if adjustment.Hours == 0 {
		return fmt.Errorf("%w: hours cannot be zero", ErrInvalidAdjustmentData)
	}

	// Validate hours range (-100 to 100)
	if adjustment.Hours < -100 || adjustment.Hours > 100 {
		return fmt.Errorf("%w: hours must be between -100 and 100", ErrInvalidAdjustmentData)
	}

	// Validate reason
	if adjustment.Reason == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidAdjustmentData)
	}

	// Validate adjuster information
	if adjustment.AdjustedBy.IsZero() {
		return fmt.Errorf("%w: adjuster ID is required", ErrInvalidAdjustmentData)
	}

	if adjustment.AdjusterName == "" {
		return fmt.Errorf("%w: adjuster name is required", ErrInvalidAdjustmentData)
	}

	// Validate status
	if adjustment.Status != "" {
		validStatuses := map[string]bool{
			StatusPending:  true,
			StatusApproved: true,
			StatusRejected: true,
		}

		if !validStatuses[adjustment.Status] {
			return fmt.Errorf("%w: %s", ErrInvalidStatus, adjustment.Status)
		}
	}

	return nil
}

// Create erstellt eine neue Überstunden-Anpassung
func (r *OvertimeAdjustmentRepository) Create(adjustment *model.OvertimeAdjustment) error {
	// Validate adjustment
	if err := r.ValidateAdjustment(adjustment); err != nil {
		return err
	}

	// Set default status if not provided
	if adjustment.Status == "" {
		adjustment.Status = StatusPending
	}

	// Set timestamps
	adjustment.CreatedAt = time.Now()
	adjustment.UpdatedAt = time.Now()

	id, err := r.InsertOne(adjustment)
	if err != nil {
		return fmt.Errorf("failed to create adjustment: %w", err)
	}

	adjustment.ID = *id
	return nil
}

// FindByID findet eine Anpassung anhand der ID
func (r *OvertimeAdjustmentRepository) FindByID(id string) (*model.OvertimeAdjustment, error) {
	var adjustment model.OvertimeAdjustment
	err := r.BaseRepository.FindByID(id, &adjustment)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrAdjustmentNotFound
		}
		return nil, err
	}
	return &adjustment, nil
}

// FindByEmployeeID findet alle Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) FindByEmployeeID(employeeID string, skip, limit int64) ([]*model.OvertimeAdjustment, int64, error) {
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return nil, 0, err
	}

	var adjustments []*model.OvertimeAdjustment

	// Options for sorting and pagination
	findOptions := options.Find().
		SetSort(bson.M{"createdAt": -1}). // Newest first
		SetSkip(skip).
		SetLimit(limit)

	filter := bson.M{"employeeId": objID}
	err = r.FindAll(filter, &adjustments, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return adjustments, total, nil
}

// FindPending findet alle ausstehenden Anpassungen
func (r *OvertimeAdjustmentRepository) FindPending(skip, limit int64) ([]*model.OvertimeAdjustment, int64, error) {
	var adjustments []*model.OvertimeAdjustment

	// Options for sorting and pagination
	findOptions := options.Find().
		SetSort(bson.M{"createdAt": 1}). // Oldest first for processing
		SetSkip(skip).
		SetLimit(limit)

	filter := bson.M{"status": StatusPending}
	err := r.FindAll(filter, &adjustments, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return adjustments, total, nil
}

// FindApprovedByEmployeeID findet alle genehmigten Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) FindApprovedByEmployeeID(employeeID string) ([]*model.OvertimeAdjustment, error) {
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return nil, err
	}

	var adjustments []*model.OvertimeAdjustment

	filter := bson.M{
		"employeeId": objID,
		"status":     StatusApproved,
	}

	// Sort by creation date
	findOptions := options.Find().SetSort(bson.M{"createdAt": -1})

	err = r.FindAll(filter, &adjustments, findOptions)
	if err != nil {
		return nil, err
	}

	return adjustments, nil
}

// FindByDateRange findet Anpassungen in einem Zeitraum
func (r *OvertimeAdjustmentRepository) FindByDateRange(start, end time.Time, status string) ([]*model.OvertimeAdjustment, error) {
	var adjustments []*model.OvertimeAdjustment

	filter := bson.M{
		"createdAt": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	// Add status filter if provided
	if status != "" {
		filter["status"] = status
	}

	findOptions := options.Find().SetSort(bson.M{"createdAt": -1})

	err := r.FindAll(filter, &adjustments, findOptions)
	if err != nil {
		return nil, err
	}

	return adjustments, nil
}

// UpdateStatus aktualisiert den Status einer Anpassung mit Validierung
func (r *OvertimeAdjustmentRepository) UpdateStatus(adjustmentID string, status string, approverID primitive.ObjectID, approverName string) error {
	// Validate status
	validStatuses := map[string]bool{
		StatusApproved: true,
		StatusRejected: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}

	objID, err := r.ValidateObjectID(adjustmentID)
	if err != nil {
		return err
	}

	// Use transaction to ensure atomicity
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// First check if adjustment exists and is still pending
		var adjustment model.OvertimeAdjustment
		err := r.collection.FindOne(sessCtx, bson.M{"_id": objID}).Decode(&adjustment)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrAdjustmentNotFound
			}
			return err
		}

		// Check if already processed
		if adjustment.Status != StatusPending {
			return fmt.Errorf("%w: current status is %s", ErrAlreadyProcessed, adjustment.Status)
		}

		// Update status
		update := bson.M{
			"$set": bson.M{
				"status":       status,
				"approvedBy":   approverID,
				"approverName": approverName,
				"approvedAt":   time.Now(),
				"updatedAt":    time.Now(),
			},
		}

		_, err = r.collection.UpdateOne(sessCtx, bson.M{"_id": objID}, update)
		return err
	})
}

// Update aktualisiert eine Anpassung (nur wenn noch pending)
func (r *OvertimeAdjustmentRepository) Update(adjustment *model.OvertimeAdjustment) error {
	// Validate adjustment
	if err := r.ValidateAdjustment(adjustment); err != nil {
		return err
	}

	// Use transaction to check status before update
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Check current status
		var current model.OvertimeAdjustment
		err := r.collection.FindOne(sessCtx, bson.M{"_id": adjustment.ID}).Decode(&current)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrAdjustmentNotFound
			}
			return err
		}

		// Only allow updates if still pending
		if current.Status != StatusPending {
			return fmt.Errorf("%w: cannot update adjustment with status %s", ErrAlreadyProcessed, current.Status)
		}

		adjustment.UpdatedAt = time.Now()

		_, err = r.collection.UpdateOne(
			sessCtx,
			bson.M{"_id": adjustment.ID},
			bson.M{"$set": adjustment},
		)
		return err
	})
}

// Delete löscht eine Anpassung (nur wenn noch pending)
func (r *OvertimeAdjustmentRepository) Delete(id string) error {
	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return err
	}

	// Use transaction to check status before deletion
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Check current status
		var adjustment model.OvertimeAdjustment
		err := r.collection.FindOne(sessCtx, bson.M{"_id": objID}).Decode(&adjustment)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrAdjustmentNotFound
			}
			return err
		}

		// Only allow deletion if still pending
		if adjustment.Status != StatusPending {
			return fmt.Errorf("%w: cannot delete adjustment with status %s", ErrAlreadyProcessed, adjustment.Status)
		}

		_, err = r.collection.DeleteOne(sessCtx, bson.M{"_id": objID})
		return err
	})
}

// GetSummaryByEmployee gibt eine Zusammenfassung der Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) GetSummaryByEmployee(employeeID string) (*model.OvertimeAdjustmentSummary, error) {
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := r.GetContext()
	defer cancel()

	// Aggregation pipeline
	pipeline := []bson.M{
		{
			"$match": bson.M{"employeeId": objID},
		},
		{
			"$group": bson.M{
				"_id":        "$status",
				"count":      bson.M{"$sum": 1},
				"totalHours": bson.M{"$sum": "$hours"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get adjustment summary: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Build summary
	summary := &model.OvertimeAdjustmentSummary{
		EmployeeID:    employeeID,
		TotalPending:  0,
		TotalApproved: 0,
		TotalRejected: 0,
		HoursPending:  0,
		HoursApproved: 0,
		HoursRejected: 0,
	}

	for _, result := range results {
		status := result["_id"].(string)
		count := int(result["count"].(int32))
		hours := result["totalHours"].(float64)

		switch status {
		case StatusPending:
			summary.TotalPending = count
			summary.HoursPending = hours
		case StatusApproved:
			summary.TotalApproved = count
			summary.HoursApproved = hours
		case StatusRejected:
			summary.TotalRejected = count
			summary.HoursRejected = hours
		}
	}

	return summary, nil
}

// BulkUpdateStatus updates multiple adjustments' status
func (r *OvertimeAdjustmentRepository) BulkUpdateStatus(adjustmentIDs []string, status string, approverID primitive.ObjectID, approverName string) error {
	// Validate status
	validStatuses := map[string]bool{
		StatusApproved: true,
		StatusRejected: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}

	// Convert string IDs to ObjectIDs
	var objectIDs []primitive.ObjectID
	for _, id := range adjustmentIDs {
		objID, err := r.ValidateObjectID(id)
		if err != nil {
			return fmt.Errorf("invalid adjustment ID %s: %w", id, err)
		}
		objectIDs = append(objectIDs, *objID)
	}

	// Use transaction for bulk update
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Update all adjustments
		update := bson.M{
			"$set": bson.M{
				"status":       status,
				"approvedBy":   approverID,
				"approverName": approverName,
				"approvedAt":   time.Now(),
				"updatedAt":    time.Now(),
			},
		}

		_, err := r.collection.UpdateMany(
			sessCtx,
			bson.M{
				"_id":    bson.M{"$in": objectIDs},
				"status": StatusPending, // Only update pending adjustments
			},
			update,
		)

		return err
	})
}

// CreateIndexes erstellt erforderliche Indizes
func (r *OvertimeAdjustmentRepository) CreateIndexes() error {
	// Index on employeeId for employee queries
	if err := r.CreateIndex(bson.M{"employeeId": 1}, false); err != nil {
		return fmt.Errorf("failed to create employeeId index: %w", err)
	}

	// Index on status for pending queries
	if err := r.CreateIndex(bson.M{"status": 1}, false); err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Compound index for employee and status
	if err := r.CreateIndex(bson.M{"employeeId": 1, "status": 1}, false); err != nil {
		return fmt.Errorf("failed to create employeeId-status index: %w", err)
	}

	// Index on createdAt for sorting
	if err := r.CreateIndex(bson.M{"createdAt": -1}, false); err != nil {
		return fmt.Errorf("failed to create createdAt index: %w", err)
	}

	// Index on approvedBy for approver queries
	if err := r.CreateIndex(bson.M{"approvedBy": 1}, false); err != nil {
		return fmt.Errorf("failed to create approvedBy index: %w", err)
	}

	return nil
}
