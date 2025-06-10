// backend/repository/systemSettingsRepository.go
package repository

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemSettingsRepository errors
var (
	ErrInvalidSystemSettings       = errors.New("invalid system settings")
	ErrInvalidState                = errors.New("invalid German state")
	ErrInvalidNotificationSettings = errors.New("invalid notification settings")
)

// SystemSettingsRepository enthält alle Datenbankoperationen für System-Einstellungen
type SystemSettingsRepository struct {
	*BaseRepository
	collection *mongo.Collection
	mu         sync.RWMutex
	cache      *model.SystemSettings
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// Singleton instance
var (
	settingsRepoInstance *SystemSettingsRepository
	settingsRepoOnce     sync.Once
)

// NewSystemSettingsRepository erstellt ein neues SystemSettingsRepository (Singleton)
func NewSystemSettingsRepository() *SystemSettingsRepository {
	settingsRepoOnce.Do(func() {
		collection := db.GetCollection("system_settings")
		settingsRepoInstance = &SystemSettingsRepository{
			BaseRepository: NewBaseRepository(collection),
			collection:     collection,
			cacheTTL:       5 * time.Minute, // Cache for 5 minutes
		}
	})
	return settingsRepoInstance
}

// ValidateSystemSettings validates system settings
func (r *SystemSettingsRepository) ValidateSystemSettings(settings *model.SystemSettings) error {
	// Validate German state
	if settings.State != "" {
		validStates := map[model.GermanState]bool{
			model.StateBadenWuerttemberg:     true,
			model.StateBayern:                true,
			model.StateBerlin:                true,
			model.StateBrandenburg:           true,
			model.StateBremen:                true,
			model.StateHamburg:               true,
			model.StateHessen:                true,
			model.StateMecklenburgVorpommern: true,
			model.StateNiedersachsen:         true,
			model.StateNordrheinWestfalen:    true,
			model.StateRheinlandPfalz:        true,
			model.StateSaarland:              true,
			model.StateSachsen:               true,
			model.StateSachsenAnhalt:         true,
			model.StateSchleswigHolstein:     true,
			model.StateThueringen:            true,
		}

		if !validStates[model.GermanState(settings.State)] {
			return fmt.Errorf("%w: %s", ErrInvalidState, settings.State)
		}
	}

	// Validate email notifications
	if settings.EmailNotifications != nil {
		if settings.EmailNotifications.SMTPHost != "" && settings.EmailNotifications.SMTPPort <= 0 {
			return fmt.Errorf("%w: SMTP port must be positive", ErrInvalidNotificationSettings)
		}
	}

	// Validate working hours
	if settings.DefaultWorkingHours < 0 || settings.DefaultWorkingHours > 60 {
		return fmt.Errorf("%w: default working hours must be between 0 and 60", ErrInvalidSystemSettings)
	}

	// Validate vacation days
	if settings.DefaultVacationDays < 0 || settings.DefaultVacationDays > 365 {
		return fmt.Errorf("%w: default vacation days must be between 0 and 365", ErrInvalidSystemSettings)
	}

	return nil
}

// GetSettings ruft die aktuellen System-Einstellungen ab (mit Cache)
func (r *SystemSettingsRepository) GetSettings() (*model.SystemSettings, error) {
	// Check cache first
	r.mu.RLock()
	if r.cache != nil && time.Since(r.cacheTime) < r.cacheTTL {
		cached := *r.cache // Return a copy
		r.mu.RUnlock()
		return &cached, nil
	}
	r.mu.RUnlock()

	// Cache miss or expired, fetch from database
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.cache != nil && time.Since(r.cacheTime) < r.cacheTTL {
		cached := *r.cache
		return &cached, nil
	}

	var settings model.SystemSettings
	err := r.FindOne(bson.M{}, &settings)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// Create default settings if none exist
			defaultSettings := model.DefaultSystemSettings()
			if err := r.createWithoutCache(defaultSettings); err != nil {
				return nil, err
			}
			r.updateCache(defaultSettings)
			return defaultSettings, nil
		}
		return nil, err
	}

	r.updateCache(&settings)
	return &settings, nil
}

// Create erstellt neue System-Einstellungen
func (r *SystemSettingsRepository) Create(settings *model.SystemSettings) error {
	// Validate settings
	if err := r.ValidateSystemSettings(settings); err != nil {
		return err
	}

	// Ensure only one settings document exists
	count, err := r.Count(bson.M{})
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%w: system settings already exist", ErrInvalidSystemSettings)
	}

	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	id, err := r.InsertOne(settings)
	if err != nil {
		return err
	}

	settings.ID = *id

	// Update cache
	r.mu.Lock()
	r.updateCache(settings)
	r.mu.Unlock()

	return nil
}

// createWithoutCache creates settings without updating cache (used internally)
func (r *SystemSettingsRepository) createWithoutCache(settings *model.SystemSettings) error {
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	id, err := r.InsertOne(settings)
	if err != nil {
		return err
	}

	settings.ID = *id
	return nil
}

