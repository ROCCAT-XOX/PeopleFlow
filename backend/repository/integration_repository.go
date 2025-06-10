// backend/repository/integrationRepository.go
package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IntegrationRepository errors
var (
	ErrIntegrationNotFound    = errors.New("integration not found")
	ErrInvalidIntegrationType = errors.New("invalid integration type")
	ErrInvalidApiKey          = errors.New("invalid API key")
	ErrEncryptionFailed       = errors.New("failed to encrypt data")
	ErrDecryptionFailed       = errors.New("failed to decrypt data")
	ErrInvalidMetadata        = errors.New("invalid metadata")
)

// Supported integration types
const (
	IntegrationTypeTimebutler = "timebutler"
	IntegrationType123Erfasst = "123erfasst"
	IntegrationTypeAwork      = "awork"
)

// IntegrationRepository enthält Datenbankoperationen für Integrationen
type IntegrationRepository struct {
	*BaseRepository
	collection *mongo.Collection
}

// NewIntegrationRepository erstellt ein neues IntegrationRepository
func NewIntegrationRepository() *IntegrationRepository {
	collection := db.GetCollection("integrations")
	return &IntegrationRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
	}
}

// ValidateIntegrationType validates the integration type
func (r *IntegrationRepository) ValidateIntegrationType(integrationType string) error {
	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	validTypes := map[string]bool{
		IntegrationTypeTimebutler: true,
		IntegrationType123Erfasst: true,
		IntegrationTypeAwork:      true,
	}

	if !validTypes[integrationType] {
		return fmt.Errorf("%w: %s", ErrInvalidIntegrationType, integrationType)
	}

	return nil
}

// SaveApiKey speichert einen API-Schlüssel für eine Integration (verschlüsselt)
func (r *IntegrationRepository) SaveApiKey(integrationType string, apiKey string) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return ErrInvalidApiKey
	}

	// Encrypt API key
	encryptedKey, err := utils.EncryptString(apiKey)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	// Use upsert to create or update
	filter := bson.M{"type": integrationType}
	update := bson.M{
		"$set": bson.M{
			"type":      integrationType,
			"name":      getIntegrationName(integrationType),
			"apiKey":    encryptedKey,
			"active":    true,
			"updatedAt": time.Now(),
		},
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
			"metadata":  make(map[string]string),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = r.UpdateOne(filter, update, opts)

	return err
}

// GetApiKey holt einen API-Schlüssel für eine Integration und entschlüsselt ihn
func (r *IntegrationRepository) GetApiKey(integrationType string) (string, error) {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return "", err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	var integration model.Integration
	err := r.FindOne(bson.M{"type": integrationType}, &integration)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrIntegrationNotFound
		}
		return "", err
	}

	// Check if integration is active
	if !integration.Active {
		return "", fmt.Errorf("integration %s is not active", integrationType)
	}

	// Decrypt API key
	apiKey, err := utils.DecryptString(integration.ApiKey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return apiKey, nil
}

// GetIntegrationStatus prüft, ob eine Integration aktiv ist
func (r *IntegrationRepository) GetIntegrationStatus(integrationType string) (bool, error) {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return false, err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	var integration model.Integration
	err := r.FindOne(bson.M{"type": integrationType}, &integration)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	return integration.Active, nil
}

// SetIntegrationStatus setzt den Status einer Integration
func (r *IntegrationRepository) SetIntegrationStatus(integrationType string, active bool) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	update := bson.M{
		"$set": bson.M{
			"active":    active,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"type": integrationType}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrIntegrationNotFound
	}

	return nil
}

// GetIntegration holt eine vollständige Integration
func (r *IntegrationRepository) GetIntegration(integrationType string) (*model.Integration, error) {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return nil, err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	var integration model.Integration
	err := r.FindOne(bson.M{"type": integrationType}, &integration)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrIntegrationNotFound
		}
		return nil, err
	}

	return &integration, nil
}

// GetAllIntegrations holt alle Integrationen
func (r *IntegrationRepository) GetAllIntegrations() ([]model.Integration, error) {
	var integrations []model.Integration

	err := r.FindAll(bson.M{}, &integrations, options.Find().SetSort(bson.M{"name": 1}))
	if err != nil {
		return nil, err
	}

	return integrations, nil
}

// GetActiveIntegrations holt alle aktiven Integrationen
func (r *IntegrationRepository) GetActiveIntegrations() ([]model.Integration, error) {
	var integrations []model.Integration

	filter := bson.M{"active": true}
	err := r.FindAll(filter, &integrations, options.Find().SetSort(bson.M{"name": 1}))
	if err != nil {
		return nil, err
	}

	return integrations, nil
}

