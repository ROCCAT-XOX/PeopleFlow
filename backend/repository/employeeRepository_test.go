package repository

import (
	"context"
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
	empRepo       *EmployeeRepository
	empCollection *mongo.Collection
	testClient    *mongo.Client
)

func setupEmployeeTest(t *testing.T) {
	// Initialize logger for testing
	err := utils.InitLogger(utils.LoggerConfig{
		Level:  utils.LogLevelDebug,
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to test database
	testClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
	}

	// Use test database
	testDB := testClient.Database("peopleflow_employee_test")
	empCollection = testDB.Collection("employees")
	
	// Create employee repository with test collection
	empRepo = &EmployeeRepository{
		BaseRepository: NewBaseRepository(empCollection),
		collection:     empCollection,
		userRepo:       NewUserRepository(), // This might need mocking in real tests
	}

	// Clean up any existing test data
	_, err = empCollection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		t.Fatalf("Failed to clean test collection: %v", err)
	}
}

func teardownEmployeeTest(t *testing.T) {
	if empCollection != nil {
		// Clean up test data
		_, _ = empCollection.DeleteMany(context.Background(), bson.M{})
		_ = empCollection.Drop(context.Background())
	}
	if testClient != nil {
		_ = testClient.Disconnect(context.Background())
	}
}