// Update aktualisiert die System-Einstellungen
func (r *SystemSettingsRepository) Update(settings *model.SystemSettings) error {
	// Validate settings
	if err := r.ValidateSystemSettings(settings); err != nil {
		return err
	}

	settings.UpdatedAt = time.Now()

	// Use transaction to ensure atomicity
	err := r.Transaction(func(sessCtx mongo.SessionContext) error {
		// If ID is not set, find the existing settings
		if settings.ID.IsZero() {
			var existing model.SystemSettings
			err := r.collection.FindOne(sessCtx, bson.M{}).Decode(&existing)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// No settings exist, create new
					return r.createWithoutCache(settings)
				}
				return err
			}
			settings.ID = existing.ID
		}

		// Update existing settings
		update := bson.M{
			"$set": settings,
		}

		result, err := r.collection.UpdateOne(
			sessCtx,
			bson.M{"_id": settings.ID},
			update,
			options.Update().SetUpsert(true),
		)

		if err != nil {
			return err
		}

		if result.UpsertedID != nil {
			settings.ID = result.UpsertedID.(primitive.ObjectID)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Update cache
	r.mu.Lock()
	r.updateCache(settings)
	r.mu.Unlock()

	return nil
}

// updateCache updates the internal cache
func (r *SystemSettingsRepository) updateCache(settings *model.SystemSettings) {
	r.cache = &model.SystemSettings{}
	*r.cache = *settings // Deep copy
	r.cacheTime = time.Now()
}

// InvalidateCache invalidates the cache
func (r *SystemSettingsRepository) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = nil
	r.cacheTime = time.Time{}
}

// GetCompanyState gibt das eingestellte Bundesland zurück
func (r *SystemSettingsRepository) GetCompanyState() (model.GermanState, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return model.StateNordrheinWestfalen, err // Default fallback
	}

	if settings.State == "" {
		return model.StateNordrheinWestfalen, nil // Default
	}

	return model.GermanState(settings.State), nil
}

// UpdateCompanyInfo aktualisiert nur die Firmeninformationen
func (r *SystemSettingsRepository) UpdateCompanyInfo(name, address, state string) error {
	// Validate state
	if state != "" {
		tempSettings := &model.SystemSettings{State: state}
		if err := r.ValidateSystemSettings(tempSettings); err != nil {
			return err
		}
	}

	settings, err := r.GetSettings()
	if err != nil {
		return err
	}

	// Update only company info
	settings.CompanyName = name
	settings.CompanyAddress = address
	if state != "" {
		settings.State = state
	}

	return r.Update(settings)
}

// UpdateEmailNotifications aktualisiert nur die E-Mail-Benachrichtigungseinstellungen
func (r *SystemSettingsRepository) UpdateEmailNotifications(notifications *model.EmailNotificationSettings) error {
	if notifications == nil {
		return fmt.Errorf("%w: notifications cannot be nil", ErrInvalidNotificationSettings)
	}

	// Validate notification settings
	tempSettings := &model.SystemSettings{EmailNotifications: notifications}
	if err := r.ValidateSystemSettings(tempSettings); err != nil {
		return err
	}

	settings, err := r.GetSettings()
	if err != nil {
		return err
	}

	settings.EmailNotifications = notifications
	return r.Update(settings)
}

// UpdateWorkDefaults aktualisiert nur die Arbeitszeit-Standardwerte
func (r *SystemSettingsRepository) UpdateWorkDefaults(workingHours float64, vacationDays int) error {
	// Validate
	tempSettings := &model.SystemSettings{
		DefaultWorkingHours: workingHours,
		DefaultVacationDays: vacationDays,
	}
	if err := r.ValidateSystemSettings(tempSettings); err != nil {
		return err
	}

	settings, err := r.GetSettings()
	if err != nil {
		return err
	}

	settings.DefaultWorkingHours = workingHours
	settings.DefaultVacationDays = vacationDays
	return r.Update(settings)
}

// IsEmailNotificationEnabled prüft, ob E-Mail-Benachrichtigungen aktiviert sind
func (r *SystemSettingsRepository) IsEmailNotificationEnabled() (bool, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return false, err
	}

	return settings.EmailNotifications != nil &&
		settings.EmailNotifications.Enabled &&
		settings.EmailNotifications.SMTPHost != "", nil
}

// GetSMTPConfig gibt die SMTP-Konfiguration zurück
func (r *SystemSettingsRepository) GetSMTPConfig() (*model.EmailNotificationSettings, error) {
	settings, err := r.GetSettings()
	if err != nil {
		return nil, err
	}

	if settings.EmailNotifications == nil || !settings.EmailNotifications.Enabled {
		return nil, fmt.Errorf("email notifications are not enabled")
	}

	return settings.EmailNotifications, nil
}

// ResetToDefaults setzt die Einstellungen auf Standardwerte zurück
func (r *SystemSettingsRepository) ResetToDefaults() error {
	defaultSettings := model.DefaultSystemSettings()

	// Get current settings to preserve ID
	current, err := r.GetSettings()
	if err == nil {
		defaultSettings.ID = current.ID
		defaultSettings.CreatedAt = current.CreatedAt
	}

	return r.Update(defaultSettings)
}

// CreateIndexes erstellt erforderliche Indizes
func (r *SystemSettingsRepository) CreateIndexes() error {
	// No specific indexes needed for system settings
	// as there should only be one document
	return nil
}

// EnsureSingleDocument stellt sicher, dass nur ein Einstellungsdokument existiert
func (r *SystemSettingsRepository) EnsureSingleDocument() error {
	ctx, cancel := r.GetContext()
	defer cancel()

	// Count documents
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count settings documents: %w", err)
	}

	if count > 1 {
		// Keep only the most recent document
		var settings []model.SystemSettings
		cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"updatedAt": -1}))
		if err != nil {
			return fmt.Errorf("failed to find settings: %w", err)
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &settings); err != nil {
			return fmt.Errorf("failed to decode settings: %w", err)
		}

		// Delete all but the first (most recent)
		for i := 1; i < len(settings); i++ {
			if _, err := r.collection.DeleteOne(ctx, bson.M{"_id": settings[i].ID}); err != nil {
				return fmt.Errorf("failed to delete duplicate settings: %w", err)
			}
		}

		// Invalidate cache
		r.InvalidateCache()
	}

	return nil
}
