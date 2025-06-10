package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserEmailTaken    = errors.New("user email already taken")
	ErrInvalidUserData   = errors.New("invalid user data")
	ErrWeakPassword      = errors.New("password is too weak")
	ErrInvalidRole       = errors.New("invalid user role")
	ErrInvalidUserStatus = errors.New("invalid user status")
)

// ImprovedUserRepository provides enhanced user operations with logging and validation
type ImprovedUserRepository struct {
	*BaseRepository
	collection *mongo.Collection
	logger     *slog.Logger
}

// NewImprovedUserRepository creates a new improved user repository
func NewImprovedUserRepository() *ImprovedUserRepository {
	collection := db.GetCollection("users")
	return &ImprovedUserRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
		logger:         utils.GetLogger(),
	}
}

// ValidateUser validates user data
func (r *ImprovedUserRepository) ValidateUser(user *model.User, isUpdate bool) error {
	// Basic field validation for new users
	if !isUpdate {
		if strings.TrimSpace(user.FirstName) == "" {
			return fmt.Errorf("%w: first name cannot be empty", ErrInvalidUserData)
		}
		if strings.TrimSpace(user.LastName) == "" {
			return fmt.Errorf("%w: last name cannot be empty", ErrInvalidUserData)
		}
		if strings.TrimSpace(user.Email) == "" {
			return fmt.Errorf("%w: email cannot be empty", ErrInvalidUserData)
		}
		if user.Password == "" {
			return fmt.Errorf("%w: password cannot be empty", ErrInvalidUserData)
		}
	}

	// Email validation
	if user.Email != "" && !strings.Contains(user.Email, "@") {
		return fmt.Errorf("%w: invalid email format", ErrInvalidUserData)
	}

	// Password validation (only if password is provided)
	if user.Password != "" {
		if len(user.Password) < 6 {
			return fmt.Errorf("%w: password must be at least 6 characters", ErrWeakPassword)
		}
	}

	// Role validation
	if user.Role != "" {
		switch user.Role {
		case model.RoleAdmin, model.RoleEmployee:
			// Valid roles
		default:
			return fmt.Errorf("%w: %s", ErrInvalidRole, user.Role)
		}
	}

	// Status validation
	if user.Status != "" {
		switch user.Status {
		case model.StatusActive, model.StatusInactive:
			// Valid statuses
		default:
			return fmt.Errorf("%w: %s", ErrInvalidUserStatus, user.Status)
		}
	}

	return nil
}

// Create creates a new user with validation and logging
func (r *ImprovedUserRepository) Create(user *model.User) error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.Create")
	defer perf.End("email", user.Email, "role", user.Role)

	// Validate user data
	if err := r.ValidateUser(user, false); err != nil {
		perf.EndWithError(err)
		return err
	}

	// Check if email already exists
	exists, err := r.EmailExists(user.Email)
	if err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		err := ErrUserEmailTaken
		perf.EndWithError(err)
		return err
	}

	// Use transaction for atomic operation
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Set timestamps
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now

		// Set default status if not provided
		if user.Status == "" {
			user.Status = model.StatusActive
		}

		// Normalize email
		user.Email = strings.ToLower(strings.TrimSpace(user.Email))

		// Hash password
		if err := user.HashPassword(); err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Insert user
		result, err := r.collection.InsertOne(sessCtx, user)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return ErrUserEmailTaken
			}
			return fmt.Errorf("failed to insert user: %w", err)
		}

		user.ID = result.InsertedID.(primitive.ObjectID)

		r.logger.Info("User created successfully",
			"user_id", user.ID.Hex(),
			"email", user.Email,
			"role", user.Role,
		)

		return nil
	})
}

// FindByID finds a user by ID with proper error handling
func (r *ImprovedUserRepository) FindByID(id string) (*model.User, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.FindByID")
	defer perf.End("user_id", id)

	var user model.User
	err := r.BaseRepository.FindByID(id, &user)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrUserNotFound
		}
		perf.EndWithError(err)
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email with case-insensitive search
func (r *ImprovedUserRepository) FindByEmail(email string) (*model.User, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.FindByEmail")
	defer perf.End("email", email)

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		err := fmt.Errorf("%w: email cannot be empty", ErrInvalidUserData)
		perf.EndWithError(err)
		return nil, err
	}

	var user model.User
	filter := bson.M{
		"email":  email,
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	}

	err := r.FindOne(filter, &user)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrUserNotFound
		}
		perf.EndWithError(err)
		return nil, err
	}

	return &user, nil
}

