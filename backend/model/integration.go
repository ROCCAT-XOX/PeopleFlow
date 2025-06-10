package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Integration repräsentiert eine externe System-Integration
type Integration struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`         // z.B. "timebutler", "123erfasst", "awork"
	Name      string             `bson:"name" json:"name"`         // Display name
	ApiKey    string             `bson:"apiKey" json:"-"`          // Encrypted API key (never exposed)
	Active    bool               `bson:"active" json:"active"`     // Whether the integration is active
	LastSync  time.Time          `bson:"lastSync" json:"lastSync"` // Last successful sync
	Metadata  map[string]string  `bson:"metadata" json:"metadata"` // Additional configuration
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// IsConfigured prüft, ob die Integration konfiguriert ist
func (i *Integration) IsConfigured() bool {
	return i.ApiKey != "" && i.Active
}

// NeedsSyncSoon prüft, ob die Integration bald synchronisiert werden sollte
func (i *Integration) NeedsSyncSoon(syncInterval time.Duration) bool {
	return time.Since(i.LastSync) > syncInterval
}

// GetSyncStatus gibt den Sync-Status als Text zurück
func (i *Integration) GetSyncStatus() string {
	if !i.Active {
		return "Inaktiv"
	}

	if i.LastSync.IsZero() {
		return "Noch nie synchronisiert"
	}

	duration := time.Since(i.LastSync)
	if duration < time.Hour {
		return "Kürzlich synchronisiert"
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "Vor 1 Stunde synchronisiert"
		}
		return fmt.Sprintf("Vor %d Stunden synchronisiert", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "Vor 1 Tag synchronisiert"
		}
		return fmt.Sprintf("Vor %d Tagen synchronisiert", days)
	}
}

// GetMetadataValue holt einen Wert aus den Metadaten
func (i *Integration) GetMetadataValue(key string) string {
	if i.Metadata == nil {
		return ""
	}
	return i.Metadata[key]
}

// SetMetadataValue setzt einen Wert in den Metadaten
func (i *Integration) SetMetadataValue(key, value string) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]string)
	}
	i.Metadata[key] = value
}

// HasMetadata prüft, ob ein Metadaten-Schlüssel existiert
func (i *Integration) HasMetadata(key string) bool {
	if i.Metadata == nil {
		return false
	}
	_, exists := i.Metadata[key]
	return exists
}
