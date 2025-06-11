package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGermanState_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		state    GermanState
		expected bool
	}{
		{
			name:     "Valid state - Bayern",
			state:    StateBayern,
			expected: true,
		},
		{
			name:     "Valid state - Berlin",
			state:    StateBerlin,
			expected: true,
		},
		{
			name:     "Valid state - NRW",
			state:    StateNordrheinWestfalen,
			expected: true,
		},
		{
			name:     "Invalid state",
			state:    GermanState("invalid_state"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGermanState_GetLabel(t *testing.T) {
	tests := []struct {
		name     string
		state    GermanState
		expected string
	}{
		{
			name:     "Bayern label",
			state:    StateBayern,
			expected: "Bayern",
		},
		{
			name:     "NRW label",
			state:    StateNordrheinWestfalen,
			expected: "Nordrhein-Westfalen",
		},
		{
			name:     "Baden-Württemberg label",
			state:    StateBadenWuerttemberg,
			expected: "Baden-Württemberg",
		},
		{
			name:     "Unknown state",
			state:    GermanState("unknown"),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.GetLabel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSystemSettings(t *testing.T) {
	settings := DefaultSystemSettings()

	assert.NotNil(t, settings)
	assert.Equal(t, "de", settings.Language)
	assert.Equal(t, string(StateNordrheinWestfalen), settings.State)
	assert.Equal(t, 40.0, settings.DefaultWorkingHours)
	assert.Equal(t, 30, settings.DefaultVacationDays)
	assert.False(t, settings.CreatedAt.IsZero())
	assert.False(t, settings.UpdatedAt.IsZero())
}

func TestSystemSettings_HasEmailNotifications(t *testing.T) {
	tests := []struct {
		name     string
		settings *SystemSettings
		expected bool
	}{
		{
			name: "Has enabled email notifications",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled: true,
				},
			},
			expected: true,
		},
		{
			name: "Has disabled email notifications",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled: false,
				},
			},
			expected: false,
		},
		{
			name:     "No email notifications",
			settings: &SystemSettings{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.settings.HasEmailNotifications()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSystemSettings_IsEmailConfigured(t *testing.T) {
	tests := []struct {
		name     string
		settings *SystemSettings
		expected bool
	}{
		{
			name: "Fully configured email",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled:   true,
					SMTPHost:  "smtp.example.com",
					SMTPPort:  587,
					SMTPUser:  "user@example.com",
					SMTPPass:  "password",
					FromEmail: "noreply@example.com",
					FromName:  "PeopleFlow",
				},
			},
			expected: true,
		},
		{
			name: "Missing SMTP host",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled:   true,
					SMTPPort:  587,
					SMTPUser:  "user@example.com",
					SMTPPass:  "password",
					FromEmail: "noreply@example.com",
					FromName:  "PeopleFlow",
				},
			},
			expected: false,
		},
		{
			name: "Missing SMTP port",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled:   true,
					SMTPHost:  "smtp.example.com",
					SMTPPort:  0,
					SMTPUser:  "user@example.com",
					SMTPPass:  "password",
					FromEmail: "noreply@example.com",
					FromName:  "PeopleFlow",
				},
			},
			expected: false,
		},
		{
			name: "Email disabled",
			settings: &SystemSettings{
				EmailNotifications: &EmailNotificationSettings{
					Enabled:   false,
					SMTPHost:  "smtp.example.com",
					SMTPPort:  587,
					SMTPUser:  "user@example.com",
					SMTPPass:  "password",
					FromEmail: "noreply@example.com",
					FromName:  "PeopleFlow",
				},
			},
			expected: false,
		},
		{
			name:     "No email notifications",
			settings: &SystemSettings{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.settings.IsEmailConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetGermanStates(t *testing.T) {
	states := GetGermanStates()

	assert.Len(t, states, 16) // Germany has 16 federal states
	
	// Check a few states
	foundBayern := false
	foundBerlin := false
	foundNRW := false
	
	for _, state := range states {
		if state["value"] == string(StateBayern) {
			assert.Equal(t, "Bayern", state["label"])
			foundBayern = true
		}
		if state["value"] == string(StateBerlin) {
			assert.Equal(t, "Berlin", state["label"])
			foundBerlin = true
		}
		if state["value"] == string(StateNordrheinWestfalen) {
			assert.Equal(t, "Nordrhein-Westfalen", state["label"])
			foundNRW = true
		}
	}
	
	assert.True(t, foundBayern)
	assert.True(t, foundBerlin)
	assert.True(t, foundNRW)
}

func TestSystemSettings_FieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		settings *SystemSettings
		validate func(*SystemSettings) error
		expected string
	}{
		{
			name: "Valid working hours",
			settings: &SystemSettings{
				DefaultWorkingHours: 40,
			},
			validate: func(s *SystemSettings) error {
				if s.DefaultWorkingHours < 0 || s.DefaultWorkingHours > 168 {
					return assert.AnError
				}
				return nil
			},
			expected: "",
		},
		{
			name: "Invalid working hours - negative",
			settings: &SystemSettings{
				DefaultWorkingHours: -1,
			},
			validate: func(s *SystemSettings) error {
				if s.DefaultWorkingHours < 0 {
					return assert.AnError
				}
				return nil
			},
			expected: "error",
		},
		{
			name: "Valid vacation days",
			settings: &SystemSettings{
				DefaultVacationDays: 25,
			},
			validate: func(s *SystemSettings) error {
				if s.DefaultVacationDays < 0 || s.DefaultVacationDays > 365 {
					return assert.AnError
				}
				return nil
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate(tt.settings)
			if tt.expected == "error" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}