// FindAll finds all active users with pagination
func (r *ImprovedUserRepository) FindAll(skip, limit int64) ([]*model.User, int64, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.FindAll")
	defer perf.End("skip", skip, "limit", limit)

	var users []*model.User

	// Set up options
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	// Only find active users
	filter := bson.M{
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	}

	err := r.BaseRepository.FindAll(filter, &users, findOptions)
	if err != nil {
		perf.EndWithError(err)
		return nil, 0, err
	}

	// Get total count
	total, err := r.Count(filter)
	if err != nil {
		perf.EndWithError(err)
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates a user with validation
func (r *ImprovedUserRepository) Update(user *model.User) error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.Update")
	defer perf.End("user_id", user.ID.Hex(), "email", user.Email)

	// Validate user data
	if err := r.ValidateUser(user, true); err != nil {
		perf.EndWithError(err)
		return err
	}

	// Check if email is taken by another user
	if user.Email != "" {
		var existing model.User
		err := r.FindOne(bson.M{
			"email": strings.ToLower(strings.TrimSpace(user.Email)),
			"_id":   bson.M{"$ne": user.ID},
			"status": model.StatusActive,
		}, &existing)

		if err == nil {
			err := ErrUserEmailTaken
			perf.EndWithError(err)
			return err
		} else if !errors.Is(err, ErrNotFound) {
			perf.EndWithError(err)
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}

	user.UpdatedAt = time.Now()

	// Build update document dynamically
	updateDoc := bson.M{
		"$set": bson.M{
			"updatedAt": user.UpdatedAt,
		},
	}
	setFields := updateDoc["$set"].(bson.M)

	// Only update non-zero fields
	if user.FirstName != "" {
		setFields["firstName"] = user.FirstName
	}
	if user.LastName != "" {
		setFields["lastName"] = user.LastName
	}
	if user.Email != "" {
		setFields["email"] = strings.ToLower(strings.TrimSpace(user.Email))
	}
	if user.Role != "" {
		setFields["role"] = user.Role
	}
	if user.Status != "" {
		setFields["status"] = user.Status
	}

	// Handle password update
	if user.Password != "" {
		if err := user.HashPassword(); err != nil {
			err := fmt.Errorf("failed to hash password: %w", err)
			perf.EndWithError(err)
			return err
		}
		setFields["passwordHash"] = user.PasswordHash
	}

	err := r.UpdateByID(user.ID.Hex(), updateDoc)
	if err != nil {
		perf.EndWithError(err)
		return err
	}

	r.logger.Info("User updated successfully",
		"user_id", user.ID.Hex(),
		"email", user.Email,
	)

	return nil
}

// Delete performs a soft delete on a user
func (r *ImprovedUserRepository) Delete(id string) error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.Delete")
	defer perf.End("user_id", id)

	objID, err := r.ValidateObjectID(id)
	if err != nil {
		perf.EndWithError(err)
		return err
	}

	// Soft delete: mark as inactive and set deletedAt timestamp
	update := bson.M{
		"$set": bson.M{
			"status":    model.StatusInactive,
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"_id": objID}, update)
	if err != nil {
		perf.EndWithError(err)
		return err
	}

	if result.MatchedCount == 0 {
		err := ErrUserNotFound
		perf.EndWithError(err)
		return err
	}

	r.logger.Info("User deleted successfully",
		"user_id", id,
	)

	return nil
}

// FindByRole finds all users with a specific role
func (r *ImprovedUserRepository) FindByRole(role model.UserRole) ([]*model.User, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.FindByRole")
	defer perf.End("role", role)

	var users []*model.User

	filter := bson.M{
		"role":   role,
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	}

	err := r.BaseRepository.FindAll(filter, &users, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}

	return users, nil
}

// EmailExists checks if an email already exists for an active user
func (r *ImprovedUserRepository) EmailExists(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, fmt.Errorf("%w: email cannot be empty", ErrInvalidUserData)
	}

	filter := bson.M{
		"email":  email,
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	}

	return r.Exists(filter)
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *ImprovedUserRepository) UpdateLastLogin(userID string) error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.UpdateLastLogin")
	defer perf.End("user_id", userID)

	objID, err := r.ValidateObjectID(userID)
	if err != nil {
		perf.EndWithError(err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"lastLogin": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"_id": objID}, update)
	if err != nil {
		perf.EndWithError(err)
		return err
	}

	if result.MatchedCount == 0 {
		err := ErrUserNotFound
		perf.EndWithError(err)
		return err
	}

	return nil
}

