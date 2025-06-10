package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ActivityRepositoryTestSuite defines the test suite for ActivityRepository
type ActivityRepositoryTestSuite struct {
	suite.Suite
	repo       *ActivityRepository
	collection *mongo.Collection
	client     *mongo.Client
}

// SetupSuite runs once before all tests
func (suite *ActivityRepositoryTestSuite) SetupSuite() {
	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(suite.T(), err)

	suite.client = client
	suite.collection = client.Database("peopleflow_test").Collection("activities")

	// Initialize test repository
	db.SetTestCollection("activities", suite.collection)
	suite.repo = NewActivityRepository()
}

// SetupTest runs before each test
func (suite *ActivityRepositoryTestSuite) SetupTest() {
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
func (suite *ActivityRepositoryTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if suite.client != nil {
		suite.client.Disconnect(ctx)
	}
}

// Helper function to create a valid activity
func (suite *ActivityRepositoryTestSuite) createValidActivity() *model.Activity {
	return &model.Activity{
		Type:        model.ActivityTypeEmployeeAdded,
		UserID:      primitive.NewObjectID(),
		UserName:    "Test Admin",
		TargetID:    primitive.NewObjectID(),
		TargetType:  "employee",
		TargetName:  "John Doe",
		Description: "Added new employee John Doe",
		Timestamp:   time.Now(),
	}
}

// Test ValidateActivity
func (suite *ActivityRepositoryTestSuite) TestValidateActivity() {
	testCases := []struct {
		name        string
		activity    *model.Activity
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid activity",
			activity:    suite.createValidActivity(),
			shouldError: false,
		},
		{
			name: "Invalid activity type",
			activity: &model.Activity{
				Type:     "invalid_type",
				UserID:   primitive.NewObjectID(),
				UserName: "Test User",
			},
			shouldError: true,
			errorMsg:    "invalid activity type",
		},
		{
			name: "Missing user ID",
			activity: &model.Activity{
				Type:     model.ActivityTypeEmployeeAdded,
				UserName: "Test User",
			},
			shouldError: true,
			errorMsg:    "user ID is required",
		},
		{
			name: "Missing user name",
			activity: &model.Activity{
				Type:   model.ActivityTypeEmployeeAdded,
				UserID: primitive.NewObjectID(),
			},
			shouldError: true,
			errorMsg:    "user name is required",
		},
		{
			name: "Missing target ID for employee activity",
			activity: &model.Activity{
				Type:       model.ActivityTypeEmployeeAdded,
				UserID:     primitive.NewObjectID(),
				UserName:   "Test User",
				TargetType: "employee",
				TargetName: "John Doe",
			},
			shouldError: true,
			errorMsg:    "target ID is required",
		},
		{
			name: "System setting change without target",
			activity: &model.Activity{
				Type:        model.ActivityTypeSystemSettingChanged,
				UserID:      primitive.NewObjectID(),
				UserName:    "Admin",
				Description: "Changed company name",
			},
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.repo.ValidateActivity(tc.activity)
			if tc.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test Create Activity
func (suite *ActivityRepositoryTestSuite) TestCreateActivity() {
	activity := suite.createValidActivity()

	err := suite.repo.Create(activity)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), activity.ID)
	assert.NotZero(suite.T(), activity.Timestamp)

	// Verify activity was created
	found, err := suite.repo.FindByID(activity.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), activity.Type, found.Type)
	assert.Equal(suite.T(), activity.UserName, found.UserName)
}

// Test FindByID
func (suite *ActivityRepositoryTestSuite) TestFindByID() {
	// Create activity
	activity := suite.createValidActivity()
	err := suite.repo.Create(activity)
	require.NoError(suite.T(), err)

	// Find by ID
	found, err := suite.repo.FindByID(activity.ID.Hex())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), activity.Type, found.Type)

	// Test not found
	_, err = suite.repo.FindByID("507f1f77bcf86cd799439011")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrActivityNotFound, err)

	// Test invalid ID
	_, err = suite.repo.FindByID("invalid-id")
	assert.Error(suite.T(), err)
}