func createTestEmployee() *model.Employee {
	return &model.Employee{
		FirstName:           "John",
		LastName:            "Doe",
		Email:               "john.doe@example.com",
		Phone:               "+1234567890",
		InternalPhone:       "123",
		InternalExtension:   "456",
		Address:             "123 Main St, City, State",
		DateOfBirth:         time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		HireDate:            time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Position:            "Software Engineer",
		Department:          model.DepartmentIT,
		Status:              model.EmployeeStatusActive,
		WorkingHoursPerWeek: 40,
		WorkingDaysPerWeek:  5,
		WorkTimeModel:       model.WorkTimeModelFlexTime,
		VacationDays:        30,
		RemainingVacation:   30,
		OvertimeBalance:     0,
		Salary:              75000,
		Documents:           []model.Document{},
		Absences:            []model.Absence{},
		TimeEntries:         []model.TimeEntry{},
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

func TestEmployeeRepository_ValidateEmployee(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	tests := []struct {
		name        string
		employee    *model.Employee
		isUpdate    bool
		shouldError bool
		errorType   error
	}{
		{
			name:        "valid new employee",
			employee:    createTestEmployee(),
			isUpdate:    false,
			shouldError: false,
		},
		{
			name: "empty first name",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.FirstName = ""
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "empty last name",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.LastName = ""
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "invalid email",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.Email = "invalid-email"
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "negative working hours",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.WorkingHoursPerWeek = -5
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "excessive working hours",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.WorkingHoursPerWeek = 70
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "negative vacation days",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.VacationDays = -5
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "excessive vacation days",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.VacationDays = 400
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "extreme negative overtime",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.OvertimeBalance = -250
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "extreme positive overtime",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.OvertimeBalance = 250
				return emp
			}(),
			isUpdate:    false,
			shouldError: true,
		},
		{
			name: "valid update without required fields",
			employee: func() *model.Employee {
				emp := createTestEmployee()
				emp.FirstName = ""
				emp.LastName = ""
				return emp
			}(),
			isUpdate:    true,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := empRepo.ValidateEmployee(tt.employee, tt.isUpdate)

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

func TestEmployeeRepository_Create(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	t.Run("successful creation", func(t *testing.T) {
		emp := createTestEmployee()
		err := empRepo.Create(emp)
		if err != nil {
			t.Fatalf("Failed to create employee: %v", err)
		}

		if emp.ID.IsZero() {
			t.Error("Employee ID should be set after creation")
		}

		if emp.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set")
		}

		if emp.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should be set")
		}

		if emp.Status != model.EmployeeStatusActive {
			t.Errorf("Expected status to be active, got %s", emp.Status)
		}

		// Verify employee was actually saved
		var savedEmp model.Employee
		err = empRepo.FindByID(emp.ID.Hex(), &savedEmp)
		if err != nil {
			t.Fatalf("Failed to find created employee: %v", err)
		}

		if savedEmp.FirstName != emp.FirstName {
			t.Errorf("Expected first name %s, got %s", emp.FirstName, savedEmp.FirstName)
		}
	})

	t.Run("duplicate employee ID", func(t *testing.T) {
		emp1 := createTestEmployee()
		emp1.EmployeeID = "EMP001"
		err := empRepo.Create(emp1)
		if err != nil {
			t.Fatalf("Failed to create first employee: %v", err)
		}

		emp2 := createTestEmployee()
		emp2.EmployeeID = "EMP001" // Same ID
		emp2.Email = "different@example.com"
		err = empRepo.Create(emp2)
		if err == nil {
			t.Error("Expected error for duplicate employee ID")
		}
	})

	t.Run("invalid employee data", func(t *testing.T) {
		emp := createTestEmployee()
		emp.FirstName = "" // Invalid
		err := empRepo.Create(emp)
		if err == nil {
			t.Error("Expected validation error for invalid employee data")
		}
	})
}

func TestEmployeeRepository_FindByID(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employee
	emp := createTestEmployee()
	err := empRepo.Create(emp)
	if err != nil {
		t.Fatalf("Failed to create test employee: %v", err)
	}

	t.Run("find existing employee", func(t *testing.T) {
		foundEmp, err := empRepo.FindByID(emp.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find employee: %v", err)
		}

		if foundEmp.FirstName != emp.FirstName {
			t.Errorf("Expected first name %s, got %s", emp.FirstName, foundEmp.FirstName)
		}
		if foundEmp.Email != emp.Email {
			t.Errorf("Expected email %s, got %s", emp.Email, foundEmp.Email)
		}
	})

	t.Run("find non-existent employee", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		_, err := empRepo.FindByID(nonExistentID)
		if err == nil {
			t.Error("Expected error for non-existent employee")
		}
	})

	t.Run("invalid ID format", func(t *testing.T) {
		_, err := empRepo.FindByID("invalid-id")
		if err == nil {
			t.Error("Expected error for invalid ID format")
		}
	})
}

func TestEmployeeRepository_FindByEmail(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employee
	emp := createTestEmployee()
	err := empRepo.Create(emp)
	if err != nil {
		t.Fatalf("Failed to create test employee: %v", err)
	}

	t.Run("find existing employee by email", func(t *testing.T) {
		foundEmp, err := empRepo.FindByEmail(emp.Email)
		if err != nil {
			t.Fatalf("Failed to find employee by email: %v", err)
		}

		if foundEmp.Email != emp.Email {
			t.Errorf("Expected email %s, got %s", emp.Email, foundEmp.Email)
		}
	})

	t.Run("find non-existent employee by email", func(t *testing.T) {
		_, err := empRepo.FindByEmail("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error for non-existent employee email")
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := empRepo.FindByEmail("")
		if err == nil {
			t.Error("Expected error for empty email")
		}
	})
}

func TestEmployeeRepository_Update(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employee
	emp := createTestEmployee()
	err := empRepo.Create(emp)
	if err != nil {
		t.Fatalf("Failed to create test employee: %v", err)
	}

	t.Run("successful update", func(t *testing.T) {
		emp.FirstName = "Jane"
		emp.LastName = "Smith"
		emp.Position = "Senior Software Engineer"

		err := empRepo.Update(emp)
		if err != nil {
			t.Fatalf("Failed to update employee: %v", err)
		}

		// Verify update
		updatedEmp, err := empRepo.FindByID(emp.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find updated employee: %v", err)
		}

		if updatedEmp.FirstName != "Jane" {
			t.Errorf("Expected first name Jane, got %s", updatedEmp.FirstName)
		}
		if updatedEmp.LastName != "Smith" {
			t.Errorf("Expected last name Smith, got %s", updatedEmp.LastName)
		}
		if updatedEmp.Position != "Senior Software Engineer" {
			t.Errorf("Expected position Senior Software Engineer, got %s", updatedEmp.Position)
		}
	})

	t.Run("invalid update data", func(t *testing.T) {
		emp.WorkingHoursPerWeek = -10 // Invalid
		err := empRepo.Update(emp)
		if err == nil {
			t.Error("Expected validation error for invalid update data")
		}
	})
}

func TestEmployeeRepository_UpdateOvertimeBalance(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employee
	emp := createTestEmployee()
	err := empRepo.Create(emp)
	if err != nil {
		t.Fatalf("Failed to create test employee: %v", err)
	}

	t.Run("successful overtime update", func(t *testing.T) {
		err := empRepo.UpdateOvertimeBalance(emp.ID.Hex(), 5.5, "Manual adjustment")
		if err != nil {
			t.Fatalf("Failed to update overtime balance: %v", err)
		}

		// Verify update
		updatedEmp, err := empRepo.FindByID(emp.ID.Hex())
		if err != nil {
			t.Fatalf("Failed to find updated employee: %v", err)
		}

		if updatedEmp.OvertimeBalance != 5.5 {
			t.Errorf("Expected overtime balance 5.5, got %f", updatedEmp.OvertimeBalance)
		}
	})

	t.Run("extreme overtime balance", func(t *testing.T) {
		err := empRepo.UpdateOvertimeBalance(emp.ID.Hex(), 250, "Invalid adjustment")
		if err == nil {
			t.Error("Expected error for extreme overtime balance")
		}
	})

	t.Run("non-existent employee", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		err := empRepo.UpdateOvertimeBalance(nonExistentID, 5, "Manual adjustment")
		if err == nil {
			t.Error("Expected error for non-existent employee")
		}
	})
}