// CreateAdminUserIfNotExists creates an admin user if none exists
func (r *ImprovedUserRepository) CreateAdminUserIfNotExists() error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.CreateAdminUserIfNotExists")
	defer perf.End()

	// Check if admin user already exists
	adminUsers, err := r.FindByRole(model.RoleAdmin)
	if err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to check admin users: %w", err)
	}

	// If admin exists, nothing to do
	if len(adminUsers) > 0 {
		r.logger.Debug("Admin user already exists, skipping creation")
		return nil
	}

	// Create admin user
	admin := &model.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@peopleflow.com",
		Password:  "admin",
		Role:      model.RoleAdmin,
		Status:    model.StatusActive,
	}

	err = r.Create(admin)
	if err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	r.logger.Info("Default admin user created successfully",
		"email", admin.Email,
	)

	return nil
}

// GetUsersByEmployeeIDs finds users by their employee IDs
func (r *ImprovedUserRepository) GetUsersByEmployeeIDs(employeeIDs []primitive.ObjectID) ([]*model.User, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.GetUsersByEmployeeIDs")
	defer perf.End("count", len(employeeIDs))

	var users []*model.User

	filter := bson.M{
		"employeeId": bson.M{"$in": employeeIDs},
		"status":     model.StatusActive,
		"deletedAt":  bson.M{"$exists": false},
	}

	err := r.BaseRepository.FindAll(filter, &users, nil)
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}

	return users, nil
}

// GetUserStatistics returns user statistics
func (r *ImprovedUserRepository) GetUserStatistics() (map[string]interface{}, error) {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.GetUserStatistics")
	defer perf.End()

	stats := make(map[string]interface{})

	// Total active users
	totalActive, err := r.Count(bson.M{
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	})
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}
	stats["total_active"] = totalActive

	// Total inactive users
	totalInactive, err := r.Count(bson.M{
		"status": model.StatusInactive,
	})
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}
	stats["total_inactive"] = totalInactive

	// Users by role
	adminCount, err := r.Count(bson.M{
		"role":   model.RoleAdmin,
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	})
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}
	stats["admin_count"] = adminCount

	employeeCount, err := r.Count(bson.M{
		"role":   model.RoleEmployee,
		"status": model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	})
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}
	stats["employee_count"] = employeeCount

	// Recent logins (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentLogins, err := r.Count(bson.M{
		"lastLogin": bson.M{"$gte": thirtyDaysAgo},
		"status":    model.StatusActive,
		"deletedAt": bson.M{"$exists": false},
	})
	if err != nil {
		perf.EndWithError(err)
		return nil, err
	}
	stats["recent_logins"] = recentLogins

	r.logger.Debug("User statistics calculated",
		"total_active", totalActive,
		"total_inactive", totalInactive,
		"admin_count", adminCount,
		"employee_count", employeeCount,
		"recent_logins", recentLogins,
	)

	return stats, nil
}

// CreateIndexes creates necessary indexes for the users collection
func (r *ImprovedUserRepository) CreateIndexes() error {
	perf := utils.StartPerformanceLogging(nil, "UserRepository.CreateIndexes")
	defer perf.End()

	// Unique index on email
	if err := r.CreateIndex(bson.M{"email": 1}, true); err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create email index: %w", err)
	}

	// Index on status for queries
	if err := r.CreateIndex(bson.M{"status": 1}, false); err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Index on role for role-based queries
	if err := r.CreateIndex(bson.M{"role": 1}, false); err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create role index: %w", err)
	}

	// Index on employeeId
	if err := r.CreateIndex(bson.M{"employeeId": 1}, false); err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create employeeId index: %w", err)
	}

	// Compound index for active user queries
	if err := r.CreateIndex(bson.M{"status": 1, "deletedAt": 1}, false); err != nil {
		perf.EndWithError(err)
		return fmt.Errorf("failed to create status+deletedAt index: %w", err)
	}

	r.logger.Info("User repository indexes created successfully")
	return nil
}