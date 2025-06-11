// backend/repository/employeeRepository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"PeopleFlow/backend/db"
	"PeopleFlow/backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EmployeeRepository errors
var (
	ErrEmployeeNotFound    = errors.New("employee not found")
	ErrEmployeeIDTaken     = errors.New("employee ID already taken")
	ErrInvalidEmployeeData = errors.New("invalid employee data")
	ErrInvalidVacationDays = errors.New("vacation days must be non-negative")
	ErrInvalidOvertimeData = errors.New("invalid overtime data")
)

// EmployeeRepository enthält alle Datenbankoperationen für das Employee-Modell
type EmployeeRepository struct {
	*BaseRepository
	collection *mongo.Collection
	userRepo   *UserRepository
}

// NewEmployeeRepository erstellt ein neues EmployeeRepository
func NewEmployeeRepository() *EmployeeRepository {
	collection := db.GetCollection("employees")
	return &EmployeeRepository{
		BaseRepository: NewBaseRepository(collection),
		collection:     collection,
		userRepo:       NewUserRepository(),
	}
}

// ValidateEmployee validates employee data
func (r *EmployeeRepository) ValidateEmployee(employee *model.Employee, isUpdate bool) error {
	// Basic field validation for new employees
	if !isUpdate {
		if strings.TrimSpace(employee.FirstName) == "" {
			return fmt.Errorf("%w: first name cannot be empty", ErrInvalidEmployeeData)
		}
		if strings.TrimSpace(employee.LastName) == "" {
			return fmt.Errorf("%w: last name cannot be empty", ErrInvalidEmployeeData)
		}
		if strings.TrimSpace(employee.EmployeeID) == "" {
			return fmt.Errorf("%w: employee ID cannot be empty", ErrInvalidEmployeeData)
		}
	}

	// Email validation if provided
	if employee.Email != "" && !strings.Contains(employee.Email, "@") {
		return fmt.Errorf("%w: invalid email format", ErrInvalidEmployeeData)
	}

	// Working hours validation
	if employee.WorkingHoursPerWeek < 0 || employee.WorkingHoursPerWeek > 60 {
		return fmt.Errorf("%w: working hours must be between 0 and 60", ErrInvalidEmployeeData)
	}

	// Vacation days validation
	if employee.VacationDays < 0 || employee.VacationDays > 365 {
		return fmt.Errorf("%w: must be between 0 and 365", ErrInvalidVacationDays)
	}

	if employee.RemainingVacation < 0 {
		return fmt.Errorf("%w: remaining vacation cannot be negative", ErrInvalidVacationDays)
	}

	// Overtime balance validation
	if employee.OvertimeBalance < -200 || employee.OvertimeBalance > 200 {
		return fmt.Errorf("%w: overtime balance must be between -200 and 200 hours", ErrInvalidOvertimeData)
	}

	return nil
}

// Create erstellt einen neuen Mitarbeiter mit Validierung und Transaktion
func (r *EmployeeRepository) Create(employee *model.Employee) error {
	// Validate employee data
	if err := r.ValidateEmployee(employee, false); err != nil {
		return err
	}

	// Check if employee ID already exists
	exists, err := r.EmployeeIDExists(employee.EmployeeID)
	if err != nil {
		return fmt.Errorf("failed to check employee ID: %w", err)
	}
	if exists {
		return ErrEmployeeIDTaken
	}

	// Try to use transaction, but fall back to direct creation if transactions aren't supported
	return r.tryTransactionOrDirect(func(ctx context.Context) error {
		// Set timestamps
		now := time.Now()
		employee.CreatedAt = now
		employee.UpdatedAt = now

		// Set default status if not provided
		if employee.Status == "" {
			employee.Status = model.EmployeeStatusActive
		}

		// Set default values
		if employee.RemainingVacation == 0 && employee.VacationDays > 0 {
			employee.RemainingVacation = employee.VacationDays
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
		if employee.WeeklyTimeEntries == nil {
			employee.WeeklyTimeEntries = []model.WeeklyTimeEntry{}
		}
		if employee.OvertimeAdjustments == nil {
			employee.OvertimeAdjustments = []model.OvertimeAdjustment{}
		}
		if employee.ApplicationDocuments == nil {
			employee.ApplicationDocuments = []model.Document{}
		}
		if employee.Trainings == nil {
			employee.Trainings = []model.Training{}
		}
		if employee.Evaluations == nil {
			employee.Evaluations = []model.Evaluation{}
		}
		if employee.DevelopmentPlan == nil {
			employee.DevelopmentPlan = []model.DevelopmentItem{}
		}
		if employee.Conversations == nil {
			employee.Conversations = []model.Conversation{}
		}
		if employee.ProjectAssignments == nil {
			employee.ProjectAssignments = []model.ProjectAssignment{}
		}

		// Insert employee
		result, err := r.collection.InsertOne(ctx, employee)
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return ErrEmployeeIDTaken
			}
			return fmt.Errorf("failed to insert employee: %w", err)
		}

		employee.ID = result.InsertedID.(primitive.ObjectID)

		// Create corresponding user if email is provided
		if employee.Email != "" {
			user := &model.User{
				Email:      employee.Email,
				FirstName:  employee.FirstName,
				LastName:   employee.LastName,
				Role:       model.RoleEmployee,
				Status:     model.StatusActive,
				EmployeeID: &employee.ID,
				Password:   "changeme123", // Temporary password
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			// Hash password
			if err := user.HashPassword(); err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}

			userCollection := db.GetCollection("users")
			if _, err := userCollection.InsertOne(ctx, user); err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return fmt.Errorf("user with email %s already exists", employee.Email)
				}
				return fmt.Errorf("failed to create user: %w", err)
			}
		}

		return nil
	})
}

