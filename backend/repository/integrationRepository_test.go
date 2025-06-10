// backend/repository/integrationRepository_test.go
package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IntegrationRepositoryTestSuite defines the test suite for IntegrationRepository
type IntegrationRepositoryTestSuite struct {
	suite.Suite
	repo       *IntegrationRepository
	collection *mongo.Collection
	client     *mongo.Client
}

// SetupSuite runs once before all tests
func (suite *IntegrationRepositoryTestSuite) SetupSuite() {
	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(suite.T(), err)

	suite.client = client
	suite.collection = client.Database("peopleflow_test").Collection("integrations")

	// Initialize test repository
	db.SetTestCollection("integrations", suite.collection)
	suite.repo = NewIntegrationRepository()

	// Initialize encryption key for tests
	utils.InitializeEncryptionKey("test-encryption-key-32-characters")
}

// SetupTest runs before each test
func (suite *IntegrationRepositoryTestSuite) SetupTest() {
	// Clear collection before each test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := suite.collection.Drop(ctx)
	if err != nil && err != mongo.ErrNilDocument {
		suite.T().Fatal(err)
	}

	// Create indexes
	err = suite.repo.CreateIndexes()
	require.NoError(suite.T(), err)
}

// TearDownSuite runs once after all tests
func (suite *IntegrationRepositoryTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if suite.client != nil {
		suite.client.Disconnect(ctx)
	}
}

// Test ValidateIntegrationType
func (suite *IntegrationRepositoryTestSuite) TestValidateIntegrationType() {
	testCases := []struct {
		name            string
		integrationType string
		shouldError     bool
	}{
		{"Valid timebutler", "timebutler", false},
		{"Valid 123erfasst", "123erfasst", false},
		{"Valid awork", "awork", false},
		{"Valid with spaces", "  timebutler  ", false},
		{"Valid uppercase", "TIMEBUTLER", false},
		{"Invalid type", "invalid", true},
		{"Empty type", "", true},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.repo.ValidateIntegrationType(tc.integrationType)
			if tc.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid integration type")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test SaveApiKey - Success
func (suite *IntegrationRepositoryTestSuite) TestSaveApiKey_Success() {
	apiKey := "test-api-key-12345"

	err := suite.repo.SaveApiKey("timebutler", apiKey)
	assert.NoError(suite.T(), err)

	// Verify integration was created
	integration, err := suite.repo.GetIntegration("timebutler")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "timebutler", integration.Type)
	assert.Equal(suite.T(), "Timebutler", integration.Name)
	assert.True(suite.T(), integration.Active)
	assert.NotEmpty(suite.T(), integration.ApiKey)
	assert.NotEqual(suite.T(), apiKey, integration.ApiKey) // Should be encrypted
}

// Test SaveApiKey - Update Existing
func (suite *IntegrationRepositoryTestSuite) TestSaveApiKey_UpdateExisting() {
	// Save first API key
	err := suite.repo.SaveApiKey("timebutler", "first-key")
	require.NoError(suite.T(), err)

	// Get creation time
	firstIntegration, err := suite.repo.GetIntegration("timebutler")
	require.NoError(suite.T(), err)
	createdAt := firstIntegration.CreatedAt

	// Wait a bit to ensure UpdatedAt differs
	time.Sleep(10 * time.Millisecond)

	// Update with new API key
	err = suite.repo.SaveApiKey("timebutler", "second-key")
	assert.NoError(suite.T(), err)

	// Verify update
	updatedIntegration, err := suite.repo.GetIntegration("timebutler")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdAt, updatedIntegration.CreatedAt)
	assert.True(suite.T(), updatedIntegration.UpdatedAt.After(createdAt))
}

// Test SaveApiKey - Invalid Input
func (suite *IntegrationRepositoryTestSuite) TestSaveApiKey_InvalidInput() {
	// Invalid integration type
	err := suite.repo.SaveApiKey("invalid", "api-key")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidIntegrationType, errors.Unwrap(err))

	// Empty API key
	err = suite.repo.SaveApiKey("timebutler", "")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidApiKey, err)

	// Whitespace API key
	err = suite.repo.SaveApiKey("timebutler", "   ")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidApiKey, err)
}

// Test GetApiKey - Success
func (suite *IntegrationRepositoryTestSuite) TestGetApiKey_Success() {
	originalKey := "test-api-key-12345"

	// Save API key
	err := suite.repo.SaveApiKey("timebutler", originalKey)
	require.NoError(suite.T(), err)

	// Get and decrypt API key
	retrievedKey, err := suite.repo.GetApiKey("timebutler")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), originalKey, retrievedKey)
}

