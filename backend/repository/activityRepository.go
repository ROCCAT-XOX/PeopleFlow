package repository

import (
	"context"
	"time"

	"PeoplePilot/backend/db"
	"PeoplePilot/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActivityRepository enthält alle Datenbankoperationen für das Activity-Modell
type ActivityRepository struct {
	collection *mongo.Collection
}

// NewActivityRepository erstellt ein neues ActivityRepository
func NewActivityRepository() *ActivityRepository {
	return &ActivityRepository{
		collection: db.GetCollection("activities"),
	}
}

// Create erstellt eine neue Aktivität
func (r *ActivityRepository) Create(activity *model.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	result, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		return err
	}

	activity.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindRecent findet die neuesten Aktivitäten, begrenzt durch limit
func (r *ActivityRepository) FindRecent(limit int) ([]*model.Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Optionen für die Sortierung und Begrenzung
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}). // Absteigend nach Zeitstempel
		SetLimit(int64(limit))

	var activities []*model.Activity
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var activity model.Activity
		if err := cursor.Decode(&activity); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// LogActivity fügt eine neue Aktivität hinzu und gibt die erstellte Aktivität zurück
func (r *ActivityRepository) LogActivity(activityType model.ActivityType, userId primitive.ObjectID, userName string, targetId primitive.ObjectID, targetType, targetName, description string) (*model.Activity, error) {
	activity := &model.Activity{
		Type:        activityType,
		UserID:      userId,
		UserName:    userName,
		TargetID:    targetId,
		TargetType:  targetType,
		TargetName:  targetName,
		Description: description,
		Timestamp:   time.Now(),
	}

	err := r.Create(activity)
	if err != nil {
		return nil, err
	}

	return activity, nil
}