// tryTransactionOrDirect attempts to use a transaction, but falls back to direct execution if transactions are not supported
func (r *EmployeeRepository) tryTransactionOrDirect(fn func(context.Context) error) error {
	// First try with transaction
	err := r.Transaction(func(sessCtx mongo.SessionContext) error {
		return fn(sessCtx)
	})
	
	// If transaction failed due to not being supported, try without transaction
	if err != nil && (strings.Contains(err.Error(), "Transaction numbers are only allowed") || 
		strings.Contains(err.Error(), "IllegalOperation")) {
		// Log that we're falling back to non-transactional operation
		fmt.Printf("Transactions not supported, falling back to direct operations: %v\n", err)
		
		// Fall back to direct execution without transaction
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return fn(ctx)
	}
	
	return err
}

// FindByID findet einen Mitarbeiter anhand seiner MongoDB ID
func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var employee model.Employee
	err := r.BaseRepository.FindByID(id, &employee)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// FindByEmployeeID findet einen Mitarbeiter anhand seiner Employee ID
func (r *EmployeeRepository) FindByEmployeeID(employeeID string) (*model.Employee, error) {
	employeeID = strings.TrimSpace(employeeID)
	if employeeID == "" {
		return nil, fmt.Errorf("%w: employee ID cannot be empty", ErrInvalidEmployeeData)
	}

	var employee model.Employee
	err := r.FindOne(bson.M{"employeeId": employeeID}, &employee)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	return &employee, nil
}

// FindAll findet alle aktiven Mitarbeiter mit Pagination und Sortierung
func (r *EmployeeRepository) FindAll(skip, limit int64, sortBy string, sortOrder int) ([]*model.Employee, int64, error) {
	var employees []*model.Employee

	// Build sort options
	sortOptions := bson.M{"lastName": 1} // Default sort
	if sortBy != "" {
		sortOptions = bson.M{sortBy: sortOrder}
	}

	// Set up options
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(sortOptions)

	// Only find active employees by default
	filter := bson.M{"status": model.EmployeeStatusActive}

	err := r.BaseRepository.FindAll(filter, &employees, findOptions)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// Update aktualisiert einen Mitarbeiter mit Validierung
func (r *EmployeeRepository) Update(employee *model.Employee) error {
	// Validate employee data
	if err := r.ValidateEmployee(employee, true); err != nil {
		return err
	}

	// Check if employee ID is taken by another employee
	if employee.EmployeeID != "" {
		var existing model.Employee
		err := r.FindOne(bson.M{
			"employeeId": employee.EmployeeID,
			"_id":        bson.M{"$ne": employee.ID},
		}, &existing)

		if err == nil {
			return ErrEmployeeIDTaken
		} else if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("failed to check employee ID: %w", err)
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
	if employee.FirstName != "" {
		setFields["firstName"] = employee.FirstName
	}
	if employee.LastName != "" {
		setFields["lastName"] = employee.LastName
	}
	if employee.Email != "" {
		setFields["email"] = employee.Email
	}
	if employee.Department != "" {
		setFields["department"] = employee.Department
	}
	if employee.Position != "" {
		setFields["position"] = employee.Position
	}
	if employee.WorkingHoursPerWeek > 0 {
		setFields["workingHoursPerWeek"] = employee.WorkingHoursPerWeek
	}
	if employee.VacationDays >= 0 {
		setFields["vacationDays"] = employee.VacationDays
	}
	if employee.RemainingVacation >= 0 {
		setFields["remainingVacation"] = employee.RemainingVacation
	}

	// Update time entries if provided
	if employee.TimeEntries != nil {
		setFields["timeEntries"] = employee.TimeEntries
	}

	// Update other arrays if provided
	if employee.Absences != nil {
		setFields["absences"] = employee.Absences
	}
	if employee.WeeklyTimeEntries != nil {
		setFields["weeklyTimeEntries"] = employee.WeeklyTimeEntries
	}
	if employee.OvertimeAdjustments != nil {
		setFields["overtimeAdjustments"] = employee.OvertimeAdjustments
	}
	if employee.ProjectAssignments != nil {
		setFields["projectAssignments"] = employee.ProjectAssignments
	}

	// Update integration IDs
	if employee.TimebutlerUserID != "" {
		setFields["timebutlerUserId"] = employee.TimebutlerUserID
	}
	if employee.Erfasst123ID != "" {
		setFields["erfasst123Id"] = employee.Erfasst123ID
	}

	// Update overtime balance
	setFields["overtimeBalance"] = employee.OvertimeBalance

	return r.UpdateByID(employee.ID.Hex(), updateDoc)
}

// UpdateOvertimeBalance aktualisiert den Überstundensaldo eines Mitarbeiters
func (r *EmployeeRepository) UpdateOvertimeBalance(employeeID string, hours float64, reason string) error {
	// Validate input
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return err
	}

	// Use transaction for atomic update
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Get current employee data
		var employee model.Employee
		err := r.collection.FindOne(sessCtx, bson.M{"_id": objID}).Decode(&employee)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrEmployeeNotFound
			}
			return err
		}

		// Calculate new balance
		newBalance := employee.OvertimeBalance + hours
		if newBalance < -200 || newBalance > 200 {
			return fmt.Errorf("%w: resulting balance would be %.2f hours", ErrInvalidOvertimeData, newBalance)
		}

		// Update balance
		update := bson.M{
			"$set": bson.M{
				"overtimeBalance":    newBalance,
				"lastTimeCalculated": time.Now(),
				"updatedAt":          time.Now(),
			},
		}

		_, err = r.collection.UpdateOne(sessCtx, bson.M{"_id": objID}, update)
		return err
	})
}

