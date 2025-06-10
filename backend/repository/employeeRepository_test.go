// backend/repository/employeeRepository_test.go
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

// EmployeeRepositoryTestSuite defines the test suite for EmployeeRepository
type EmployeeRepositoryTestSuite struct {
	suite.Suite
	repo           *EmployeeRepository
	collection     *mongo.Collection
	userCollection *mongo.Collection
	client         *mongo.Client
}

// SetupSuite runs once before all tests
func (suite *EmployeeRepositoryTestSuite) SetupSuite() {
	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(suite.T(), err)

	suite.client = client
	testDB := client.Database("peopleflow_test")
	suite.collection = testDB.Collection("employees")
	suite.userCollection = testDB.Collection("users")

	// Initialize test repositories
	db.SetTestCollection("employees", suite.collection)
	db.SetTestCollection("users", suite.userCollection)
	suite.repo = NewEmployeeRepository()
}

// SetupTest runs before each test
func (suite *EmployeeRepositoryTestSuite) SetupTest() {
	// Clear collections before each test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suite.collection.Drop(ctx)
	suite.userCollection.Drop(ctx)

	// Create indexes
	err := suite.repo.CreateIndexes()
	require.NoError(suite.T(), err)
}

// TearDownSuite runs once after all tests
func (suite *EmployeeRepositoryTestSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if suite.client != nil {
		suite.client.Disconnect(ctx)
	}
}

// Helper function to create a valid employee
func (suite *EmployeeRepositoryTestSuite) createValidEmployee() *model.Employee {
	return &model.Employee{
		FirstName:             "Max",
		LastName:              "Mustermann",
		Email:                 "max.mustermann@example.com",
		EmployeeNumber:        "EMP001",
		Department:            "IT",
		Position:              "Developer",
		ContractType:          model.ContractTypeFullTime,
		WeeklyHours:           40,
		VacationDaysPerYear:   30,
		VacationDaysRemaining: 30,
		OvertimeBalance:       0,
		StartDate:             time.Now().AddDate(-1, 0, 0),
	}
}

// Test Create Employee - Success
func (suite *EmployeeRepositoryTestSuite) TestCreateEmployee_Success() {
	employee := suite.createValidEmployee()

	err := suite.repo.Create(employee)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), employee.ID)
	assert.True(suite.T(), employee.Active)
	assert.NotZero(suite.T(), employee.CreatedAt)
	assert.NotZero(suite.T(), employee.UpdatedAt)

	// Verify user was created
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user model.User
	err = suite.userCollection.FindOne(ctx, bson.M{"email": employee.Email}).Decode(&user)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), employee.ID, user.EmployeeID)
	assert.Equal(suite.T(), model.RoleEmployee, user.Role)
}

// Test Create Employee - Validation Errors
func (suite *EmployeeRepositoryTestSuite) TestCreateEmployee_ValidationErrors() {
	testCases := []struct {
		name        string
		employee    *model.Employee
		expectedErr error
	}{
		{
			name: "Empty first name",
			employee: &model.Employee{
				LastName:       "Mustermann",
				EmployeeNumber: "EMP001",
			},
			expectedErr: ErrInvalidEmployeeData,
		},
		{
			name: "Empty last name",
			employee: &model.Employee{
				FirstName:      "Max",
				EmployeeNumber: "EMP001",
			},
			expectedErr: ErrInvalidEmployeeData,
		},
		{
			name: "Empty employee number",
			employee: &model.Employee{
				FirstName: "Max",
				LastName:  "Mustermann",
			},
			expectedErr: ErrInvalidEmployeeData,
		},
		{
			name: "Invalid contract type",
			employee: &model.Employee{
				FirstName:      "Max",
				LastName:       "Mustermann",
				EmployeeNumber: "EMP001",
				ContractType:   "invalid",
			},
			expectedErr: ErrInvalidContractType,
		},
		{
			name: "Negative weekly hours",
			employee: &model.Employee{
				FirstName:      "Max",
				LastName:       "Mustermann",
				EmployeeNumber: "EMP001",
				WeeklyHours:    -10,
			},
			expectedErr: ErrInvalidWeeklyHours,
		},
		{
			name: "Too many weekly hours",
			employee: &model.Employee{
				FirstName:      "Max",
				LastName:       "Mustermann",
				EmployeeNumber: "EMP001",
				WeeklyHours:    70,
			},
			expectedErr: ErrInvalidWeeklyHours,
		},
		{
			name: "Negative vacation days",
			employee: &model.Employee{
				FirstName:           "Max",
				LastName:            "Mustermann",
				EmployeeNumber:      "EMP001",
				VacationDaysPerYear: -5,
			},
			expectedErr: ErrInvalidVacationDays,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.repo.Create(tc.employee)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr.Error())
		})
	}
}

// Test Create Employee - Duplicate Employee Number
func (suite *EmployeeRepositoryTestSuite) TestCreateEmployee_DuplicateEmployeeNumber() {
	// Create first employee
	employee1 := suite.createValidEmployee()
	err := suite.repo.Create(employee1)
	require.NoError(suite.T(), err)

	// Try to create second employee with same employee number
	employee2 := suite.createValidEmployee()
	employee2.Email = "different@example.com"
	err = suite.repo.Create(employee2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmployeeNumberTaken, err)
}

