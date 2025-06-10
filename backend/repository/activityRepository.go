// backend/repository/activityRepository.go
package repository

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActivityRepository errors
var (
	ErrActivityNotFound    = errors.New("activity not found")
	ErrInvalidActivityType = errors.New("invalid activity type")
	ErrInvalidActivityData = errors.New("invalid activity data")
)

// ActivityRepository enthält alle Datenbankoperationen für das Activity-Modell
type ActivityRepository struct {
	*BaseRepository
	collection *mongo.Collection
}

// NewActivityRepository erstellt ein neues ActivityRepository
func NewActivityRepository() *ActivityRepository {
	collection := db.GetCollection("activities")
	return &ActivityRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
	}
}

// ValidateActivity validates activity data
func (r *ActivityRepository) ValidateActivity(activity *model.Activity) error {
	// Validate activity type
	validTypes := map[model.ActivityType]bool{
		model.ActivityTypeEmployeeAdded:        true,
		model.ActivityTypeEmployeeUpdated:      true,
		model.ActivityTypeEmployeeDeleted:      true,
		model.ActivityTypeVacationRequested:    true,
		model.ActivityTypeVacationApproved:     true,
		model.ActivityTypeVacationRejected:     true,
		model.ActivityTypeOvertimeAdjusted:     true,
		model.ActivityTypeDocumentUploaded:     true,
		model.ActivityTypeSystemSettingChanged: true,
	}

	if !validTypes[activity.Type] {
		return fmt.Errorf("%w: %s", ErrInvalidActivityType, activity.Type)
	}

	// Validate required fields
	if activity.UserID.IsZero() {
		return fmt.Errorf("%w: user ID is required", ErrInvalidActivityData)
	}

	if activity.UserName == "" {
		return fmt.Errorf("%w: user name is required", ErrInvalidActivityData)
	}

	// Target validation depends on activity type
	requiresTarget := map[model.ActivityType]bool{
		model.ActivityTypeEmployeeAdded:     true,
		model.ActivityTypeEmployeeUpdated:   true,
		model.ActivityTypeEmployeeDeleted:   true,
		model.ActivityTypeVacationRequested: true,
		model.ActivityTypeVacationApproved:  true,
		model.ActivityTypeVacationRejected:  true,
		model.ActivityTypeOvertimeAdjusted:  true,
		model.ActivityTypeDocumentUploaded:  true,
	}

	if requiresTarget[activity.Type] {
		if activity.TargetID.IsZero() {
			return fmt.Errorf("%w: target ID is required for activity type %s", ErrInvalidActivityData, activity.Type)
		}
		if activity.TargetType == "" {
			return fmt.Errorf("%w: target type is required for activity type %s", ErrInvalidActivityData, activity.Type)
		}
		if activity.TargetName == "" {
			return fmt.Errorf("%w: target name is required for activity type %s", ErrInvalidActivityData, activity.Type)
		}
	}

	return nil
}

// Create erstellt eine neue Aktivität mit Validierung
func (r *ActivityRepository) Create(activity *model.Activity) error {
	// Validate activity
	if err := r.ValidateActivity(activity); err != nil {
		return err
	}

	// Set timestamp if not provided
	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	id, err := r.InsertOne(activity)
	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	activity.ID = *id
	return nil
}

// FindByID findet eine Aktivität anhand ihrer ID
func (r *ActivityRepository) FindByID(id string) (*model.Activity, error) {
	var activity model.Activity
	err := r.BaseRepository.FindByID(id, &activity)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}
	return &activity, nil
}

