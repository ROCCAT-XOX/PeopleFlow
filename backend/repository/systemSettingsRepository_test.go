package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemSettingsRepositoryTestSuite defines the test suite for SystemSettingsRepository
type SystemSettingsRepositoryTestSuite struct {
	suite.Suite
	repo       *SystemSettingsRepository
	collection *mongo.Collection
	client     *mongo.Client
}

// SetupSuite runs once before all tests
func (suite *SystemSettingsRepositoryTestSuite) SetupSuite() {
	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(suite.T(), err)

	suite.client = client
	suite.collection = client.Database("peopleflow_test").Collection("system_settings")

	// Initialize test repository
	db.SetTestCollection("system_settings", suite.collection)

	// Reset singleton for tests
	settingsRepoInstance = nil
	settingsRepoOnce = sync.Once{}

	suite.repo = NewSystemSettingsRepository()
}

// SetupTest runs before each test
func (suite *SystemSettingsRepositoryTestSuite) SetupTest() {
	// Clear collection before each test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := suite.collection.Drop(ctx)
	if err != nil && err != mongo.ErrNilDocument {
		suite.T().Fatal(err)
	}

	// Clear cache
	suite.repo.InvalidateCache()
}

// TearDownSuite runs once after all tests
func (suite *SystemSettingsRepositoryTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if suite.client != nil {
		suite.client.Disconnect(ctx)
	}
}

// Test GetSettings - Creates Default Settings
func (suite *SystemSettingsRepositoryTestSuite) TestGetSettings_CreatesDefault() {
	// First call should create default settings
	settings, err := suite.repo.GetSettings()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), settings)
	assert.NotEmpty(suite.T(), settings.ID)
	assert.Equal(suite.T(), model.StateNordrheinWestfalen, model.GermanState(settings.State))
	assert.Equal(suite.T(), float64(40), settings.DefaultWorkingHours)
	assert.Equal(suite.T(), 30, settings.DefaultVacationDays)
}

// Test GetSettings - Cache
func (suite *SystemSettingsRepositoryTestSuite) TestGetSettings_Cache() {
	// First call loads from database
	settings1, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Modify database directly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = suite.collection.UpdateOne(
		ctx,
		bson.M{"_id": settings1.ID},
		bson.M{"$set": bson.M{"companyName": "Modified Company"}},
	)
	require.NoError(suite.T(), err)

	// Second call should return cached value
	settings2, err := suite.repo.GetSettings()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), settings1.CompanyName, settings2.CompanyName)
	assert.NotEqual(suite.T(), "Modified Company", settings2.CompanyName)

	// Invalidate cache and get again
	suite.repo.InvalidateCache()
	settings3, err := suite.repo.GetSettings()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Modified Company", settings3.CompanyName)
}

// Test ValidateSystemSettings
func (suite *SystemSettingsRepositoryTestSuite) TestValidateSystemSettings() {
	testCases := []struct {
		name        string
		settings    *model.SystemSettings
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid settings",
			settings: &model.SystemSettings{
				State:               string(model.StateBayern),
				DefaultWorkingHours: 40,
				DefaultVacationDays: 25,
			},
			shouldError: false,
		},
		{
			name: "Invalid state",
			settings: &model.SystemSettings{
				State: "InvalidState",
			},
			shouldError: true,
			errorMsg:    "invalid German state",
		},
		{
			name: "Negative working hours",
			settings: &model.SystemSettings{
				DefaultWorkingHours: -5,
			},
			shouldError: true,
			errorMsg:    "between 0 and 60",
		},
		{
			name: "Too many working hours",
			settings: &model.SystemSettings{
				DefaultWorkingHours: 80,
			},
			shouldError: true,
			errorMsg:    "between 0 and 60",
		},
		{
			name: "Negative vacation days",
			settings: &model.SystemSettings{
				DefaultVacationDays: -10,
			},
			shouldError: true,
			errorMsg:    "between 0 and 365",
		},
		{
			name: "Too many vacation days",
			settings: &model.SystemSettings{
				DefaultVacationDays: 400,
			},
			shouldError: true,
			errorMsg:    "between 0 and 365",
		},
		{
			name: "Invalid SMTP port",
			settings: &model.SystemSettings{
				EmailNotifications: &model.EmailNotificationSettings{
					SMTPHost: "smtp.example.com",
					SMTPPort: 0,
				},
			},
			shouldError: true,
			errorMsg:    "SMTP port must be positive",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.repo.ValidateSystemSettings(tc.settings)
			if tc.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Create
func (suite *SystemSettingsRepositoryTestSuite) TestCreate() {
	settings := &model.SystemSettings{
		CompanyName:         "Test Company",
		CompanyAddress:      "Test Address",
		State:               string(model.StateBerlin),
		DefaultWorkingHours: 38.5,
		DefaultVacationDays: 28,
	}

	err := suite.repo.Create(settings)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), settings.ID)
	assert.NotZero(suite.T(), settings.CreatedAt)
	assert.NotZero(suite.T(), settings.UpdatedAt)

	// Try to create another settings document (should fail)
	settings2 := &model.SystemSettings{
		CompanyName: "Another Company",
		State:       string(model.StateHamburg),
	}
	err = suite.repo.Create(settings2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "already exist")
}