// Test GetApiKey - Not Found
func (suite *IntegrationRepositoryTestSuite) TestGetApiKey_NotFound() {
	_, err := suite.repo.GetApiKey("timebutler")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrIntegrationNotFound, err)
}

// Test GetApiKey - Inactive Integration
func (suite *IntegrationRepositoryTestSuite) TestGetApiKey_InactiveIntegration() {
	// Save API key
	err := suite.repo.SaveApiKey("timebutler", "test-key")
	require.NoError(suite.T(), err)

	// Deactivate integration
	err = suite.repo.SetIntegrationStatus("timebutler", false)
	require.NoError(suite.T(), err)

	// Try to get API key
	_, err = suite.repo.GetApiKey("timebutler")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not active")
}

// Test Integration Status
func (suite *IntegrationRepositoryTestSuite) TestIntegrationStatus() {
	// Check status for non-existent integration
	active, err := suite.repo.GetIntegrationStatus("timebutler")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), active)

	// Save API key (sets active to true)
	err = suite.repo.SaveApiKey("timebutler", "test-key")
	require.NoError(suite.T(), err)

	// Check status
	active, err = suite.repo.GetIntegrationStatus("timebutler")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), active)

	// Deactivate
	err = suite.repo.SetIntegrationStatus("timebutler", false)
	assert.NoError(suite.T(), err)

	// Check status again
	active, err = suite.repo.GetIntegrationStatus("timebutler")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), active)

	// Try to set status for non-existent integration
	err = suite.repo.SetIntegrationStatus("awork", true)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrIntegrationNotFound, err)
}

// Test Metadata Operations
func (suite *IntegrationRepositoryTestSuite) TestMetadataOperations() {
	// Set metadata for new integration
	err := suite.repo.SetMetadata("timebutler", "sync_interval", "30m")
	assert.NoError(suite.T(), err)

	// Verify integration was created
	integration, err := suite.repo.GetIntegration("timebutler")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), integration.Active) // Should be inactive since no API key

	// Get metadata
	value, err := suite.repo.GetMetadata("timebutler", "sync_interval")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "30m", value)

	// Add more metadata
	err = suite.repo.SetMetadata("timebutler", "last_error", "connection timeout")
	assert.NoError(suite.T(), err)

	// Get all metadata
	metadata, err := suite.repo.GetAllMetadata("timebutler")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), metadata, 2)
	assert.Equal(suite.T(), "30m", metadata["sync_interval"])
	assert.Equal(suite.T(), "connection timeout", metadata["last_error"])

	// Update existing metadata
	err = suite.repo.SetMetadata("timebutler", "sync_interval", "1h")
	assert.NoError(suite.T(), err)

	value, err = suite.repo.GetMetadata("timebutler", "sync_interval")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "1h", value)

	// Delete metadata
	err = suite.repo.DeleteMetadata("timebutler", "last_error")
	assert.NoError(suite.T(), err)

	// Verify deletion
	value, err = suite.repo.GetMetadata("timebutler", "last_error")
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), value)

	// Get all metadata after deletion
	metadata, err = suite.repo.GetAllMetadata("timebutler")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), metadata, 1)
}

// Test Metadata Invalid Input
func (suite *IntegrationRepositoryTestSuite) TestMetadata_InvalidInput() {
	// Empty key
	err := suite.repo.SetMetadata("timebutler", "", "value")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "key cannot be empty")

	// Get with empty key
	_, err = suite.repo.GetMetadata("timebutler", "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "key cannot be empty")

	// Delete with empty key
	err = suite.repo.DeleteMetadata("timebutler", "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "key cannot be empty")

	// Invalid integration type
	err = suite.repo.SetMetadata("invalid", "key", "value")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid integration type")
}

// Test Last Sync
func (suite *IntegrationRepositoryTestSuite) TestLastSync() {
	// Create integration
	err := suite.repo.SaveApiKey("timebutler", "test-key")
	require.NoError(suite.T(), err)

	// Set last sync
	syncTime := time.Now().Add(-1 * time.Hour)
	err = suite.repo.SetLastSync("timebutler", syncTime)
	assert.NoError(suite.T(), err)

	// Get last sync
	retrievedTime, err := suite.repo.GetLastSync("timebutler")
	assert.NoError(suite.T(), err)
	assert.WithinDuration(suite.T(), syncTime, retrievedTime, time.Second)

	// Update last sync
	newSyncTime := time.Now()
	err = suite.repo.SetLastSync("timebutler", newSyncTime)
	assert.NoError(suite.T(), err)

	retrievedTime, err = suite.repo.GetLastSync("timebutler")
	assert.NoError(suite.T(), err)
	assert.WithinDuration(suite.T(), newSyncTime, retrievedTime, time.Second)

	// Try to set last sync for non-existent integration
	err = suite.repo.SetLastSync("awork", time.Now())
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrIntegrationNotFound, err)
}

