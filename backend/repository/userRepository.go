// backend/repository/userRepository.go
package repository

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository errors
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already taken")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// UserRepository enthält alle Datenbankoperationen für das User-Modell
type UserRepository struct {
	*BaseRepository
	collection *mongo.Collection
}

// NewUserRepository erstellt ein neues UserRepository
func NewUserRepository() *UserRepository {
	collection := db.GetCollection("users")
	return &UserRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
	}
}

// ValidateUser validates user data before operations
func (r *UserRepository) ValidateUser(user *model.User, isUpdate bool) error {
	// Email validation
	if !isUpdate || user.Email != "" {
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if email == "" {
			return fmt.Errorf("%w: email cannot be empty", ErrValidation)
		}
		if !emailRegex.MatchString(email) {
			return fmt.Errorf("%w: %s", ErrInvalidEmail, email)
		}
		user.Email = email
	}

	// Password validation (only for create or if password is being updated)
	if !isUpdate && user.Password != "" {
		if len(user.Password) < 8 {
			return ErrInvalidPassword
		}
	}

	// Role validation
	if !isUpdate || user.Role != "" {
		if user.Role != model.RoleAdmin && user.Role != model.RoleEmployee && user.Role != model.RoleManager && user.Role != model.RoleHR {
			return fmt.Errorf("%w: invalid role %s", ErrValidation, user.Role)
		}
	}

	// Status validation
	if !isUpdate || user.Status != "" {
		if user.Status != model.StatusActive && user.Status != model.StatusInactive {
			return fmt.Errorf("%w: invalid status %s", ErrValidation, user.Status)
		}
	}

	return nil
}

// Create erstellt einen neuen Benutzer mit Validierung
func (r *UserRepository) Create(user *model.User) error {
	// Validate user data
	if err := r.ValidateUser(user, false); err != nil {
		return err
	}

	// Check if email already exists
	exists, err := r.EmailExists(user.Email)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return ErrEmailTaken
	}

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Hash password
	if err := user.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	id, err := r.InsertOne(user)
	if err != nil {
		if errors.Is(err, ErrDuplicateEntry) {
			return ErrEmailTaken
		}
		return err
	}

	user.ID = *id
	return nil
}

// FindByID findet einen Benutzer anhand seiner ID
func (r *UserRepository) FindByID(id string) (*model.User, error) {
	var user model.User
	err := r.BaseRepository.FindByID(id, &user)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail findet einen Benutzer anhand seiner E-Mail
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	var user model.User
	err := r.FindOne(bson.M{"email": email}, &user)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindAll findet alle Benutzer mit Pagination
func (r *UserRepository) FindAll(skip, limit int64) ([]*model.User, int64, error) {
	var users []*model.User

	// Set up options for pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.M{"createdAt": -1})

	err := r.BaseRepository.FindAll(bson.M{}, &users, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.Count(bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update aktualisiert einen Benutzer
func (r *UserRepository) Update(user *model.User) error {
	// Validate user data
	if err := r.ValidateUser(user, true); err != nil {
		return err
	}

	// Check if email is taken by another user
	if user.Email != "" {
		var existingUser model.User
		err := r.FindOne(bson.M{
			"email": user.Email,
			"_id":   bson.M{"$ne": user.ID},
		}, &existingUser)

		if err == nil {
			return ErrEmailTaken
		} else if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("failed to check email existence: %w", err)
		}
	}

	user.UpdatedAt = time.Now()

	// Prepare update document
	update := bson.M{
		"$set": bson.M{
			"updatedAt": user.UpdatedAt,
		},
	}

	// Only update non-empty fields
	if user.Email != "" {
		update["$set"].(bson.M)["email"] = user.Email
	}
	if user.FirstName != "" {
		update["$set"].(bson.M)["firstName"] = user.FirstName
	}
	if user.LastName != "" {
		update["$set"].(bson.M)["lastName"] = user.LastName
	}
	if user.Role != "" {
		update["$set"].(bson.M)["role"] = user.Role
	}
	if user.Status != "" {
		update["$set"].(bson.M)["status"] = user.Status
	}

	return r.UpdateByID(user.ID.Hex(), update)
}

// UpdatePassword aktualisiert das Passwort eines Benutzers
func (r *UserRepository) UpdatePassword(userID string, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidPassword
	}

	// Create temporary user to hash password
	tempUser := &model.User{Password: newPassword}
	if err := tempUser.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"passwordHash": tempUser.PasswordHash,
			"updatedAt":    time.Now(),
		},
	}

	return r.UpdateByID(userID, update)
}

// Delete löscht einen Benutzer (soft delete)
func (r *UserRepository) Delete(id string) error {
	update := bson.M{
		"$set": bson.M{
			"status":    model.StatusInactive,
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	return r.UpdateByID(id, update)
}

func (r *UserRepository) FindByRole(role model.UserRole) ([]*model.User, error) {
	var users []*model.User

	filter := bson.M{
		"role":   role,
		"status": model.StatusActive,
	}

	err := r.BaseRepository.FindAll(filter, &users, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		return nil, err
	}

	return users, nil
}

// EmailExists prüft, ob eine E-Mail bereits existiert
func (r *UserRepository) EmailExists(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return r.Exists(bson.M{"email": email})
}

// GetActiveUsersCount zählt aktive Benutzer
func (r *UserRepository) GetActiveUsersCount() (int64, error) {
	return r.Count(bson.M{"status": model.StatusActive})
}

// BulkUpdateStatus aktualisiert den Status mehrerer Benutzer
func (r *UserRepository) BulkUpdateStatus(userIDs []string, status model.UserStatus) error {
	var objectIDs []primitive.ObjectID

	// Convert string IDs to ObjectIDs
	for _, id := range userIDs {
		objID, err := r.ValidateObjectID(id)
		if err != nil {
			return fmt.Errorf("invalid user ID %s: %w", id, err)
		}
		objectIDs = append(objectIDs, *objID)
	}

	// Use transaction for bulk update
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		update := bson.M{
			"$set": bson.M{
				"status":    status,
				"updatedAt": time.Now(),
			},
		}

		_, err := r.collection.UpdateMany(
			sessCtx,
			bson.M{"_id": bson.M{"$in": objectIDs}},
			update,
		)

		return err
	})
}

// CreateIndexes erstellt erforderliche Indizes
func (r *UserRepository) CreateIndexes() error {
	// Unique index on email
	if err := r.CreateIndex(bson.M{"email": 1}, true); err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}

	// Index on role and status for queries
	if err := r.CreateIndex(bson.M{"role": 1, "status": 1}, false); err != nil {
		return fmt.Errorf("failed to create role-status index: %w", err)
	}

	// Index on createdAt for sorting
	if err := r.CreateIndex(bson.M{"createdAt": -1}, false); err != nil {
		return fmt.Errorf("failed to create createdAt index: %w", err)
	}

	return nil
}

// CreateAdminUserIfNotExists erstellt einen Admin-Benutzer, falls keiner existiert
func (r *UserRepository) CreateAdminUserIfNotExists() error {
	// Prüfen, ob bereits ein Admin-Benutzer existiert
	count, err := r.Count(bson.M{"role": model.RoleAdmin})
	if err != nil {
		return err
	}

	// Wenn bereits ein Admin existiert, nichts tun
	if count > 0 {
		return nil
	}

	// Admin-Benutzer erstellen
	admin := &model.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@peopleflow.com",
		Password:  "admin",
		Role:      model.RoleAdmin,
		Status:    model.StatusActive,
	}

	return r.Create(admin)
}
