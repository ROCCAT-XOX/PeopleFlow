// backend/repository/employee_repository_improved.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// EmployeeRepository errors
var (
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrEmployeeNumberTaken  = errors.New("employee number already taken")
	ErrInvalidEmployeeData  = errors.New("invalid employee data")
	ErrInvalidContractType  = errors.New("invalid contract type")
	ErrInvalidWeeklyHours   = errors.New("weekly hours must be between 0 and 60")
	ErrInvalidVacationDays  = errors.New("vacation days must be non-negative")
	ErrInvalidOvertimeData  = errors.New("invalid overtime data")
	ErrEmployeeEmailTaken   = errors.New("employee email already taken")
	ErrInsufficientVacation = errors.New("insufficient vacation days")
	ErrExcessiveVacation    = errors.New("vacation days exceed limit")
	ErrOvertimeOutOfBounds  = errors.New("overtime balance out of allowed bounds")
)

// EmployeeRepository enthält alle Datenbankoperationen für das Employee-Modell
type EmployeeRepository struct {
	*BaseRepository
	collection *mongo.Collection
	userRepo   *UserRepository
	logger     *slog.Logger
}

// NewEmployeeRepository erstellt ein neues EmployeeRepository
func NewEmployeeRepository() *EmployeeRepository {
	collection := db.GetCollection("employees")
	baseRepo := NewBaseRepository(collection)

	repo := &EmployeeRepository{
		BaseRepository: baseRepo,
		collection:     collection,
		userRepo:       NewUserRepository(),
		logger:         slog.Default().With("repository", "employee"),
	}

	return repo
}

// WithContext creates a new repository instance with context logger
func (r *EmployeeRepository) WithContext(ctx context.Context) *EmployeeRepository {
	logger := r.logger
	if reqID, ok := ctx.Value("requestID").(string); ok {
		logger = logger.With("requestID", reqID)
	}
	if userID, ok := ctx.Value("userID").(string); ok {
		logger = logger.With("userID", userID)
	}

	newRepo := *r
	newRepo.logger = logger
	return &newRepo
}

// ValidateEmployee validates employee data comprehensively
func (r *EmployeeRepository) ValidateEmployee(employee *model.Employee, isUpdate bool) error {
	r.logger.Debug("Validating employee", "isUpdate", isUpdate)

	// Basic field validation for new employees
	if !isUpdate {
		if strings.TrimSpace(employee.FirstName) == "" {
			return fmt.Errorf("%w: first name cannot be empty", ErrInvalidEmployeeData)
		}
		if strings.TrimSpace(employee.LastName) == "" {
			return fmt.Errorf("%w: last name cannot be empty", ErrInvalidEmployeeData)
		}
		if strings.TrimSpace(employee.EmployeeNumber) == "" {
			return fmt.Errorf("%w: employee number cannot be empty", ErrInvalidEmployeeData)
		}
		// Validate employee number format (example: EMP001)
		if !isValidEmployeeNumber(employee.EmployeeNumber) {
			return fmt.Errorf("%w: employee number must match pattern EMP[0-9]+", ErrInvalidEmployeeData)
		}
	}

	// Email validation if provided
	if employee.Email != "" && !isValidEmail(employee.Email) {
		return fmt.Errorf("%w: invalid email format", ErrInvalidEmployeeData)
	}

	// Contract type validation
	if employee.ContractType != "" {
		if !employee.ContractType.IsValid() {
			return fmt.Errorf("%w: %s", ErrInvalidContractType, employee.ContractType)
		}
	}

	// Weekly hours validation
	if employee.WeeklyHours < 0 || employee.WeeklyHours > 60 {
		return ErrInvalidWeeklyHours
	}

	// Vacation days validation
	if employee.VacationDaysPerYear < 0 || employee.VacationDaysPerYear > 365 {
		return fmt.Errorf("%w: must be between 0 and 365", ErrInvalidVacationDays)
	}

	if employee.VacationDaysRemaining < 0 {
		return fmt.Errorf("%w: remaining days cannot be negative", ErrInvalidVacationDays)
	}

	// Overtime balance validation
	if employee.OvertimeBalance < -200 || employee.OvertimeBalance > 200 {
		return fmt.Errorf("%w: overtime balance must be between -200 and 200 hours", ErrInvalidOvertimeData)
	}

	// Start date validation
	if !isUpdate && !employee.StartDate.IsZero() && employee.StartDate.After(time.Now().AddDate(1, 0, 0)) {
		return fmt.Errorf("%w: start date cannot be more than 1 year in the future", ErrInvalidEmployeeData)
	}

	r.logger.Debug("Employee validation passed")
	return nil
}

