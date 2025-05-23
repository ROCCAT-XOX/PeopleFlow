package repository

import (
	"context"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// IntegrationRepository enthält Datenbankoperationen für Integrationen
type IntegrationRepository struct {
	collection *mongo.Collection
}

// NewIntegrationRepository erstellt ein neues IntegrationRepository
func NewIntegrationRepository() *IntegrationRepository {
	return &IntegrationRepository{
		collection: db.GetCollection("integrations"),
	}
}

// SaveApiKey speichert einen API-Schlüssel für eine Integration (verschlüsselt)
func (r *IntegrationRepository) SaveApiKey(integrationType string, apiKey string) error {
	// API-Schlüssel verschlüsseln
	encryptedKey, err := utils.EncryptString(apiKey)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob bereits ein Eintrag existiert
	var existing model.Integration
	err = r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		// Neuen Eintrag erstellen
		integration := model.Integration{
			Type:      integrationType,
			Name:      getIntegrationName(integrationType),
			ApiKey:    encryptedKey, // Verschlüsselter Schlüssel
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]string),
		}
		_, err = r.collection.InsertOne(ctx, integration)
		return err
	} else if err != nil {
		return err
	}

	// Bestehenden Eintrag aktualisieren
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"type": integrationType},
		bson.M{
			"$set": bson.M{
				"apiKey":    encryptedKey, // Verschlüsselter Schlüssel
				"active":    true,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

// GetApiKey holt einen API-Schlüssel für eine Integration und entschlüsselt ihn
func (r *IntegrationRepository) GetApiKey(integrationType string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var integration model.Integration
	err := r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&integration)
	if err != nil {
		return "", err
	}

	// Entschlüsseln des API-Schlüssels
	apiKey, err := utils.DecryptString(integration.ApiKey)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

// GetIntegrationStatus prüft, ob eine Integration aktiv ist
func (r *IntegrationRepository) GetIntegrationStatus(integrationType string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var integration model.Integration
	err := r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&integration)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return integration.Active, nil
}

// SetIntegrationStatus setzt den Status einer Integration
func (r *IntegrationRepository) SetIntegrationStatus(integrationType string, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"type": integrationType},
		bson.M{
			"$set": bson.M{
				"active":    active,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

// GetAllIntegrations holt alle Integrationen
func (r *IntegrationRepository) GetAllIntegrations() ([]model.Integration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var integrations []model.Integration
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var integration model.Integration
		if err := cursor.Decode(&integration); err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return integrations, nil
}

// SetMetadata speichert Metadaten für eine Integration
func (r *IntegrationRepository) SetMetadata(integrationType string, key string, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prüfen, ob Integration existiert
	var integration model.Integration
	err := r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&integration)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Neue Integration erstellen, falls sie nicht existiert
			integration := model.Integration{
				Type:      integrationType,
				Name:      getIntegrationName(integrationType),
				Active:    false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Metadata:  map[string]string{key: value},
			}
			_, err = r.collection.InsertOne(ctx, integration)
			return err
		}
		return err
	}

	// Update the metadata using the $set operator with dot notation for the specific field
	updateField := "metadata." + key
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"type": integrationType},
		bson.M{
			"$set": bson.M{
				updateField: value,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

// GetMetadata holt Metadaten für eine Integration
func (r *IntegrationRepository) GetMetadata(integrationType string, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var integration model.Integration
	err := r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&integration)
	if err != nil {
		return "", err
	}

	// Check if metadata and the specific key exist
	if integration.Metadata == nil {
		return "", nil
	}

	return integration.Metadata[key], nil
}

// SetLastSync setzt den Zeitstempel der letzten Synchronisierung
func (r *IntegrationRepository) SetLastSync(integrationType string, lastSync time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"type": integrationType},
		bson.M{
			"$set": bson.M{
				"lastSync":  lastSync,
				"updatedAt": time.Now(),
			},
		},
	)
	return err
}

// GetLastSync holt den Zeitstempel der letzten Synchronisierung
func (r *IntegrationRepository) GetLastSync(integrationType string) (time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var integration model.Integration
	err := r.collection.FindOne(ctx, bson.M{"type": integrationType}).Decode(&integration)
	if err != nil {
		return time.Time{}, err
	}

	return integration.LastSync, nil
}

// Helper to get the display name for an integration type
func getIntegrationName(integrationType string) string {
	switch integrationType {
	case "timebutler":
		return "Timebutler"
	case "123erfasst":
		return "123erfasst"
	case "awork":
		return "AWork"
	default:
		return integrationType
	}
}