// Test FindByUser
func (suite *ActivityRepositoryTestSuite) TestFindByUser() {
	userID := primitive.NewObjectID()

	// Create multiple activities for the user
	for i := 0; i < 5; i++ {
		activity := suite.createValidActivity()
		activity.UserID = userID
		activity.Description = fmt.Sprintf("Action %d", i)
		err := suite.repo.Create(activity)
		require.NoError(suite.T(), err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Create activities for different user
	for i := 0; i < 3; i++ {
		activity := suite.createValidActivity()
		err := suite.repo.Create(activity)
		require.NoError(suite.T(), err)
	}

	// Find by user
	activities, err := suite.repo.FindByUser(userID.Hex(), 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 5)

	// Verify sorting (newest first)
	assert.Contains(suite.T(), activities[0].Description, "Action 4")
}

// Test FindByTarget
func (suite *ActivityRepositoryTestSuite) TestFindByTarget() {
	targetID := primitive.NewObjectID()

	// Create activities for the target
	for i := 0; i < 4; i++ {
		activity := suite.createValidActivity()
		activity.TargetID = targetID
		err := suite.repo.Create(activity)
		require.NoError(suite.T(), err)
	}

	// Find by target
	activities, err := suite.repo.FindByTarget(targetID.Hex(), 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 4)
}

// Test FindByType
func (suite *ActivityRepositoryTestSuite) TestFindByType() {
	// Create activities of different types
	types := []model.ActivityType{
		model.ActivityTypeEmployeeAdded,
		model.ActivityTypeEmployeeAdded,
		model.ActivityTypeVacationRequested,
		model.ActivityTypeVacationApproved,
		model.ActivityTypeEmployeeAdded,
	}

	for _, actType := range types {
		activity := suite.createValidActivity()
		activity.Type = actType
		err := suite.repo.Create(activity)
		require.NoError(suite.T(), err)
	}

	// Find employee added activities
	activities, err := suite.repo.FindByType(model.ActivityTypeEmployeeAdded, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 3)
}

// Test FindByDateRange
func (suite *ActivityRepositoryTestSuite) TestFindByDateRange() {
	now := time.Now()

	// Create activities at different times
	times := []time.Time{
		now.Add(-72 * time.Hour), // 3 days ago
		now.Add(-48 * time.Hour), // 2 days ago
		now.Add(-24 * time.Hour), // 1 day ago
		now.Add(-12 * time.Hour), // 12 hours ago
		now.Add(-1 * time.Hour),  // 1 hour ago
	}

	for _, timestamp := range times {
		activity := suite.createValidActivity()
		activity.Timestamp = timestamp

		// Insert directly to set specific timestamp
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := suite.collection.InsertOne(ctx, activity)
		require.NoError(suite.T(), err)
	}

	// Find activities from last 2 days
	start := now.Add(-50 * time.Hour)
	end := now
	activities, total, err := suite.repo.FindByDateRange(start, end, 0, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 4)
	assert.Equal(suite.T(), int64(4), total)
}

// Test LogActivity
func (suite *ActivityRepositoryTestSuite) TestLogActivity() {
	userID := primitive.NewObjectID()
	targetID := primitive.NewObjectID()

	activity, err := suite.repo.LogActivity(
		model.ActivityTypeEmployeeUpdated,
		userID,
		"Admin User",
		targetID,
		"employee",
		"Jane Doe",
		"Updated employee information",
	)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), activity)
	assert.NotEmpty(suite.T(), activity.ID)
	assert.Equal(suite.T(), model.ActivityTypeEmployeeUpdated, activity.Type)
}

// Test DeleteOldActivities
func (suite *ActivityRepositoryTestSuite) TestDeleteOldActivities() {
	now := time.Now()

	// Create old and new activities
	oldActivity := suite.createValidActivity()
	oldActivity.Timestamp = now.Add(-100 * 24 * time.Hour) // 100 days ago

	newActivity := suite.createValidActivity()
	newActivity.Timestamp = now.Add(-10 * 24 * time.Hour) // 10 days ago

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := suite.collection.InsertOne(ctx, oldActivity)
	require.NoError(suite.T(), err)

	_, err = suite.collection.InsertOne(ctx, newActivity)
	require.NoError(suite.T(), err)

	// Delete activities older than 90 days
	deleted, err := suite.repo.DeleteOldActivities(90 * 24 * time.Hour)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), deleted)

	// Verify only new activity remains
	count, err := suite.repo.Count(bson.M{})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

// Test GetActivityStats
func (suite *ActivityRepositoryTestSuite) TestGetActivityStats() {
	// Create activities of different types
	activities := []struct {
		actType model.ActivityType
		count   int
	}{
		{model.ActivityTypeEmployeeAdded, 5},
		{model.ActivityTypeVacationRequested, 3},
		{model.ActivityTypeOvertimeAdjusted, 2},
	}

	for _, a := range activities {
		for i := 0; i < a.count; i++ {
			activity := suite.createValidActivity()
			activity.Type = a.actType
			err := suite.repo.Create(activity)
			require.NoError(suite.T(), err)
		}
	}

	// Get stats for last 7 days
	stats, err := suite.repo.GetActivityStats(7)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)

	assert.Equal(suite.T(), 7, stats["period_days"])
	assert.Equal(suite.T(), 10, stats["total_activities"])

	typeStats := stats["by_type"].(map[string]int)
	assert.Equal(suite.T(), 5, typeStats[string(model.ActivityTypeEmployeeAdded)])
	assert.Equal(suite.T(), 3, typeStats[string(model.ActivityTypeVacationRequested)])
	assert.Equal(suite.T(), 2, typeStats[string(model.ActivityTypeOvertimeAdjusted)])
}

// Test FindRecent
func (suite *ActivityRepositoryTestSuite) TestFindRecent() {
	// Create 10 activities
	for i := 0; i < 10; i++ {
		activity := suite.createValidActivity()
		activity.Description = fmt.Sprintf("Activity %d", i)
		err := suite.repo.Create(activity)
		require.NoError(suite.T(), err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Find recent 5
	activities, err := suite.repo.FindRecent(5)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), activities, 5)

	// Verify newest first
	assert.Contains(suite.T(), activities[0].Description, "Activity 9")
}

// Test Indexes
func (suite *ActivityRepositoryTestSuite) TestIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get index info
	cursor, err := suite.collection.Indexes().List(ctx)
	require.NoError(suite.T(), err)

	var indexes []bson.M
	err = cursor.All(ctx, &indexes)
	require.NoError(suite.T(), err)

	// Verify indexes exist (including default _id index)
	assert.GreaterOrEqual(suite.T(), len(indexes), 5)

	indexNames := make(map[string]bool)
	for _, idx := range indexes {
		if name, ok := idx["name"].(string); ok {
			indexNames[name] = true
		}
	}

	// Check for expected indexes
	assert.True(suite.T(), indexNames["timestamp_-1"])
	assert.True(suite.T(), indexNames["userId_1"])
	assert.True(suite.T(), indexNames["targetId_1"])
}

// Run the test suite
func TestActivityRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityRepositoryTestSuite))
}