// Create erstellt einen neuen Mitarbeiter mit umfassender Validierung und Transaktion
func (r *EmployeeRepository) Create(employee *model.Employee) error {
	startTime := time.Now()
	r.logger.Info("Creating new employee", "employeeNumber", employee.EmployeeNumber)

	// Validate employee data
	if err := r.ValidateEmployee(employee, false); err != nil {
		r.logger.Error("Employee validation failed", "error", err)
		return err
	}

	// Check if employee number already exists
	exists, err := r.EmployeeNumberExists(employee.EmployeeNumber)
	if err != nil {
		r.logger.Error("Failed to check employee number existence", "error", err)
		return fmt.Errorf("failed to check employee number: %w", err)
	}
	if exists {
		r.logger.Info("Employee number already taken", "employeeNumber", employee.EmployeeNumber)
		return ErrEmployeeNumberTaken
	}

	// Check if email already exists (if provided)
	if employee.Email != "" {
		emailExists, err := r.EmailExists(employee.Email)
		if err != nil {
			r.logger.Error("Failed to check email existence", "error", err)
			return fmt.Errorf("failed to check email: %w", err)
		}
		if emailExists {
			r.logger.Info("Employee email already taken", "email", employee.Email)
			return ErrEmployeeEmailTaken
		}
	}

	// Use transaction to create employee and user atomically
	err = r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Set timestamps
		now := time.Now()
		employee.CreatedAt = now
		employee.UpdatedAt = now
		employee.Active = true

		// Set default values
		if employee.VacationDaysRemaining == 0 && employee.VacationDaysPerYear > 0 {
			employee.VacationDaysRemaining = employee.VacationDaysPerYear
		}

		// Set default working hours if not specified
		if employee.WeeklyHours == 0 {
			employee.WeeklyHours = 40 // Default full-time
		}

		// Initialize empty slices to avoid nil
		if employee.Documents == nil {
			employee.Documents = []model.Document{}
		}
		if employee.Absences == nil {
			employee.Absences = []model.Absence{}
		}
		if employee.TimeEntries == nil {
			employee.TimeEntries = []model.TimeEntry{}
		}

		// Insert employee
		result, err := r.collection.InsertOne(sessCtx, employee)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return ErrEmployeeNumberTaken
			}
			return fmt.Errorf("failed to insert employee: %w", err)
		}

		employee.ID = result.InsertedID.(primitive.ObjectID)
		r.logger.Debug("Employee document inserted", "id", employee.ID.Hex())

		// Create corresponding user if email is provided
		if employee.Email != "" {
			user := &model.User{
				Email:      employee.Email,
				FirstName:  employee.FirstName,
				LastName:   employee.LastName,
				Role:       model.RoleEmployee,
				Status:     model.StatusActive,
				EmployeeID: employee.ID,
				Password:   generateTemporaryPassword(), // Generate secure temporary password
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			// Hash password
			if err := user.HashPassword(); err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}

			userCollection := db.GetCollection("users")
			if _, err := userCollection.InsertOne(sessCtx, user); err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return fmt.Errorf("user with email %s already exists", employee.Email)
				}
				return fmt.Errorf("failed to create user: %w", err)
			}

			r.logger.Debug("User account created for employee", "email", employee.Email)
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Failed to create employee", "error", err, "duration", time.Since(startTime))
		return err
	}

	r.logger.Info("Employee created successfully",
		"id", employee.ID.Hex(),
		"employeeNumber", employee.EmployeeNumber,
		"duration", time.Since(startTime))

	return nil
}