// FindByUser findet alle Aktivitäten eines Benutzers
func (r *ActivityRepository) FindByUser(userID string, limit int64) ([]*model.Activity, error) {
	objID, err := r.ValidateObjectID(userID)
	if err != nil {
		return nil, err
	}

	var activities []*model.Activity

	findOptions := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	filter := bson.M{"userId": objID}
	err = r.FindAll(filter, &activities, findOptions)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

// FindByTarget findet alle Aktivitäten für ein bestimmtes Ziel
func (r *ActivityRepository) FindByTarget(targetID string, limit int64) ([]*model.Activity, error) {
	objID, err := r.ValidateObjectID(targetID)
	if err != nil {
		return nil, err
	}

	var activities []*model.Activity

	findOptions := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	filter := bson.M{"targetId": objID}
	err = r.FindAll(filter, &activities, findOptions)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

// FindByType findet alle Aktivitäten eines bestimmten Typs
func (r *ActivityRepository) FindByType(activityType model.ActivityType, limit int64) ([]*model.Activity, error) {
	var activities []*model.Activity

	findOptions := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(limit)

	filter := bson.M{"type": activityType}
	err = r.FindAll(filter, &activities, findOptions)
	if err != nil {
		return nil, err
	}

	return activities, nil
}

// FindByDateRange findet Aktivitäten in einem Zeitraum
func (r *ActivityRepository) FindByDateRange(start, end time.Time, skip, limit int64) ([]*model.Activity, int64, error) {
	var activities []*model.Activity

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	findOptions := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetSkip(skip).
		SetLimit(limit)

	err := r.FindAll(filter, &activities, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// Get total count for date range
	total, err := r.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// LogActivity fügt eine neue Aktivität hinzu und gibt die erstellte Aktivität zurück
func (r *ActivityRepository) LogActivity(
	activityType model.ActivityType,
	userID primitive.ObjectID,
	userName string,
	targetID primitive.ObjectID,
	targetType, targetName, description string,
) (*model.Activity, error) {

	activity := &model.Activity{
		Type:        activityType,
		UserID:      userID,
		UserName:    userName,
		TargetID:    targetID,
		TargetType:  targetType,
		TargetName:  targetName,
		Description: description,
		Timestamp:   time.Now(),
	}

	if err := r.Create(activity); err != nil {
		return nil, err
	}

	return activity, nil
}

// DeleteOldActivities löscht Aktivitäten, die älter als die angegebene Dauer sind
func (r *ActivityRepository) DeleteOldActivities(olderThan time.Duration) (int64, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	cutoffDate := time.Now().Add(-olderThan)

	result, err := r.collection.DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": cutoffDate},
	})

	if err != nil {
		return 0, fmt.Errorf("failed to delete old activities: %w", err)
	}

	return result.DeletedCount, nil
}

// GetActivityStats gibt Statistiken über Aktivitäten zurück
func (r *ActivityRepository) GetActivityStats(days int) (map[string]interface{}, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	startDate := time.Now().AddDate(0, 0, -days)

	// Aggregation pipeline
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"timestamp": bson.M{"$gte": startDate},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert to map
	stats := make(map[string]interface{})
	stats["period_days"] = days
	stats["start_date"] = startDate
	stats["total_activities"] = 0

	typeStats := make(map[string]int)
	totalCount := 0

	for _, result := range results {
		activityType := result["_id"].(string)
		count := int(result["count"].(int32))
		typeStats[activityType] = count
		totalCount += count
	}

	stats["total_activities"] = totalCount
	stats["by_type"] = typeStats

	return stats, nil
}

// CreateIndexes erstellt erforderliche Indizes
func (r *ActivityRepository) CreateIndexes() error {
	// Index on timestamp for sorting
	if err := r.CreateIndex(bson.M{"timestamp": -1}, false); err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	// Index on userId for user queries
	if err := r.CreateIndex(bson.M{"userId": 1}, false); err != nil {
		return fmt.Errorf("failed to create userId index: %w", err)
	}

	// Index on targetId for target queries
	if err := r.CreateIndex(bson.M{"targetId": 1}, false); err != nil {
		return fmt.Errorf("failed to create targetId index: %w", err)
	}

	// Compound index for type and timestamp
	if err := r.CreateIndex(bson.M{"type": 1, "timestamp": -1}, false); err != nil {
		return fmt.Errorf("failed to create type-timestamp index: %w", err)
	}

	// TTL index to automatically delete old activities after 90 days
	// This can be adjusted based on requirements
	ctx, cancel := r.GetContext()
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.M{"timestamp": 1},
		Options: options.Index().
			SetExpireAfterSeconds(90 * 24 * 60 * 60), // 90 days
	}

	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create TTL index: %w", err)
	}

	return nil
}

// FindRecent findet die neuesten Aktivitäten (ohne Pagination, mit Limit)
func (r *ActivityRepository) FindRecent(limit int) ([]*model.Activity, error) {
	var activities []*model.Activity

	findOptions := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetLimit(int64(limit))

	err := r.BaseRepository.FindAll(bson.M{}, &activities, findOptions)
	if err != nil {
		return nil, err
	}

	return activities, nil
}