// Test Update
func (suite *SystemSettingsRepositoryTestSuite) TestUpdate() {
	// Get default settings
	settings, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Update settings
	settings.CompanyName = "Updated Company"
	settings.CompanyAddress = "New Address"
	settings.State = string(model.StateBayern)
	settings.DefaultWorkingHours = 35
	settings.DefaultVacationDays = 25

	err = suite.repo.Update(settings)
	assert.NoError(suite.T(), err)

	// Verify update (bypass cache)
	suite.repo.InvalidateCache()
	updated, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Company", updated.CompanyName)
	assert.Equal(suite.T(), "New Address", updated.CompanyAddress)
	assert.Equal(suite.T(), string(model.StateBayern), updated.State)
	assert.Equal(suite.T(), float64(35), updated.DefaultWorkingHours)
	assert.Equal(suite.T(), 25, updated.DefaultVacationDays)
}

// Test Update - No ID
func (suite *SystemSettingsRepositoryTestSuite) TestUpdate_NoID() {
	// Create initial settings
	initial, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Update without ID
	settings := &model.SystemSettings{
		CompanyName:         "No ID Company",
		State:               string(model.StateHessen),
		DefaultWorkingHours: 40,
		DefaultVacationDays: 30,
	}

	err = suite.repo.Update(settings)
	assert.NoError(suite.T(), err)

	// Verify it updated the existing settings
	suite.repo.InvalidateCache()
	updated, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), initial.ID, updated.ID)
	assert.Equal(suite.T(), "No ID Company", updated.CompanyName)
}

// Test UpdateCompanyInfo
func (suite *SystemSettingsRepositoryTestSuite) TestUpdateCompanyInfo() {
	// Get default settings
	_, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Update company info
	err = suite.repo.UpdateCompanyInfo("New Company", "New Address", string(model.StateSachsen))
	assert.NoError(suite.T(), err)

	// Verify update
	settings, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Company", settings.CompanyName)
	assert.Equal(suite.T(), "New Address", settings.CompanyAddress)
	assert.Equal(suite.T(), string(model.StateSachsen), settings.State)

	// Test with invalid state
	err = suite.repo.UpdateCompanyInfo("", "", "InvalidState")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid German state")
}

// Test UpdateEmailNotifications
func (suite *SystemSettingsRepositoryTestSuite) TestUpdateEmailNotifications() {
	// Get default settings
	_, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Update email notifications
	notifications := &model.EmailNotificationSettings{
		Enabled:   true,
		SMTPHost:  "smtp.example.com",
		SMTPPort:  587,
		SMTPUser:  "user@example.com",
		SMTPPass:  "password",
		FromEmail: "noreply@example.com",
		FromName:  "PeopleFlow",
	}

	err = suite.repo.UpdateEmailNotifications(notifications)
	assert.NoError(suite.T(), err)

	// Verify update
	settings, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), settings.EmailNotifications)
	assert.True(suite.T(), settings.EmailNotifications.Enabled)
	assert.Equal(suite.T(), "smtp.example.com", settings.EmailNotifications.SMTPHost)
	assert.Equal(suite.T(), 587, settings.EmailNotifications.SMTPPort)

	// Test with nil notifications
	err = suite.repo.UpdateEmailNotifications(nil)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot be nil")

	// Test with invalid settings
	invalidNotifications := &model.EmailNotificationSettings{
		SMTPHost: "smtp.example.com",
		SMTPPort: -1,
	}
	err = suite.repo.UpdateEmailNotifications(invalidNotifications)
	assert.Error(suite.T(), err)
}

// Test UpdateWorkDefaults
func (suite *SystemSettingsRepositoryTestSuite) TestUpdateWorkDefaults() {
	// Get default settings
	_, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Update work defaults
	err = suite.repo.UpdateWorkDefaults(37.5, 24)
	assert.NoError(suite.T(), err)

	// Verify update
	settings, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 37.5, settings.DefaultWorkingHours)
	assert.Equal(suite.T(), 24, settings.DefaultVacationDays)

	// Test with invalid values
	err = suite.repo.UpdateWorkDefaults(-10, 20)
	assert.Error(suite.T(), err)

	err = suite.repo.UpdateWorkDefaults(40, 400)
	assert.Error(suite.T(), err)
}

