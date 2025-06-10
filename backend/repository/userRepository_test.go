// backend/repository/userRepository_test.go
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepositoryTestSuite defines the test suite for UserRepository
type UserRepositoryTestSuite struct {
	suite.Suite
	repo       *UserRepository
	collection *mongo.Collection
	client     *mongo.Client
}

// SetupSuite runs once before all tests
func (suite *UserRepositoryTestSuite) SetupSuite() {
	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(suite.T(), err)

	suite.client = client
	suite.collection = client.Database("peopleflow_test").Collection("users")

	// Initialize test repository
	db.SetTestCollection("users", suite.collection)
	suite.repo = NewUserRepository()
}

// SetupTest runs before each test
func (suite *UserRepositoryTestSuite) SetupTest() {
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
func (suite *UserRepositoryTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if suite.client != nil {
		suite.client.Disconnect(ctx)
	}
}

// Test Create User - Success
func (suite *UserRepositoryTestSuite) TestCreateUser_Success() {
	user := &model.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      model.RoleEmployee,
		Status:    model.StatusActive,
	}

	err := suite.repo.Create(user)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), user.ID)
	assert.NotEmpty(suite.T(), user.PasswordHash)
	assert.NotEqual(suite.T(), user.Password, user.PasswordHash)
	assert.Equal(suite.T(), "test@example.com", user.Email) // Should be normalized
}

// Test Create User - Invalid Email
func (suite *UserRepositoryTestSuite) TestCreateUser_InvalidEmail() {
	testCases := []struct {
		name  string
		email string
	}{
		{"Empty email", ""},
		{"Invalid format", "notanemail"},
		{"Missing domain", "test@"},
		{"Missing local part", "@example.com"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user := &model.User{
				Email:    tc.email,
				Password: "password123",
				Role:     model.RoleEmployee,
				Status:   model.StatusActive,
			}

			err := suite.repo.Create(user)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "email")
		})
	}
}

// Test Create User - Short Password
func (suite *UserRepositoryTestSuite) TestCreateUser_ShortPassword() {
	user := &model.User{
		Email:    "test@example.com",
		Password: "short",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrInvalidPassword, err)
}

// Test Create User - Duplicate Email
func (suite *UserRepositoryTestSuite) TestCreateUser_DuplicateEmail() {
	// Create first user
	user1 := &model.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user1)
	require.NoError(suite.T(), err)

	// Try to create second user with same email
	user2 := &model.User{
		Email:    "TEST@EXAMPLE.COM", // Test case insensitive
		Password: "password456",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err = suite.repo.Create(user2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmailTaken, err)
}

// Test FindByID - Success
func (suite *UserRepositoryTestSuite) TestFindByID_Success() {
	// Create user
	user := &model.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Find by ID
	found, err := suite.repo.FindByID(user.ID.Hex())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Email, found.Email)
	assert.Equal(suite.T(), user.Role, found.Role)
}

// Test FindByID - Not Found
func (suite *UserRepositoryTestSuite) TestFindByID_NotFound() {
	_, err := suite.repo.FindByID("507f1f77bcf86cd799439011")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrUserNotFound, err)
}

// Test FindByID - Invalid ID
func (suite *UserRepositoryTestSuite) TestFindByID_InvalidID() {
	_, err := suite.repo.FindByID("invalid-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// Test FindByEmail - Success
func (suite *UserRepositoryTestSuite) TestFindByEmail_Success() {
	// Create user
	user := &model.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Find by email (case insensitive)
	found, err := suite.repo.FindByEmail("TEST@EXAMPLE.COM")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, found.ID)
}

// Test Update User - Success
func (suite *UserRepositoryTestSuite) TestUpdateUser_Success() {
	// Create user
	user := &model.User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Old",
		LastName:  "Name",
		Role:      model.RoleEmployee,
		Status:    model.StatusActive,
	}

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Update user
	user.FirstName = "New"
	user.LastName = "Name"
	user.Role = model.RoleAdmin

	err = suite.repo.Update(user)
	assert.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.repo.FindByID(user.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New", updated.FirstName)
	assert.Equal(suite.T(), model.RoleAdmin, updated.Role)
}

// Test Update User - Email Taken
func (suite *UserRepositoryTestSuite) TestUpdateUser_EmailTaken() {
	// Create two users
	user1 := &model.User{
		Email:    "user1@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	user2 := &model.User{
		Email:    "user2@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(user2)
	require.NoError(suite.T(), err)

	// Try to update user2 with user1's email
	user2.Email = user1.Email
	err = suite.repo.Update(user2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmailTaken, err)
}

// Test UpdatePassword
func (suite *UserRepositoryTestSuite) TestUpdatePassword() {
	// Create user
	user := &model.User{
		Email:    "test@example.com",
		Password: "oldpassword",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	oldHash := user.PasswordHash

	// Update password
	err = suite.repo.UpdatePassword(user.ID.Hex(), "newpassword123")
	assert.NoError(suite.T(), err)

	// Verify password was updated
	updated, err := suite.repo.FindByID(user.ID.Hex())
	require.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), oldHash, updated.PasswordHash)
	assert.True(suite.T(), updated.CheckPassword("newpassword123"))
}

// Test Delete User (Soft Delete)
func (suite *UserRepositoryTestSuite) TestDeleteUser() {
	// Create user
	user := &model.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	err := suite.repo.Create(user)
	require.NoError(suite.T(), err)

	// Delete user
	err = suite.repo.Delete(user.ID.Hex())
	assert.NoError(suite.T(), err)

	// Verify user is soft deleted
	deleted, err := suite.repo.FindByID(user.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.StatusInactive, deleted.Status)
	assert.NotNil(suite.T(), deleted.DeletedAt)
}

// Test FindAll with Pagination
func (suite *UserRepositoryTestSuite) TestFindAll_Pagination() {
	// Create multiple users
	for i := 0; i < 15; i++ {
		user := &model.User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password123",
			Role:     model.RoleEmployee,
			Status:   model.StatusActive,
		}
		err := suite.repo.Create(user)
		require.NoError(suite.T(), err)
	}

	// Test first page
	users, total, err := suite.repo.FindAll(0, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 10)
	assert.Equal(suite.T(), int64(15), total)

	// Test second page
	users, total, err = suite.repo.FindAll(10, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), users, 5)
	assert.Equal(suite.T(), int64(15), total)
}

