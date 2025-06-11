package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestIntegration_IsConfigured(t *testing.T) {
	tests := []struct {
		name        string
		integration *Integration
		expected    bool
	}{
		{
			name: "Configured integration",
			integration: &Integration{
				Type:   "timebutler",
				ApiKey: "test-api-key",
				Active: true,
			},
			expected: true,
		},
		{
			name: "Missing API key",
			integration: &Integration{
				Type:   "timebutler",
				ApiKey: "",
				Active: true,
			},
			expected: false,
		},
		{
			name: "Inactive integration",
			integration: &Integration{
				Type:   "timebutler",
				ApiKey: "test-api-key",
				Active: false,
			},
			expected: false,
		},
		{
			name: "Missing API key and inactive",
			integration: &Integration{
				Type:   "timebutler",
				ApiKey: "",
				Active: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.integration.IsConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_NeedsSyncSoon(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		integration  *Integration
		syncInterval time.Duration
		expected     bool
	}{
		{
			name: "Needs sync - old sync",
			integration: &Integration{
				LastSync: now.Add(-2 * time.Hour),
			},
			syncInterval: time.Hour,
			expected:     true,
		},
		{
			name: "No sync needed - recent sync",
			integration: &Integration{
				LastSync: now.Add(-30 * time.Minute),
			},
			syncInterval: time.Hour,
			expected:     false,
		},
		{
			name: "Needs sync - never synced",
			integration: &Integration{
				LastSync: time.Time{},
			},
			syncInterval: time.Hour,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.integration.NeedsSyncSoon(tt.syncInterval)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_GetSyncStatus(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		integration *Integration
		expected    string
	}{
		{
			name: "Inactive integration",
			integration: &Integration{
				Active:   false,
				LastSync: now.Add(-1 * time.Hour),
			},
			expected: "Inaktiv",
		},
		{
			name: "Never synchronized",
			integration: &Integration{
				Active:   true,
				LastSync: time.Time{},
			},
			expected: "Noch nie synchronisiert",
		},
		{
			name: "Recently synchronized",
			integration: &Integration{
				Active:   true,
				LastSync: now.Add(-30 * time.Minute),
			},
			expected: "Kürzlich synchronisiert",
		},
		{
			name: "Synchronized 1 hour ago",
			integration: &Integration{
				Active:   true,
				LastSync: now.Add(-1 * time.Hour),
			},
			expected: "Vor 1 Stunde synchronisiert",
		},
		{
			name: "Synchronized 3 hours ago",
			integration: &Integration{
				Active:   true,
				LastSync: now.Add(-3 * time.Hour),
			},
			expected: "Vor 3 Stunden synchronisiert",
		},
		{
			name: "Synchronized 1 day ago",
			integration: &Integration{
				Active:   true,
				LastSync: now.Add(-24 * time.Hour),
			},
			expected: "Vor 1 Tag synchronisiert",
		},
		{
			name: "Synchronized 3 days ago",
			integration: &Integration{
				Active:   true,
				LastSync: now.Add(-3 * 24 * time.Hour),
			},
			expected: "Vor 3 Tagen synchronisiert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.integration.GetSyncStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_GetMetadataValue(t *testing.T) {
	tests := []struct {
		name        string
		integration *Integration
		key         string
		expected    string
	}{
		{
			name: "Get existing metadata value",
			integration: &Integration{
				Metadata: map[string]string{
					"endpoint": "https://api.example.com",
					"version":  "v1",
				},
			},
			key:      "endpoint",
			expected: "https://api.example.com",
		},
		{
			name: "Get non-existing metadata value",
			integration: &Integration{
				Metadata: map[string]string{
					"endpoint": "https://api.example.com",
				},
			},
			key:      "version",
			expected: "",
		},
		{
			name: "Get from nil metadata",
			integration: &Integration{
				Metadata: nil,
			},
			key:      "endpoint",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.integration.GetMetadataValue(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_SetMetadataValue(t *testing.T) {
	tests := []struct {
		name        string
		integration *Integration
		key         string
		value       string
		validate    func(*testing.T, *Integration)
	}{
		{
			name: "Set metadata on existing map",
			integration: &Integration{
				Metadata: map[string]string{
					"existing": "value",
				},
			},
			key:   "new_key",
			value: "new_value",
			validate: func(t *testing.T, i *Integration) {
				assert.Equal(t, "new_value", i.Metadata["new_key"])
				assert.Equal(t, "value", i.Metadata["existing"])
			},
		},
		{
			name: "Set metadata on nil map",
			integration: &Integration{
				Metadata: nil,
			},
			key:   "new_key",
			value: "new_value",
			validate: func(t *testing.T, i *Integration) {
				assert.NotNil(t, i.Metadata)
				assert.Equal(t, "new_value", i.Metadata["new_key"])
			},
		},
		{
			name: "Overwrite existing metadata",
			integration: &Integration{
				Metadata: map[string]string{
					"key": "old_value",
				},
			},
			key:   "key",
			value: "new_value",
			validate: func(t *testing.T, i *Integration) {
				assert.Equal(t, "new_value", i.Metadata["key"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.integration.SetMetadataValue(tt.key, tt.value)
			tt.validate(t, tt.integration)
		})
	}
}

func TestIntegration_HasMetadata(t *testing.T) {
	tests := []struct {
		name        string
		integration *Integration
		key         string
		expected    bool
	}{
		{
			name: "Has existing metadata",
			integration: &Integration{
				Metadata: map[string]string{
					"endpoint": "https://api.example.com",
					"version":  "v1",
				},
			},
			key:      "endpoint",
			expected: true,
		},
		{
			name: "Does not have metadata",
			integration: &Integration{
				Metadata: map[string]string{
					"endpoint": "https://api.example.com",
				},
			},
			key:      "version",
			expected: false,
		},
		{
			name: "Check with nil metadata",
			integration: &Integration{
				Metadata: nil,
			},
			key:      "endpoint",
			expected: false,
		},
		{
			name: "Has metadata with empty value",
			integration: &Integration{
				Metadata: map[string]string{
					"endpoint": "",
				},
			},
			key:      "endpoint",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.integration.HasMetadata(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntegration_FieldValidation(t *testing.T) {
	tests := []struct {
		name        string
		integration *Integration
		validate    func(*Integration) error
		expectError bool
	}{
		{
			name: "Valid integration",
			integration: &Integration{
				ID:        primitive.NewObjectID(),
				Type:      "timebutler",
				Name:      "Timebutler Integration",
				ApiKey:    "valid-api-key",
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			validate: func(i *Integration) error {
				if i.Type == "" {
					return assert.AnError
				}
				if i.Name == "" {
					return assert.AnError
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "Missing type",
			integration: &Integration{
				ID:     primitive.NewObjectID(),
				Type:   "",
				Name:   "Integration",
				ApiKey: "api-key",
				Active: true,
			},
			validate: func(i *Integration) error {
				if i.Type == "" {
					return assert.AnError
				}
				return nil
			},
			expectError: true,
		},
		{
			name: "Missing name",
			integration: &Integration{
				ID:     primitive.NewObjectID(),
				Type:   "timebutler",
				Name:   "",
				ApiKey: "api-key",
				Active: true,
			},
			validate: func(i *Integration) error {
				if i.Name == "" {
					return assert.AnError
				}
				return nil
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate(tt.integration)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}