// Test GetAllIntegrations
func (suite *IntegrationRepositoryTestSuite) TestGetAllIntegrations() {
	// Initially empty
	integrations, err := suite.repo.GetAllIntegrations()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), integrations, 0)

	// Add integrations
	err = suite.repo.SaveApiKey("timebutler", "key1")
	require.NoError(suite.T(), err)

	err = suite.repo.SaveApiKey("123erfasst", "key2")
	require.NoError(suite.T(), err)

	err = suite.repo.SaveApiKey("awork", "key3")
	require.NoError(suite.T(), err)

	// Get all integrations
	integrations, err = suite.repo.GetAllIntegrations()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), integrations, 3)

	// Verify sorting by name
	assert.Equal(suite.T(), "123erfasst", integrations[0].Name)
	assert.Equal(suite.T(), "AWork", integrations[1].Name)
	assert.Equal(suite.T(), "Timebutler", integrations[2].Name)
}

// Test GetActiveIntegrations
func (suite *IntegrationRepositoryTestSuite) TestGetActiveIntegrations() {
	// Add integrations
	err := suite.repo.SaveApiKey("timebutler", "key1")
	require.NoError(suite.T(), err)

	err = suite.repo.SaveApiKey("123erfasst", "key2")
	require.NoError(suite.T(), err)

	err = suite.repo.SaveApiKey("awork", "key3")
	require.NoError(suite.T(), err)

	// Deactivate one
	err = suite.repo.SetIntegrationStatus("123erfasst", false)
	require.NoError(suite.T(), err)

	// Get active integrations
	integrations, err := suite.repo.GetActiveIntegrations()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), integrations, 2)

	// Verify only active integrations returned
	for _, integration := range integrations {
		assert.True(suite.T(), integration.Active)
		assert.NotEqual(suite.T(), "123erfasst", integration.Type)
	}
}

// Test DeleteIntegration
func (suite *IntegrationRepositoryTestSuite) TestDeleteIntegration() {
	// Create integration
	err := suite.repo.SaveApiKey("timebutler", "test-key")
	require.NoError(suite.T(), err)

	// Add metadata
	err = suite.repo.SetMetadata("timebutler", "config", "value")
	require.NoError(suite.T(), err)

	// Delete integration
	err = suite.repo.DeleteIntegration("timebutler")
	assert.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repo.GetIntegration("timebutler")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrIntegrationNotFound, err)

	// Try to delete non-existent integration
	err = suite.repo.DeleteIntegration("timebutler")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrIntegrationNotFound, err)
}

// Test Transaction in Metadata
func (suite *IntegrationRepositoryTestSuite) TestTransactionInMetadata() {
	// This tests the transaction behavior in SetMetadata
	// when creating a new integration

	// Set metadata which should create a new integration
	err := suite.repo.SetMetadata("timebutler", "key1", "value1")
	assert.NoError(suite.T(), err)

	// Verify integration was created with correct defaults
	integration, err := suite.repo.GetIntegration("timebutler")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "timebutler", integration.Type)
	assert.Equal(suite.T(), "Timebutler", integration.Name)
	assert.False(suite.T(), integration.Active)
	assert.NotNil(suite.T(), integration.Metadata)
	assert.Equal(suite.T(), "value1", integration.Metadata["key1"])
}

// Test Indexes
func (suite *IntegrationRepositoryTestSuite) TestIndexes() {
	// Test unique type index by trying to insert duplicate directly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert first document
	_, err := suite.collection.InsertOne(ctx, bson.M{
		"type":   "timebutler",
		"name":   "Timebutler",
		"active": true,
	})
	require.NoError(suite.T(), err)

	// Try to insert duplicate type
	_, err = suite.collection.InsertOne(ctx, bson.M{
		"type":   "timebutler",
		"name":   "Timebutler Duplicate",
		"active": false,
	})
	assert.Error(suite.T(), err)
	assert.True(suite.T(), mongo.IsDuplicateKeyError(err))
}

// Test Case Insensitive Integration Type
func (suite *IntegrationRepositoryTestSuite) TestCaseInsensitiveIntegrationType() {
	// Save with uppercase
	err := suite.repo.SaveApiKey("TIMEBUTLER", "test-key")
	assert.NoError(suite.T(), err)

	// Retrieve with lowercase
	key, err := suite.repo.GetApiKey("timebutler")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-key", key)

	// Retrieve with mixed case
	key, err = suite.repo.GetApiKey("TimeBUTLER")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-key", key)
}

// Run the test suite
func TestIntegrationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationRepositoryTestSuite))
}