// Test FindByID
func (suite *EmployeeRepositoryTestSuite) TestFindByID() {
	// Create employee
	employee := suite.createValidEmployee()
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Find by ID
	found, err := suite.repo.FindByID(employee.ID.Hex())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), employee.EmployeeNumber, found.EmployeeNumber)
	assert.Equal(suite.T(), employee.Email, found.Email)

	// Test not found
	_, err = suite.repo.FindByID("507f1f77bcf86cd799439011")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmployeeNotFound, err)

	// Test invalid ID
	_, err = suite.repo.FindByID("invalid-id")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// Test FindByEmployeeNumber
func (suite *EmployeeRepositoryTestSuite) TestFindByEmployeeNumber() {
	// Create employee
	employee := suite.createValidEmployee()
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Find by employee number
	found, err := suite.repo.FindByEmployeeNumber("EMP001")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), employee.ID, found.ID)

	// Test not found
	_, err = suite.repo.FindByEmployeeNumber("EMP999")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmployeeNotFound, err)

	// Test empty employee number
	_, err = suite.repo.FindByEmployeeNumber("")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "empty")
}

// Test Update Employee
func (suite *EmployeeRepositoryTestSuite) TestUpdateEmployee() {
	// Create employee
	employee := suite.createValidEmployee()
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Update employee
	employee.Department = "Sales"
	employee.Position = "Manager"
	employee.WeeklyHours = 35

	err = suite.repo.Update(employee)
	assert.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.repo.FindByID(employee.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Sales", updated.Department)
	assert.Equal(suite.T(), "Manager", updated.Position)
	assert.Equal(suite.T(), float64(35), updated.WeeklyHours)
}

// Test Update Employee - Duplicate Employee Number
func (suite *EmployeeRepositoryTestSuite) TestUpdateEmployee_DuplicateEmployeeNumber() {
	// Create two employees
	employee1 := suite.createValidEmployee()
	employee1.EmployeeNumber = "EMP001"
	err := suite.repo.Create(employee1)
	require.NoError(suite.T(), err)

	employee2 := suite.createValidEmployee()
	employee2.EmployeeNumber = "EMP002"
	employee2.Email = "different@example.com"
	err = suite.repo.Create(employee2)
	require.NoError(suite.T(), err)

	// Try to update employee2 with employee1's number
	employee2.EmployeeNumber = "EMP001"
	err = suite.repo.Update(employee2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmployeeNumberTaken, err)
}

// Test UpdateOvertimeBalance
func (suite *EmployeeRepositoryTestSuite) TestUpdateOvertimeBalance() {
	// Create employee
	employee := suite.createValidEmployee()
	employee.OvertimeBalance = 10
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Add overtime
	err = suite.repo.UpdateOvertimeBalance(employee.ID.Hex(), 5.5, "Extra project work")
	assert.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.repo.FindByID(employee.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 15.5, updated.OvertimeBalance)

	// Subtract overtime
	err = suite.repo.UpdateOvertimeBalance(employee.ID.Hex(), -10, "Time off")
	assert.NoError(suite.T(), err)

	updated, err = suite.repo.FindByID(employee.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5.5, updated.OvertimeBalance)

	// Test exceeding limits
	err = suite.repo.UpdateOvertimeBalance(employee.ID.Hex(), 300, "Too much")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "overtime")
}

// Test UpdateVacationDays
func (suite *EmployeeRepositoryTestSuite) TestUpdateVacationDays() {
	// Create employee
	employee := suite.createValidEmployee()
	employee.VacationDaysPerYear = 30
	employee.VacationDaysRemaining = 25
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Use vacation days
	err = suite.repo.UpdateVacationDays(employee.ID.Hex(), -5, "Summer vacation")
	assert.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.repo.FindByID(employee.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 20, updated.VacationDaysRemaining)

	// Add vacation days (e.g., from previous year)
	err = suite.repo.UpdateVacationDays(employee.ID.Hex(), 10, "Carried over from last year")
	assert.NoError(suite.T(), err)

	updated, err = suite.repo.FindByID(employee.ID.Hex())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 30, updated.VacationDaysRemaining)

	// Test insufficient days
	err = suite.repo.UpdateVacationDays(employee.ID.Hex(), -35, "Too many days")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "insufficient")

	// Test exceeding limit
	err = suite.repo.UpdateVacationDays(employee.ID.Hex(), 100, "Too many days")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "exceed")
}

// Test Delete Employee (Soft Delete)
func (suite *EmployeeRepositoryTestSuite) TestDeleteEmployee() {
	// Create employee with user
	employee := suite.createValidEmployee()
	err := suite.repo.Create(employee)
	require.NoError(suite.T(), err)

	// Delete employee
	err = suite.repo.Delete(employee.ID.Hex())
	assert.NoError(suite.T(), err)

	// Verify employee is soft deleted
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var deleted model.Employee
	err = suite.collection.FindOne(ctx, bson.M{"_id": employee.ID}).Decode(&deleted)
	require.NoError(suite.T(), err)
	assert.False(suite.T(), deleted.Active)
	assert.NotNil(suite.T(), deleted.DeletedAt)

	// Verify associated user is deactivated
	var user model.User
	err = suite.userCollection.FindOne(ctx, bson.M{"employeeId": employee.ID}).Decode(&user)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), model.StatusInactive, user.Status)
}

