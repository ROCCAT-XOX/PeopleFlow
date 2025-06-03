package repository

import (
	"context"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OvertimeAdjustmentRepository enthält alle Datenbankoperationen für Überstunden-Anpassungen
type OvertimeAdjustmentRepository struct {
	collection *mongo.Collection
}

// NewOvertimeAdjustmentRepository erstellt ein neues OvertimeAdjustmentRepository
func NewOvertimeAdjustmentRepository() *OvertimeAdjustmentRepository {
	return &OvertimeAdjustmentRepository{
		collection: db.GetCollection("overtime_adjustments"),
	}
}

// Create erstellt eine neue Überstunden-Anpassung
func (r *OvertimeAdjustmentRepository) Create(adjustment *model.OvertimeAdjustment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adjustment.CreatedAt = time.Now()
	adjustment.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, adjustment)
	if err != nil {
		return err
	}

	adjustment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByID findet eine Anpassung anhand der ID
func (r *OvertimeAdjustmentRepository) FindByID(id string) (*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var adjustment model.OvertimeAdjustment
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&adjustment)
	if err != nil {
		return nil, err
	}

	return &adjustment, nil
}

// FindByEmployeeID findet alle Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) FindByEmployeeID(employeeID string) ([]*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(employeeID)
	if err != nil {
		return nil, err
	}

	// Nach Erstellungsdatum sortieren (neueste zuerst)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	var adjustments []*model.OvertimeAdjustment
	cursor, err := r.collection.Find(ctx, bson.M{"employeeId": objID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var adjustment model.OvertimeAdjustment
		if err := cursor.Decode(&adjustment); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, &adjustment)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return adjustments, nil
}

// FindPending findet alle ausstehenden Anpassungen
func (r *OvertimeAdjustmentRepository) FindPending() ([]*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Nach Erstellungsdatum sortieren (älteste zuerst für Genehmigung)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}})

	var adjustments []*model.OvertimeAdjustment
	cursor, err := r.collection.Find(ctx, bson.M{"status": "pending"}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var adjustment model.OvertimeAdjustment
		if err := cursor.Decode(&adjustment); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, &adjustment)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return adjustments, nil
}

// UpdateStatus aktualisiert den Status einer Anpassung
func (r *OvertimeAdjustmentRepository) UpdateStatus(adjustmentID string, status string, approverID primitive.ObjectID, approverName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(adjustmentID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"approvedBy":   approverID,
			"approverName": approverName,
			"approvedAt":   time.Now(),
			"updatedAt":    time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// Update aktualisiert eine Anpassung
func (r *OvertimeAdjustmentRepository) Update(adjustment *model.OvertimeAdjustment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adjustment.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": adjustment.ID},
		bson.M{"$set": adjustment},
	)
	return err
}

// Delete löscht eine Anpassung
func (r *OvertimeAdjustmentRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

// FindApprovedByEmployeeID findet alle genehmigten Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) FindApprovedByEmployeeID(employeeID string) ([]*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(employeeID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"employeeId": objID,
		"status":     "approved",
	}

	var adjustments []*model.OvertimeAdjustment
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var adjustment model.OvertimeAdjustment
		if err := cursor.Decode(&adjustment); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, &adjustment)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return adjustments, nil
}
