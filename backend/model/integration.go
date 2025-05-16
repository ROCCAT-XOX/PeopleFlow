// backend/model/integration.go
package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Integration repräsentiert eine externe Integration/API
type Integration struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"` // z.B. "timebutler", "awork"
	Name      string             `bson:"name" json:"name"` // Anzeigename
	ApiKey    string             `bson:"apiKey" json:"-"`  // API-Schlüssel (nicht in JSON)
	Active    bool               `bson:"active" json:"active"`
	LastSync  time.Time          `bson:"lastSync,omitempty" json:"lastSync,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	Metadata  map[string]string  `bson:"metadata,omitempty" json:"metadata,omitempty"`
	// Synchronisierungseinstellungen
	AutoSync bool `bson:"autoSync" json:"autoSync"` // Automatische Synchronisierung aktiviert
}