// Test Email Notification Methods
func (suite *SystemSettingsRepositoryTestSuite) TestEmailNotificationMethods() {
	// Initially disabled
	enabled, err := suite.repo.IsEmailNotificationEnabled()
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), enabled)

	// Get SMTP config should fail
	_, err = suite.repo.GetSMTPConfig()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not enabled")

	// Enable notifications
	notifications := &model.EmailNotificationSettings{
		Enabled:  true,
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
	}
	err = suite.repo.UpdateEmailNotifications(notifications)
	require.NoError(suite.T(), err)

	// Check again
	enabled, err = suite.repo.IsEmailNotificationEnabled()
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), enabled)

	// Get SMTP config should succeed
	config, err := suite.repo.GetSMTPConfig()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "smtp.example.com", config.SMTPHost)
	assert.Equal(suite.T(), 587, config.SMTPPort)
}

// Test ResetToDefaults
func (suite *SystemSettingsRepositoryTestSuite) TestResetToDefaults() {
	// Create and modify settings
	settings, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	settings.CompanyName = "Test Company"
	settings.State = string(model.StateBerlin)
	settings.DefaultWorkingHours = 35
	err = suite.repo.Update(settings)
	require.NoError(suite.T(), err)

	originalID := settings.ID
	originalCreatedAt := settings.CreatedAt

	// Reset to defaults
	err = suite.repo.ResetToDefaults()
	assert.NoError(suite.T(), err)

	// Verify reset
	suite.repo.InvalidateCache()
	reset, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), originalID, reset.ID)               // ID should be preserved
	assert.Equal(suite.T(), originalCreatedAt, reset.CreatedAt) // CreatedAt should be preserved
	assert.Empty(suite.T(), reset.CompanyName)
	assert.Equal(suite.T(), string(model.StateNordrheinWestfalen), reset.State)
	assert.Equal(suite.T(), float64(40), reset.DefaultWorkingHours)
	assert.Equal(suite.T(), 30, reset.DefaultVacationDays)
}

// Test GetCompanyState
func (suite *SystemSettingsRepositoryTestSuite) TestGetCompanyState() {
	// Default state
	state, err := suite.repo.GetCompanyState()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.StateNordrheinWestfalen, state)

	// Update state
	err = suite.repo.UpdateCompanyInfo("", "", string(model.StateBayern))
	require.NoError(suite.T(), err)

	state, err = suite.repo.GetCompanyState()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.StateBayern, state)
}

// Test EnsureSingleDocument
func (suite *SystemSettingsRepositoryTestSuite) TestEnsureSingleDocument() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create multiple settings documents directly
	for i := 0; i < 3; i++ {
		settings := model.SystemSettings{
			CompanyName:         fmt.Sprintf("Company %d", i),
			State:               string(model.StateNordrheinWestfalen),
			DefaultWorkingHours: 40,
			DefaultVacationDays: 30,
			CreatedAt:           time.Now().Add(time.Duration(i) * time.Hour),
			UpdatedAt:           time.Now().Add(time.Duration(i) * time.Hour),
		}
		_, err := suite.collection.InsertOne(ctx, settings)
		require.NoError(suite.T(), err)
	}

	// Verify multiple documents exist
	count, err := suite.collection.CountDocuments(ctx, bson.M{})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), count)

	// Ensure single document
	err = suite.repo.EnsureSingleDocument()
	assert.NoError(suite.T(), err)

	// Verify only one document remains
	count, err = suite.collection.CountDocuments(ctx, bson.M{})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)

	// Verify it kept the most recent
	var remaining model.SystemSettings
	err = suite.collection.FindOne(ctx, bson.M{}).Decode(&remaining)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Company 2", remaining.CompanyName)
}

// Test Concurrent Access
func (suite *SystemSettingsRepositoryTestSuite) TestConcurrentAccess() {
	// Create initial settings
	_, err := suite.repo.GetSettings()
	require.NoError(suite.T(), err)

	// Concurrent reads and writes
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Multiple readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := suite.repo.GetSettings()
			if err != nil {
				errors <- err
			}
		}()
	}

	// Multiple writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			err := suite.repo.UpdateCompanyInfo(
				fmt.Sprintf("Company %d", index),
				fmt.Sprintf("Address %d", index),
				"",
			)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(suite.T(), err)
	}
}

// Test All German States
func (suite *SystemSettingsRepositoryTestSuite) TestAllGermanStates() {
	states := []model.GermanState{
		model.StateBadenWuerttemberg,
		model.StateBayern,
		model.StateBerlin,
		model.StateBrandenburg,
		model.StateBremen,
		model.StateHamburg,
		model.StateHessen,
		model.StateMecklenburgVorpommern,
		model.StateNiedersachsen,
		model.StateNordrheinWestfalen,
		model.StateRheinlandPfalz,
		model.StateSaarland,
		model.StateSachsen,
		model.StateSachsenAnhalt,
		model.StateSchleswigHolstein,
		model.StateThueringen,
	}

	for _, state := range states {
		err := suite.repo.UpdateCompanyInfo("", "", string(state))
		assert.NoError(suite.T(), err, "State %s should be valid", state)

		retrievedState, err := suite.repo.GetCompanyState()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), state, retrievedState)
	}
}

// Run the test suite
func TestSystemSettingsRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SystemSettingsRepositoryTestSuite))
}
