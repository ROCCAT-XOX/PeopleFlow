package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	userRepo       *UserRepository
	userCollection *mongo.Collection
	userTestClient *mongo.Client
)

func setupUserTest(t *testing.T) {
	// Initialize logger for testing
	err := utils.InitLogger(utils.LoggerConfig{
		Level:  utils.LogLevelDebug,
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to test database
	userTestClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	// Use test database
	testDB := userTestClient.Database("peopleflow_user_test")
	userCollection = testDB.Collection("users")
	
	// Create user repository with test collection
	userRepo = &UserRepository{
		BaseRepository: NewBaseRepository(userCollection),
		collection:     userCollection,
	}

	// Clean up any existing test data
	_, err = userCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to clean test collection: %v", err)
	}
}

func teardownUserTest(t *testing.T) {
	if userCollection != nil {
		// Clean up test data
		_, _ = userCollection.DeleteMany(context.Background(), bson.M{})
		_ = userCollection.Drop(context.Background())
	}
	if userTestClient != nil {
		_ = userTestClient.Disconnect(context.Background())
	}
}

func createTestUser() *model.User {
	return &model.User{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		Password:   "password123",
		Role:       model.RoleEmployee,
		Status:     model.StatusActive,
		EmployeeID: primitive.NewObjectID(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func TestUserRepository_ValidateUser(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	tests := []struct {
		name        string
		user        *model.User
		isUpdate    bool
		shouldError bool
	}{
		{
			name:        "valid new user",
			user:        createTestUser(),
			isUpdate:    false,
			shouldError: false,
		},
		{
			name: "empty first name",
			user: func() *model.User {
				user := createTestUser()
				user.FirstName = ""
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "empty last name",
			user: func() *model.User {
				user := createTestUser()
				user.LastName = ""
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "empty email",
			user: func() *model.User {
				user := createTestUser()
				user.Email = ""
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "invalid email format",
			user: func() *model.User {
				user := createTestUser()
				user.Email = "invalid-email"
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "empty password",
			user: func() *model.User {
				user := createTestUser()
				user.Password = ""
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "weak password",
			user: func() *model.User {
				user := createTestUser()
				user.Password = "123"
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "invalid role",
			user: func() *model.User {
				user := createTestUser()
				user.Role = "invalid_role"
				return user
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "valid update without password",
			user: func() *model.User {
				user := createTestUser()
				user.Password = ""
				return user
			}(),
			isUpdate:    true,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := userRepo.ValidateUser(tt.user, tt.isUpdate)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

func TestUserRepository_Create(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	t.Run("successful creation", func(t *testing.T) {
		user := createTestUser()
		err := userRepo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		if user.ID.IsZero() {
			t.Error("User ID should be set after creation")
		}

		if user.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set")
		}

		if user.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should be set")
		}

		// Password should be hashed
		if user.PasswordHash == "" {
			t.Error("Password should be hashed")
		}

		if user.Password != "" {
			t.Error("Plain text password should be cleared")
		}

		// Verify user was actually saved
		var savedUser model.User
		err = userRepo.FindByID(user.ID.Hex(), &savedUser)
		if err != nil {
			t.Fatalf("Failed to find created user: %v", err)
		}

		if savedUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, savedUser.Email)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		user1 := createTestUser()
		err := userRepo.Create(user1)
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		user2 := createTestUser()
		user2.Email = user1.Email // Same email
		err = userRepo.Create(user2)
		if err == nil {
			t.Error("Expected error for duplicate email")
		}
	})

	t.Run("invalid user data", func(t *testing.T) {
		user := createTestUser()
		user.Email = "invalid-email" // Invalid
		err := userRepo.Create(user)
		if err == nil {
			t.Error("Expected validation error for invalid user data")
		}
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("find existing user", func(t *testing.T) {
		foundUser, err := userRepo.FindByID(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find user: %v", err)
		}

		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}
		if foundUser.FirstName != user.FirstName {
			t.Errorf("Expected first name %s, got %s", user.FirstName, foundUser.FirstName)
		}
	})

	t.Run("find non-existent user", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		_, err := userRepo.FindByID(nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
	})

	t.Run("invalid ID format", func(t *testing.T) {
		_, err := userRepo.FindByID("invalid-id")
		if err == nil {
			t.Error("Expected error for invalid ID format")
		}
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("find existing user by email", func(t *testing.T) {
		foundUser, err := userRepo.FindByEmail(user.Email)
		if err != nil {
			t.Fatalf("Failed to find user by email: %v", err)
		}

		if foundUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, foundUser.Email)
		}
	})

	t.Run("find user by email case insensitive", func(t *testing.T) {
		foundUser, err := userRepo.FindByEmail(strings.ToUpper(user.Email))
		if err != nil {
			t.Fatalf("Failed to find user by email (case insensitive): %v", err)
		}

		if foundUser.Email != strings.ToLower(user.Email) {
			t.Errorf("Expected email %s, got %s", strings.ToLower(user.Email), foundUser.Email)
		}
	})

	t.Run("find non-existent user by email", func(t *testing.T) {
		_, err := userRepo.FindByEmail("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error for non-existent user email")
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := userRepo.FindByEmail("")
		if err == nil {
			t.Error("Expected error for empty email")
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		user.FirstName = "Jane"
		user.LastName = "Smith"
		user.Role = model.RoleAdmin

		err := userRepo.Update(user)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		// Verify update
		updatedUser, err := userRepo.FindByID(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find updated user: %v", err)
		}

		if updatedUser.FirstName != "Jane" {
			t.Errorf("Expected first name Jane, got %s", updatedUser.FirstName)
		}
		if updatedUser.LastName != "Smith" {
			t.Errorf("Expected last name Smith, got %s", updatedUser.LastName)
		}
		if updatedUser.Role != model.RoleAdmin {
			t.Errorf("Expected role admin, got %s", updatedUser.Role)
		}
	})

	t.Run("update password", func(t *testing.T) {
		oldPasswordHash := user.PasswordHash
		user.Password = "newpassword123"

		err := userRepo.Update(user)
		if err != nil {
			t.Fatalf("Failed to update user password: %v", err)
		}

		// Verify password hash changed
		updatedUser, err := userRepo.FindByID(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find updated user: %v", err)
		}

		if updatedUser.PasswordHash == oldPasswordHash {
			t.Error("Password hash should have changed")
		}

		// Verify new password works
		if !updatedUser.CheckPassword("newpassword123") {
			t.Error("New password should be valid")
		}
	})

	t.Run("invalid update data", func(t *testing.T) {
		user.Email = "invalid-email" // Invalid
		err := userRepo.Update(user)
		if err == nil {
			t.Error("Expected validation error for invalid update data")
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("successful soft delete", func(t *testing.T) {
		err := userRepo.Delete(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify user is marked as deleted
		var deletedUser model.User
		err = userRepo.collection.FindOne(context.Background(), bson.M{"_id": user.ID}).Decode(&deletedUser)
		if err != nil {
			t.Fatalf("Failed to find deleted user: %v", err)
		}

		if deletedUser.Status != model.StatusInactive {
			t.Errorf("Expected status inactive, got %s", deletedUser.Status)
		}

		if deletedUser.DeletedAt == nil {
			t.Error("DeletedAt should be set")
		}
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		err := userRepo.Delete(nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
	})

	t.Run("invalid ID format", func(t *testing.T) {
		err := userRepo.Delete("invalid-id")
		if err == nil {
			t.Error("Expected error for invalid ID format")
		}
	})
}

func TestUserRepository_FindByRole(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test users with different roles
	adminUser := createTestUser()
	adminUser.Email = "admin@example.com"
	adminUser.Role = model.RoleAdmin
	err := userRepo.Create(adminUser)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	employeeUser1 := createTestUser()
	employeeUser1.Email = "employee1@example.com"
	employeeUser1.Role = model.RoleEmployee
	err = userRepo.Create(employeeUser1)
	if err != nil {
		t.Fatalf("Failed to create employee user 1: %v", err)
	}

	employeeUser2 := createTestUser()
	employeeUser2.Email = "employee2@example.com"
	employeeUser2.Role = model.RoleEmployee
	err = userRepo.Create(employeeUser2)
	if err != nil {
		t.Fatalf("Failed to create employee user 2: %v", err)
	}

	t.Run("find admin users", func(t *testing.T) {
		adminUsers, err := userRepo.FindByRole(model.RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to find admin users: %v", err)
		}

		if len(adminUsers) != 1 {
			t.Errorf("Expected 1 admin user, got %d", len(adminUsers))
		}

		if adminUsers[0].Role != model.RoleAdmin {
			t.Errorf("Expected role admin, got %s", adminUsers[0].Role)
		}
	})

	t.Run("find employee users", func(t *testing.T) {
		employeeUsers, err := userRepo.FindByRole(model.RoleEmployee)
		if err != nil {
			t.Fatalf("Failed to find employee users: %v", err)
		}

		if len(employeeUsers) != 2 {
			t.Errorf("Expected 2 employee users, got %d", len(employeeUsers))
		}
	})
}

func TestUserRepository_EmailExists(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("existing email", func(t *testing.T) {
		exists, err := userRepo.EmailExists(user.Email)
		if err != nil {
			t.Fatalf("Failed to check email existence: %v", err)
		}

		if !exists {
			t.Error("Expected email to exist")
		}
	})

	t.Run("non-existing email", func(t *testing.T) {
		exists, err := userRepo.EmailExists("nonexistent@example.com")
		if err != nil {
			t.Fatalf("Failed to check email existence: %v", err)
		}

		if exists {
			t.Error("Expected email not to exist")
		}
	})

	t.Run("case insensitive check", func(t *testing.T) {
		exists, err := userRepo.EmailExists(strings.ToUpper(user.Email))
		if err != nil {
			t.Fatalf("Failed to check email existence (case insensitive): %v", err)
		}

		if !exists {
			t.Error("Expected email to exist (case insensitive)")
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := userRepo.EmailExists("")
		if err == nil {
			t.Error("Expected error for empty email")
		}
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	// Create test user
	user := createTestUser()
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("successful login update", func(t *testing.T) {
		err := userRepo.UpdateLastLogin(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to update last login: %v", err)
		}

		// Verify update
		updatedUser, err := userRepo.FindByID(user.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find updated user: %v", err)
		}

		if updatedUser.LastLogin == nil {
			t.Error("LastLogin should be set")
		}

		if time.Since(*updatedUser.LastLogin) > time.Minute {
			t.Error("LastLogin should be recent")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		err := userRepo.UpdateLastLogin(nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
	})
}

func TestUserRepository_CreateAdminUserIfNotExists(t *testing.T) {
	setupUserTest(t)
	defer teardownUserTest(t)

	t.Run("create admin when none exists", func(t *testing.T) {
		err := userRepo.CreateAdminUserIfNotExists()
		if err != nil {
			t.Fatalf("Failed to create admin user: %v", err)
		}

		// Verify admin user was created
		adminUsers, err := userRepo.FindByRole(model.RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to find admin users: %v", err)
		}

		if len(adminUsers) != 1 {
			t.Errorf("Expected 1 admin user, got %d", len(adminUsers))
		}

		if adminUsers[0].Email != "admin@peopleflow.com" {
			t.Errorf("Expected email admin@peopleflow.com, got %s", adminUsers[0].Email)
		}
	})

	t.Run("don't create admin when one exists", func(t *testing.T) {
		// Create another admin user
		adminUser := createTestUser()
		adminUser.Email = "another-admin@example.com"
		adminUser.Role = model.RoleAdmin
		err := userRepo.Create(adminUser)
		if err != nil {
			t.Fatalf("Failed to create admin user: %v", err)
		}

		// Call CreateAdminUserIfNotExists again
		err = userRepo.CreateAdminUserIfNotExists()
		if err != nil {
			t.Fatalf("Failed to run CreateAdminUserIfNotExists: %v", err)
		}

		// Should still have only 2 admin users (the default one + the one we created)
		adminUsers, err := userRepo.FindByRole(model.RoleAdmin)
		if err != nil {
			t.Fatalf("Failed to find admin users: %v", err)
		}

		if len(adminUsers) != 2 {
			t.Errorf("Expected 2 admin users, got %d", len(adminUsers))
		}
	})
}

// Test password hashing and verification
func TestUser_PasswordMethods(t *testing.T) {
	user := createTestUser()
	originalPassword := user.Password

	t.Run("hash password", func(t *testing.T) {
		err := user.HashPassword()
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		if user.PasswordHash == "" {
			t.Error("Password hash should be set")
		}

		if user.Password != "" {
			t.Error("Plain text password should be cleared")
		}
	})

	t.Run("check correct password", func(t *testing.T) {
		if !user.CheckPassword(originalPassword) {
			t.Error("CheckPassword should return true for correct password")
		}
	})

	t.Run("check incorrect password", func(t *testing.T) {
		if user.CheckPassword("wrongpassword") {
			t.Error("CheckPassword should return false for incorrect password")
		}
	})
}

// Benchmark tests
func BenchmarkUserRepository_Create(b *testing.B) {
	setupUserTest(&testing.T{})
	defer teardownUserTest(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := createTestUser()
		user.Email = fmt.Sprintf("bench%d@example.com", i)
		_ = userRepo.Create(user)
	}
}

func BenchmarkUserRepository_FindByEmail(b *testing.B) {
	setupUserTest(&testing.T{})
	defer teardownUserTest(&testing.T{})

	// Create test user
	user := createTestUser()
	_ = userRepo.Create(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = userRepo.FindByEmail(user.Email)
	}
}

func BenchmarkUser_CheckPassword(b *testing.B) {
	user := createTestUser()
	_ = user.HashPassword()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.CheckPassword("password123")
	}
}