// Test FindByRole
func (suite *UserRepositoryTestSuite) TestFindByRole() {
	// Create users with different roles
	admin := &model.User{
		Email:    "admin@example.com",
		Password: "password123",
		Role:     model.RoleAdmin,
		Status:   model.StatusActive,
	}

	employee1 := &model.User{
		Email:    "employee1@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	employee2 := &model.User{
		Email:    "employee2@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusActive,
	}

	inactiveEmployee := &model.User{
		Email:    "inactive@example.com",
		Password: "password123",
		Role:     model.RoleEmployee,
		Status:   model.StatusInactive,
	}

	err := suite.repo.Create(admin)
	require.NoError(suite.T(), err)
	err = suite.repo.Create(employee1)
	require.NoError(suite.T(), err)
	err = suite.repo.Create(employee2)
	require.NoError(suite.T(), err)
	err = suite.repo.Create(inactiveEmployee)
	require.NoError(suite.T(), err)

	// Find employees
	employees, err := suite.repo.FindByRole(model.RoleEmployee)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 2) // Only active employees

	// Find admins
	admins, err := suite.repo.FindByRole(model.RoleAdmin)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), admins, 1)
}

// Test BulkUpdateStatus with Transaction
func (suite *UserRepositoryTestSuite) TestBulkUpdateStatus() {
	// Create multiple users
	var userIDs []string
	for i := 0; i < 5; i++ {
		user := &model.User{
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password123",
			Role:     model.RoleEmployee,
			Status:   model.StatusActive,
		}
		err := suite.repo.Create(user)
		require.NoError(suite.T(), err)
		userIDs = append(userIDs, user.ID.Hex())
	}

	// Bulk update status
	err := suite.repo.BulkUpdateStatus(userIDs, model.StatusInactive)
	assert.NoError(suite.T(), err)

	// Verify all users are inactive
	for _, id := range userIDs {
		user, err := suite.repo.FindByID(id)
		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), model.StatusInactive, user.Status)
	}
}

// Test Email Validation
func (suite *UserRepositoryTestSuite) TestEmailValidation() {
	testCases := []struct {
		email   string
		isValid bool
	}{
		{"test@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.co.uk", true},
		{"", false},
		{"notanemail", false},
		{"@example.com", false},
		{"test@", false},
		{"test @example.com", false},
	}

	for _, tc := range testCases {
		user := &model.User{
			Email:    tc.email,
			Password: "password123",
			Role:     model.RoleEmployee,
			Status:   model.StatusActive,
		}

		err := suite.repo.ValidateUser(user, false)
		if tc.isValid {
			assert.NoError(suite.T(), err, "Email %s should be valid", tc.email)
		} else {
			assert.Error(suite.T(), err, "Email %s should be invalid", tc.email)
		}
	}
}

// Test Indexes
func (suite *UserRepositoryTestSuite) TestIndexes() {
	// Indexes should be created in SetupTest

	// Test unique email index by trying to insert duplicate
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert first document directly
	_, err := suite.collection.InsertOne(ctx, bson.M{
		"email":  "test@example.com",
		"role":   "employee",
		"status": "active",
	})
	require.NoError(suite.T(), err)

	// Try to insert duplicate
	_, err = suite.collection.InsertOne(ctx, bson.M{
		"email":  "test@example.com",
		"role":   "admin",
		"status": "active",
	})
	assert.Error(suite.T(), err)
	assert.True(suite.T(), mongo.IsDuplicateKeyError(err))
}

// Run the test suite
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