// UpdateVacationDays aktualisiert die verbleibenden Urlaubstage
func (r *EmployeeRepository) UpdateVacationDays(employeeID string, days int, reason string) error {
	// Validate input
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return err
	}

	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Get current employee data
		var employee model.Employee
		err := r.collection.FindOne(sessCtx, bson.M{"_id": objID}).Decode(&employee)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ErrEmployeeNotFound
			}
			return err
		}

		// Calculate new vacation days
		newDays := employee.RemainingVacation + days
		if newDays < 0 {
			return fmt.Errorf("%w: insufficient vacation days (available: %d)", ErrInvalidVacationDays, employee.RemainingVacation)
		}
		if newDays > employee.VacationDays*2 {
			return fmt.Errorf("%w: cannot exceed twice the annual allowance", ErrInvalidVacationDays)
		}

		// Update vacation days
		update := bson.M{
			"$set": bson.M{
				"remainingVacation": newDays,
				"updatedAt":         time.Now(),
			},
		}

		_, err = r.collection.UpdateOne(sessCtx, bson.M{"_id": objID}, update)
		return err
	})
}

// Delete performs a soft delete on an employee
func (r *EmployeeRepository) Delete(id string) error {
	objID, err := r.ValidateObjectID(id)
	if err != nil {
		return err
	}

	// Use transaction to deactivate employee and associated user
	return r.Transaction(func(sessCtx mongo.SessionContext) error {
		// Update employee status to inactive
		update := bson.M{
			"$set": bson.M{
				"status":    model.EmployeeStatusInactive,
				"updatedAt": time.Now(),
			},
		}

		result, err := r.collection.UpdateOne(sessCtx, bson.M{"_id": objID}, update)
		if err != nil {
			return err
		}

		if result.MatchedCount == 0 {
			return ErrEmployeeNotFound
		}

		// Also deactivate associated user
		userCollection := db.GetCollection("users")
		_, err = userCollection.UpdateOne(
			sessCtx,
			bson.M{"employeeId": objID},
			bson.M{
				"$set": bson.M{
					"status":    model.StatusInactive,
					"updatedAt": time.Now(),
				},
			},
		)

		return err
	})
}