// Update aktualisiert einen Mitarbeiter mit umfassender Validierung
func (r *EmployeeRepository) Update(employee *model.Employee) error {
	startTime := time.Now()
	r.logger.Info("Updating employee", "id", employee.ID.Hex())

	// Validate employee data
	if err := r.ValidateEmployee(employee, true); err != nil {
		r.logger.Error("Employee validation failed", "error", err)
		return err
	}

	// Use transaction for consistency
	err := r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Get current employee data first
		var currentEmployee model.Employee
		err := r.collection.FindOne(sessCtx, bson.M{"_id": employee.ID}).Decode(&currentEmployee)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrEmployeeNotFound
			}
			return fmt.Errorf("failed to find current employee: %w", err)
		}

		// Check if employee number is being changed and is taken by another employee
		if employee.EmployeeNumber != "" && employee.EmployeeNumber != currentEmployee.EmployeeNumber {
			var existing model.Employee
			err := r.collection.FindOne(sessCtx, bson.M{
				"employeeNumber": employee.EmployeeNumber,
				"_id":            bson.M{"$ne": employee.ID},
			}).Decode(&existing)

			if err == nil {
				return ErrEmployeeNumberTaken
			} else if err != mongo.ErrNoDocuments {
				return fmt.Errorf("failed to check employee number: %w", err)
			}
		}

		// Check if email is being changed
		if employee.Email != "" && employee.Email != currentEmployee.Email {
			var existing model.Employee
			err := r.collection.FindOne(sessCtx, bson.M{
				"email": employee.Email,
				"_id":   bson.M{"$ne": employee.ID},
			}).Decode(&existing)

			if err == nil {
				return ErrEmployeeEmailTaken
			} else if err != mongo.ErrNoDocuments {
				return fmt.Errorf("failed to check email: %w", err)
			}
		}

		employee.UpdatedAt = time.Now()

		// Build update document dynamically
		updateDoc := bson.M{
			"$set": bson.M{
				"updatedAt": employee.UpdatedAt,
			},
		}
		setFields := updateDoc["$set"].(bson.M)

		// Only update non-zero fields
		updateFields := map[string]interface{}{
			"firstName":             employee.FirstName,
			"lastName":              employee.LastName,
			"email":                 employee.Email,
			"employeeNumber":        employee.EmployeeID,
			"department":            employee.Department,
			"position":              employee.Position,
			"contractType":          employee.ContractType,
			"weeklyHours":           employee.WeeklyHours,
			"vacationDaysPerYear":   employee.VacationDaysPerYear,
			"vacationDaysRemaining": employee.VacationDaysRemaining,
			"overtimeBalance":       employee.OvertimeBalance,
		}

		for field, value := range updateFields {
			// Check if value is not zero value
			if !isZeroValue(value) {
				setFields[field] = value
			}
		}

		result, err := r.collection.UpdateOne(sessCtx, bson.M{"_id": employee.ID}, updateDoc)
		if err != nil {
			return fmt.Errorf("failed to update employee: %w", err)
		}

		if result.MatchedCount == 0 {
			return ErrEmployeeNotFound
		}

		// Update user if email changed
		if employee.Email != "" && employee.Email != currentEmployee.Email {
			userCollection := db.GetCollection("users")
			_, err = userCollection.UpdateOne(
				sessCtx,
				bson.M{"employeeId": employee.ID},
				bson.M{
					"$set": bson.M{
						"email":     employee.Email,
						"firstName": employee.FirstName,
						"lastName":  employee.LastName,
						"updatedAt": time.Now(),
					},
				},
			)
			if err != nil && err != mongo.ErrNoDocuments {
				return fmt.Errorf("failed to update user: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Failed to update employee", "error", err, "duration", time.Since(startTime))
		return err
	}

	r.logger.Info("Employee updated successfully",
		"id", employee.ID.Hex(),
		"duration", time.Since(startTime))

	return nil
}

// Helper functions
func isValidEmployeeNumber(number string) bool {
	// Simple pattern matching for employee number (customize as needed)
	return strings.HasPrefix(strings.ToUpper(number), "EMP") && len(number) >= 6
}

func isValidEmail(email string) bool {
	// Simple email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	return len(parts[0]) > 0 && strings.Contains(parts[1], ".")
}

func generateTemporaryPassword() string {
	// Generate a secure temporary password
	return fmt.Sprintf("Welcome@%d", time.Now().Unix()%10000)
}

func isZeroValue(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return val == ""
	case int, int64, float64:
		return val == 0
	case bool:
		return !val
	case time.Time:
		return val.IsZero()
	default:
		return false
	}
}

// EmailExists checks if an email already exists for an active employee
func (r *EmployeeRepository) EmailExists(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return r.Exists(bson.M{"email": email, "active": true})
}