// SetMetadata speichert Metadaten für eine Integration
func (r *IntegrationRepository) SetMetadata(integrationType string, key string, value string) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))
	key = strings.TrimSpace(key)

	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrInvalidMetadata)
	}

	// Use transaction to ensure atomic update
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Check if integration exists
		var integration model.Integration
		err := r.collection.FindOne(sessCtx, bson.M{"type": integrationType}).Decode(&integration)

		if err == mongo.ErrNoDocuments {
			// Create new integration if it doesn't exist
			integration = model.Integration{
				Type:      integrationType,
				Name:      getIntegrationName(integrationType),
				Active:    false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Metadata:  map[string]string{key: value},
			}
			_, err = r.collection.InsertOne(sessCtx, integration)
			return err
		} else if err != nil {
			return err
		}

		// Update existing integration
		updateField := fmt.Sprintf("metadata.%s", key)
		update := bson.M{
			"$set": bson.M{
				updateField: value,
				"updatedAt": time.Now(),
			},
		}

		_, err = r.collection.UpdateOne(sessCtx, bson.M{"type": integrationType}, update)
		return err
	})
}

// GetMetadata holt Metadaten für eine Integration
func (r *IntegrationRepository) GetMetadata(integrationType string, key string) (string, error) {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return "", err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))
	key = strings.TrimSpace(key)

	if key == "" {
		return "", fmt.Errorf("%w: key cannot be empty", ErrInvalidMetadata)
	}

	integration, err := r.GetIntegration(integrationType)
	if err != nil {
		return "", err
	}

	if integration.Metadata == nil {
		return "", nil
	}

	return integration.Metadata[key], nil
}

// GetAllMetadata holt alle Metadaten für eine Integration
func (r *IntegrationRepository) GetAllMetadata(integrationType string) (map[string]string, error) {
	integration, err := r.GetIntegration(integrationType)
	if err != nil {
		return nil, err
	}

	if integration.Metadata == nil {
		return make(map[string]string), nil
	}

	return integration.Metadata, nil
}

// DeleteMetadata löscht einen Metadaten-Eintrag
func (r *IntegrationRepository) DeleteMetadata(integrationType string, key string) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))
	key = strings.TrimSpace(key)

	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrInvalidMetadata)
	}

	updateField := fmt.Sprintf("metadata.%s", key)
	update := bson.M{
		"$unset": bson.M{
			updateField: "",
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"type": integrationType}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrIntegrationNotFound
	}

	return nil
}

// SetLastSync setzt den Zeitstempel der letzten Synchronisierung
func (r *IntegrationRepository) SetLastSync(integrationType string, lastSync time.Time) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	update := bson.M{
		"$set": bson.M{
			"lastSync":  lastSync,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"type": integrationType}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrIntegrationNotFound
	}

	return nil
}

// GetLastSync holt den Zeitstempel der letzten Synchronisierung
func (r *IntegrationRepository) GetLastSync(integrationType string) (time.Time, error) {
	integration, err := r.GetIntegration(integrationType)
	if err != nil {
		return time.Time{}, err
	}

	return integration.LastSync, nil
}

// DeleteIntegration löscht eine Integration komplett
func (r *IntegrationRepository) DeleteIntegration(integrationType string) error {
	// Validate input
	if err := r.ValidateIntegrationType(integrationType); err != nil {
		return err
	}

	integrationType = strings.ToLower(strings.TrimSpace(integrationType))

	err := r.DeleteOne(bson.M{"type": integrationType})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrIntegrationNotFound
		}
		return err
	}

	return nil
}

// CreateIndexes erstellt erforderliche Indizes
func (r *IntegrationRepository) CreateIndexes() error {
	// Unique index on integration type
	if err := r.CreateIndex(bson.M{"type": 1}, true); err != nil {
		return fmt.Errorf("failed to create type index: %w", err)
	}

	// Index on active status for queries
	if err := r.CreateIndex(bson.M{"active": 1}, false); err != nil {
		return fmt.Errorf("failed to create active index: %w", err)
	}

	// Index on lastSync for sync queries
	if err := r.CreateIndex(bson.M{"lastSync": 1}, false); err != nil {
		return fmt.Errorf("failed to create lastSync index: %w", err)
	}

	return nil
}

// Helper function to get the display name for an integration type
func getIntegrationName(integrationType string) string {
	switch integrationType {
	case IntegrationTypeTimebutler:
		return "Timebutler"
	case IntegrationType123Erfasst:
		return "123erfasst"
	case IntegrationTypeAwork:
		return "AWork"
	default:
		// Capitalize first letter
		if len(integrationType) > 0 {
			return strings.ToUpper(string(integrationType[0])) + integrationType[1:]
		}
		return integrationType
	}
}
