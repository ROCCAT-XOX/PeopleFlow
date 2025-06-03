// backend/repository/systemSettingsRepository.go
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

// SystemSettingsRepository enthält alle Datenbankoperationen für System-Einstellungen
type SystemSettingsRepository struct {
	collection *mongo.Collection
}

// NewSystemSettingsRepository erstellt ein neues SystemSettingsRepository
func NewSystemSettingsRepository() *SystemSettingsRepository {
	return &SystemSettingsRepository{
		collection: db.GetCollection("system_settings"),
	}
}

// GetSettings ruft die aktuellen System-Einstellungen ab
func (r *SystemSettingsRepository) GetSettings() (*model.SystemSettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var settings model.SystemSettings
	err := r.collection.FindOne(ctx, bson.M{}).Decode(&settings)

	if err == mongo.ErrNoDocuments {
		// Erstelle Standard-Einstellungen, wenn keine existieren
		defaultSettings := model.DefaultSystemSettings()
		err = r.Create(defaultSettings)
		if err != nil {
			return nil, err
		}
		return defaultSettings, nil
	}

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// Create erstellt neue System-Einstellungen
func (r *SystemSettingsRepository) Create(settings *model.SystemSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, settings)
	if err != nil {
		return err
	}

	settings.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// Update aktualisiert die System-Einstellungen
func (r *SystemSettingsRepository) Update(settings *model.SystemSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	settings.UpdatedAt = time.Now()

	// Wenn es noch keine Einstellungen gibt, erstelle sie
	if settings.ID.IsZero() {
		return r.Create(settings)
	}

	// Wenn keine Einstellungen in der DB existieren, erstelle einen neuen Eintrag
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}

	if count == 0 {
		return r.Create(settings)
	}

	// Andernfalls aktualisiere die erste (und einzige) Einstellung
	filter := bson.M{}
	update := bson.M{"$set": settings}

	opts := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(ctx, filter, update, opts)

	return err
}

// GetCompanyState gibt das eingestellte Bundesland zurück
func (r *SystemSettingsRepository) GetCompanyState() (model.GermanState, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return model.StateNordrheinWestfalen, err // Standard-Fallback
	}

	return model.GermanState(settings.State), nil
}