// Test FindAll with Pagination and Sorting
func (suite *EmployeeRepositoryTestSuite) TestFindAll_PaginationAndSorting() {
	// Create multiple employees
	names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
	for i, name := range names {
		employee := suite.createValidEmployee()
		employee.FirstName = name
		employee.LastName = fmt.Sprintf("Employee%d", i)
		employee.EmployeeNumber = fmt.Sprintf("EMP%03d", i+1)
		employee.Email = fmt.Sprintf("%s@example.com", name)
		err := suite.repo.Create(employee)
		require.NoError(suite.T(), err)
	}

	// Test pagination
	employees, total, err := suite.repo.FindAll(0, 3, "firstName", 1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 3)
	assert.Equal(suite.T(), int64(5), total)
	assert.Equal(suite.T(), "Alice", employees[0].FirstName)

	// Test second page
	employees, total, err = suite.repo.FindAll(3, 3, "firstName", 1)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 2)
	assert.Equal(suite.T(), int64(5), total)
	assert.Equal(suite.T(), "David", employees[0].FirstName)

	// Test descending sort
	employees, _, err = suite.repo.FindAll(0, 5, "firstName", -1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Eve", employees[0].FirstName)
}

// Test GetEmployeesByDepartment
func (suite *EmployeeRepositoryTestSuite) TestGetEmployeesByDepartment() {
	// Create employees in different departments
	departments := []string{"IT", "IT", "Sales", "HR", "IT"}
	for i, dept := range departments {
		employee := suite.createValidEmployee()
		employee.Department = dept
		employee.EmployeeNumber = fmt.Sprintf("EMP%03d", i+1)
		employee.Email = fmt.Sprintf("emp%d@example.com", i+1)
		err := suite.repo.Create(employee)
		require.NoError(suite.T(), err)
	}

	// Get IT employees
	employees, err := suite.repo.GetEmployeesByDepartment("IT")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 3)

	for _, emp := range employees {
		assert.Equal(suite.T(), "IT", emp.Department)
	}
}

// Test GetEmployeesWithLowVacationDays
func (suite *EmployeeRepositoryTestSuite) TestGetEmployeesWithLowVacationDays() {
	// Create employees with different vacation days
	vacationDays := []int{5, 15, 3, 25, 8}
	for i, days := range vacationDays {
		employee := suite.createValidEmployee()
		employee.VacationDaysRemaining = days
		employee.EmployeeNumber = fmt.Sprintf("EMP%03d", i+1)
		employee.Email = fmt.Sprintf("emp%d@example.com", i+1)
		err := suite.repo.Create(employee)
		require.NoError(suite.T(), err)
	}

	// Get employees with 10 or fewer vacation days
	employees, err := suite.repo.GetEmployeesWithLowVacationDays(10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 3)

	// Verify sorting (ascending by vacation days)
	assert.Equal(suite.T(), 3, employees[0].VacationDaysRemaining)
	assert.Equal(suite.T(), 5, employees[1].VacationDaysRemaining)
	assert.Equal(suite.T(), 8, employees[2].VacationDaysRemaining)
}

// Test GetEmployeesWithHighOvertime
func (suite *EmployeeRepositoryTestSuite) TestGetEmployeesWithHighOvertime() {
	// Create employees with different overtime balances
	overtimes := []float64{5, 25, 15, 35, 10}
	for i, hours := range overtimes {
		employee := suite.createValidEmployee()
		employee.OvertimeBalance = hours
		employee.EmployeeNumber = fmt.Sprintf("EMP%03d", i+1)
		employee.Email = fmt.Sprintf("emp%d@example.com", i+1)
		err := suite.repo.Create(employee)
		require.NoError(suite.T(), err)
	}

	// Get employees with 20 or more overtime hours
	employees, err := suite.repo.GetEmployeesWithHighOvertime(20)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), employees, 2)

	// Verify sorting (descending by overtime)
	assert.Equal(suite.T(), float64(35), employees[0].OvertimeBalance)
	assert.Equal(suite.T(), float64(25), employees[1].OvertimeBalance)
}

// Test Transaction Rollback
func (suite *EmployeeRepositoryTestSuite) TestTransactionRollback() {
	// Create an employee
	employee := suite.createValidEmployee()
	employee.Email = "willcauseerror@" // Invalid email for user creation

	// This should fail during user creation and rollback
	err := suite.repo.Create(employee)
	assert.Error(suite.T(), err)

	// Verify employee was not created
	_, err = suite.repo.FindByEmployeeNumber(employee.EmployeeNumber)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrEmployeeNotFound, err)
}

// Run the test suite
func TestEmployeeRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(EmployeeRepositoryTestSuite))
}