func TestEmployeeRepository_GetEmployeesByDepartment(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employees in different departments
	empIT := createTestEmployee()
	empIT.Department = model.DepartmentIT
	empIT.Email = "it@example.com"
	err := empRepo.Create(empIT)
	if err != nil {
		t.Fatalf("Failed to create IT employee: %v", err)
	}

	empHR := createTestEmployee()
	empHR.Department = model.DepartmentHR
	empHR.Email = "hr@example.com"
	err = empRepo.Create(empHR)
	if err != nil {
		t.Fatalf("Failed to create HR employee: %v", err)
	}

	empIT2 := createTestEmployee()
	empIT2.Department = model.DepartmentIT
	empIT2.Email = "it2@example.com"
	err = empRepo.Create(empIT2)
	if err != nil {
		t.Fatalf("Failed to create second IT employee: %v", err)
	}

	t.Run("find employees by department", func(t *testing.T) {
		itEmployees, err := empRepo.GetEmployeesByDepartment(string(model.DepartmentIT))
		if err != nil {
			t.Fatalf("Failed to find IT employees: %v", err)
		}

		if len(itEmployees) != 2 {
			t.Errorf("Expected 2 IT employees, got %d", len(itEmployees))
		}

		hrEmployees, err := empRepo.GetEmployeesByDepartment(string(model.DepartmentHR))
		if err != nil {
			t.Fatalf("Failed to find HR employees: %v", err)
		}

		if len(hrEmployees) != 1 {
			t.Errorf("Expected 1 HR employee, got %d", len(hrEmployees))
		}
	})

	t.Run("find employees in non-existent department", func(t *testing.T) {
		employees, err := empRepo.GetEmployeesByDepartment("NonExistent")
		if err != nil {
			t.Fatalf("Failed to query non-existent department: %v", err)
		}

		if len(employees) != 0 {
			t.Errorf("Expected 0 employees in non-existent department, got %d", len(employees))
		}
	})
}

func TestEmployeeRepository_GetActiveEmployeesCount(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employees with different statuses
	activeEmp1 := createTestEmployee()
	activeEmp1.Email = "active1@example.com"
	activeEmp1.Status = model.EmployeeStatusActive
	err := empRepo.Create(activeEmp1)
	if err != nil {
		t.Fatalf("Failed to create active employee 1: %v", err)
	}

	activeEmp2 := createTestEmployee()
	activeEmp2.Email = "active2@example.com"
	activeEmp2.Status = model.EmployeeStatusActive
	err = empRepo.Create(activeEmp2)
	if err != nil {
		t.Fatalf("Failed to create active employee 2: %v", err)
	}

	inactiveEmp := createTestEmployee()
	inactiveEmp.Email = "inactive@example.com"
	inactiveEmp.Status = model.EmployeeStatusInactive
	err = empRepo.Create(inactiveEmp)
	if err != nil {
		t.Fatalf("Failed to create inactive employee: %v", err)
	}

	count, err := empRepo.GetActiveEmployeesCount()
	if err != nil {
		t.Fatalf("Failed to get active employees count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 active employees, got %d", count)
	}
}

func TestEmployeeRepository_EmployeeIDExists(t *testing.T) {
	setupEmployeeTest(t)
	defer teardownEmployeeTest(t)

	// Create test employee
	emp := createTestEmployee()
	emp.EmployeeID = "EMP001"
	err := empRepo.Create(emp)
	if err != nil {
		t.Fatalf("Failed to create test employee: %v", err)
	}

	t.Run("existing employee ID", func(t *testing.T) {
		exists, err := empRepo.EmployeeIDExists("EMP001")
		if err != nil {
			t.Fatalf("Failed to check employee ID existence: %v", err)
		}

		if !exists {
			t.Error("Expected employee ID to exist")
		}
	})

	t.Run("non-existing employee ID", func(t *testing.T) {
		exists, err := empRepo.EmployeeIDExists("EMP999")
		if err != nil {
			t.Fatalf("Failed to check employee ID existence: %v", err)
		}

		if exists {
			t.Error("Expected employee ID not to exist")
		}
	})

	t.Run("empty employee ID", func(t *testing.T) {
		_, err := empRepo.EmployeeIDExists("")
		if err == nil {
			t.Error("Expected error for empty employee ID")
		}
	})
}

// Benchmark tests
func BenchmarkEmployeeRepository_Create(b *testing.B) {
	setupEmployeeTest(&testing.T{})
	defer teardownEmployeeTest(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emp := createTestEmployee()
		emp.Email = fmt.Sprintf("bench%d@example.com", i)
		emp.EmployeeID = fmt.Sprintf("EMP%d", i)
		_ = empRepo.Create(emp)
	}
}

func BenchmarkEmployeeRepository_FindByID(b *testing.B) {
	setupEmployeeTest(&testing.T{})
	defer teardownEmployeeTest(&testing.T{})

	// Create test employee
	emp := createTestEmployee()
	_ = empRepo.Create(emp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = empRepo.FindByID(emp.ID.Hex())
	}
}

func BenchmarkEmployeeRepository_FindByEmail(b *testing.B) {
	setupEmployeeTest(&testing.T{})
	defer teardownEmployeeTest(&testing.T{})

	// Create test employee
	emp := createTestEmployee()
	_ = empRepo.Create(emp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = empRepo.FindByEmail(emp.Email)
	}
}