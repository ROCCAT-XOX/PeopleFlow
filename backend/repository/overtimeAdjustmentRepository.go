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
	adjustment.Status = "pending" // Standardstatus

	result, err := r.collection.InsertOne(ctx, adjustment)
	if err != nil {
		return err
	}

	adjustment.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByEmployeeID findet alle Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) FindByEmployeeID(employeeID string) ([]*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(employeeID)
	if err != nil {
		return nil, err
	}

	// Sortiert nach Erstellungsdatum (neueste zuerst)
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

	return adjustments, cursor.Err()
}

// FindPending findet alle ausstehenden Anpassungen
func (r *OvertimeAdjustmentRepository) FindPending() ([]*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

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

	return adjustments, cursor.Err()
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
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

// FindByID findet eine Anpassung anhand ihrer ID
func (r *OvertimeAdjustmentRepository) FindByID(id string) (*model.OvertimeAdjustment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var adjustment model.OvertimeAdjustment
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&adjustment)
	if err != nil {
		return nil, err
	}

	return &adjustment, nil
}

// GetApprovedAdjustmentsByEmployee berechnet die Summe aller genehmigten Anpassungen für einen Mitarbeiter
func (r *OvertimeAdjustmentRepository) GetApprovedAdjustmentsByEmployee(employeeID string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(employeeID)
	if err != nil {
		return 0, err
	}

	pipeline := []bson.M{
		{"$match": bson.M{
			"employeeId": objID,
			"status":     "approved",
		}},
		{"$group": bson.M{
			"_id":        nil,
			"totalHours": bson.M{"$sum": "$hours"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalHours float64 `bson:"totalHours"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.TotalHours, nil
	}

	return 0, nil
}