// GetEmployeesByDepartment findet alle Mitarbeiter einer Abteilung
func (r *EmployeeRepository) GetEmployeesByDepartment(department string) ([]*model.Employee, error) {
	var employees []*model.Employee

	filter := bson.M{
		"department": department,
		"status":     model.EmployeeStatusActive,
	}

	err := r.BaseRepository.FindAll(filter, &employees, options.Find().SetSort(bson.M{"lastName": 1}))
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// EmployeeIDExists prüft, ob eine Employee ID bereits existiert
func (r *EmployeeRepository) EmployeeIDExists(employeeID string) (bool, error) {
	employeeID = strings.TrimSpace(employeeID)
	if employeeID == "" {
		return false, fmt.Errorf("%w: employee ID cannot be empty", ErrInvalidEmployeeData)
	}

	return r.Exists(bson.M{"employeeId": employeeID})
}

// GetEmployeesWithLowVacationDays findet Mitarbeiter mit wenigen verbleibenden Urlaubstagen
func (r *EmployeeRepository) GetEmployeesWithLowVacationDays(threshold int) ([]*model.Employee, error) {
	var employees []*model.Employee

	filter := bson.M{
		"status":            model.EmployeeStatusActive,
		"remainingVacation": bson.M{"$lte": threshold},
	}

	err := r.BaseRepository.FindAll(filter, &employees, options.Find().SetSort(bson.M{"remainingVacation": 1}))
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// GetEmployeesWithHighOvertime findet Mitarbeiter mit hohen Überstunden
func (r *EmployeeRepository) GetEmployeesWithHighOvertime(threshold float64) ([]*model.Employee, error) {
	var employees []*model.Employee

	filter := bson.M{
		"status":          model.EmployeeStatusActive,
		"overtimeBalance": bson.M{"$gte": threshold},
	}

	err := r.BaseRepository.FindAll(filter, &employees, options.Find().SetSort(bson.M{"overtimeBalance": -1}))
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// CreateIndexes erstellt erforderliche Indizes
func (r *EmployeeRepository) CreateIndexes() error {
	// Unique index on employee ID
	if err := r.CreateIndex(bson.M{"employeeId": 1}, true); err != nil {
		return fmt.Errorf("failed to create employeeId index: %w", err)
	}

	// Index on status for queries
	if err := r.CreateIndex(bson.M{"status": 1}, false); err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Compound index for department queries
	if err := r.CreateIndex(bson.M{"department": 1, "status": 1}, false); err != nil {
		return fmt.Errorf("failed to create department index: %w", err)
	}

	// Index on email for lookups
	if err := r.CreateIndex(bson.M{"email": 1}, false); err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}

	// Index for sorting by name
	if err := r.CreateIndex(bson.M{"lastName": 1, "firstName": 1}, false); err != nil {
		return fmt.Errorf("failed to create name index: %w", err)
	}

	return nil
}

// FindManagers findet alle Mitarbeiter mit Führungsposition
func (r *EmployeeRepository) FindManagers() ([]*model.Employee, error) {
	var employees []*model.Employee

	// Filter für Mitarbeiter mit Führungspositionen
	filter := bson.M{
		"status": model.EmployeeStatusActive,
		"$or": []bson.M{
			{"position": bson.M{"$regex": "Manager", "$options": "i"}},
			{"position": bson.M{"$regex": "Lead", "$options": "i"}},
			{"position": bson.M{"$regex": "Head", "$options": "i"}},
			{"position": bson.M{"$regex": "Director", "$options": "i"}},
			{"position": bson.M{"$regex": "Chief", "$options": "i"}},
		},
	}

	findOptions := options.Find().SetSort(bson.M{"lastName": 1})

	err := r.BaseRepository.FindAll(filter, &employees, findOptions)
	if err != nil {
		return nil, err
	}

	return employees, nil
}

// GetDepartmentCounts gibt die Anzahl der Mitarbeiter pro Abteilung zurück
func (r *EmployeeRepository) GetDepartmentCounts() (map[string]int, error) {
	ctx, cancel := r.GetContext()
	defer cancel()

	// Aggregation pipeline
	pipeline := []bson.M{
		{
			"$match": bson.M{"status": model.EmployeeStatusActive},
		},
		{
			"$group": bson.M{
				"_id":   "$department",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get department counts: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// GetActiveEmployeesCount zählt alle aktiven Mitarbeiter
func (r *EmployeeRepository) GetActiveEmployeesCount() (int64, error) {
	return r.Count(bson.M{"status": model.EmployeeStatusActive})
}

// FindByEmail findet einen Mitarbeiter anhand seiner E-Mail
func (r *EmployeeRepository) FindByEmail(email string) (*model.Employee, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return nil, fmt.Errorf("%w: email cannot be empty", ErrInvalidEmployeeData)
	}

	var employee model.Employee
	err := r.FindOne(bson.M{"email": email, "status": model.EmployeeStatusActive}, &employee)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	return &employee, nil
}

// UpdateTimebutlerUserID aktualisiert die Timebutler User ID eines Mitarbeiters
func (r *EmployeeRepository) UpdateTimebutlerUserID(employeeID string, timebutlerUserID string) error {
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"timebutlerUserId": timebutlerUserID,
			"updatedAt":        time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}

// UpdateErfasst123ID aktualisiert die 123erfasst ID eines Mitarbeiters
func (r *EmployeeRepository) UpdateErfasst123ID(employeeID string, erfasst123ID string) error {
	objID, err := r.ValidateObjectID(employeeID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"erfasst123Id": erfasst123ID,
			"updatedAt":    time.Now(),
		},
	}

	result, err := r.UpdateOne(bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}
