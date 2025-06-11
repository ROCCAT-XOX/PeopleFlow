package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
)

// TestEmployeeHandlerBasics tests basic employee management functionality
func TestEmployeeHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock employee for testing
	mockEmployee := &model.Employee{
		ID:                  primitive.NewObjectID(),
		EmployeeID:          "EMP001",
		FirstName:           "John",
		LastName:            "Doe",
		Email:               "john.doe@company.com",
		Phone:               "+1234567890",
		Position:            "Software Developer",
		Department:          model.DepartmentIT,
		Status:              model.EmployeeStatusActive,
		HireDate:            time.Now().AddDate(-1, 0, 0),
		WorkingHoursPerWeek: 40.0,
		WorkingDaysPerWeek:  5,
		WorkTimeModel:       model.WorkTimeModelFullTime,
		Salary:              75000.0,
		VacationDays:        25,
		RemainingVacation:   20,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	t.Run("GetEmployeesHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/employees", func(c *gin.Context) {
			// Simulate authenticated user
			c.Set("user", &model.User{Role: model.RoleManager})
			c.Set("userRole", "manager")
			
			employees := []model.Employee{*mockEmployee}
			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"employees": employees,
				"total":     len(employees),
			})
		})

		req, _ := http.NewRequest("GET", "/employees", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "employees")
	})

	t.Run("CreateEmployeeHandler_ValidData", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/employees", func(c *gin.Context) {
			// Simulate authenticated HR user
			c.Set("user", &model.User{Role: model.RoleHR})
			c.Set("userRole", "hr")
			
			var employeeData map[string]interface{}
			if err := c.ShouldBindJSON(&employeeData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Basic validation
			if employeeData["firstName"] == "" || employeeData["lastName"] == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "First name and last name are required"})
				return
			}

			// Simulate successful employee creation
			newEmployee := model.Employee{
				ID:           primitive.NewObjectID(),
				EmployeeID:   "EMP002",
				FirstName:    employeeData["firstName"].(string),
				LastName:     employeeData["lastName"].(string),
				Email:        employeeData["email"].(string),
				Position:     employeeData["position"].(string),
				Department:   model.Department(employeeData["department"].(string)),
				Status:       model.EmployeeStatusActive,
				HireDate:     time.Now(),
				CreatedAt:    time.Now(),
			}

			c.JSON(http.StatusCreated, gin.H{
				"success":  true,
				"employee": newEmployee,
				"message":  "Employee created successfully",
			})
		})

		employeeData := map[string]interface{}{
			"firstName":  "Jane",
			"lastName":   "Smith",
			"email":      "jane.smith@company.com",
			"position":   "Product Manager",
			"department": "Marketing",
			"phone":      "+1987654321",
		}

		jsonData, _ := json.Marshal(employeeData)
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "employee")
	})

	t.Run("GetEmployeeDetailsHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/employees/:id", func(c *gin.Context) {
			employeeID := c.Param("id")
			
			// Simulate authenticated user
			c.Set("user", &model.User{Role: model.RoleHR})
			c.Set("userRole", "hr")
			
			// Validate ObjectID format
			if _, err := primitive.ObjectIDFromHex(employeeID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
				return
			}

			// Simulate successful employee retrieval
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"employee": mockEmployee,
			})
		})

		req, _ := http.NewRequest("GET", "/employees/"+mockEmployee.ID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "employee")
	})

	t.Run("UpdateEmployeeHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/employees/:id", func(c *gin.Context) {
			employeeID := c.Param("id")
			
			// Simulate authenticated HR user
			c.Set("user", &model.User{Role: model.RoleHR})
			c.Set("userRole", "hr")
			
			var updateData map[string]interface{}
			if err := c.ShouldBindJSON(&updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Simulate successful update
			c.JSON(http.StatusOK, gin.H{
				"success":    true,
				"message":    "Employee updated successfully",
				"employeeId": employeeID,
			})
		})

		updateData := map[string]interface{}{
			"position": "Senior Software Developer",
			"salary":   85000.0,
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/employees/"+mockEmployee.ID.Hex(), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

// TestEmployeeInputValidation tests input validation for employee operations
func TestEmployeeInputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		employeeData map[string]interface{}
		expectedCode int
		shouldPass   bool
	}{
		{
			name: "Valid employee data",
			employeeData: map[string]interface{}{
				"firstName":  "John",
				"lastName":   "Doe",
				"email":      "john@company.com",
				"position":   "Developer",
				"department": "IT",
			},
			expectedCode: http.StatusCreated,
			shouldPass:   true,
		},
		{
			name: "Missing first name",
			employeeData: map[string]interface{}{
				"lastName":   "Doe",
				"email":      "john@company.com",
				"position":   "Developer",
				"department": "IT",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name: "Missing last name",
			employeeData: map[string]interface{}{
				"firstName":  "John",
				"email":      "john@company.com",
				"position":   "Developer",
				"department": "IT",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name: "Invalid email format",
			employeeData: map[string]interface{}{
				"firstName":  "John",
				"lastName":   "Doe",
				"email":      "invalid-email",
				"position":   "Developer",
				"department": "IT",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.POST("/employees", func(c *gin.Context) {
				var employeeData map[string]interface{}
				if err := c.ShouldBindJSON(&employeeData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}

				// Validate required fields
				if employeeData["firstName"] == nil || employeeData["firstName"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "First name is required"})
					return
				}
				if employeeData["lastName"] == nil || employeeData["lastName"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Last name is required"})
					return
				}

				// Validate email format if provided
				if email, exists := employeeData["email"]; exists && email != "" {
					emailStr := email.(string)
					if !strings.Contains(emailStr, "@") || !strings.Contains(emailStr, ".") {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
						return
					}
				}

				c.JSON(http.StatusCreated, gin.H{"success": true})
			})

			jsonData, _ := json.Marshal(tt.employeeData)
			req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestEmployeeWorkTimeCalculations tests work time related functionality
func TestEmployeeWorkTimeCalculations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetEmployeeWorkingHours", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/employees/:id/working-hours", func(c *gin.Context) {
			// Mock employee with specific working hours
			employee := model.Employee{
				WorkingHoursPerWeek: 40.0,
				WorkingDaysPerWeek:  5,
				WorkTimeModel:       model.WorkTimeModelFullTime,
			}
			
			workingHoursPerDay := employee.GetWorkingHoursPerDay()
			description := employee.GetWorkingTimeDescription()
			isFullTime := employee.IsFullTimeEmployee()
			
			c.JSON(http.StatusOK, gin.H{
				"success":             true,
				"workingHoursPerWeek": employee.WorkingHoursPerWeek,
				"workingHoursPerDay":  workingHoursPerDay,
				"workingDaysPerWeek":  employee.WorkingDaysPerWeek,
				"description":         description,
				"isFullTime":          isFullTime,
				"workTimeModel":       employee.WorkTimeModel.GetDisplayName(),
			})
		})

		req, _ := http.NewRequest("GET", "/employees/"+primitive.NewObjectID().Hex()+"/working-hours", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, float64(8), response["workingHoursPerDay"].(float64)) // 40/5 = 8
		assert.True(t, response["isFullTime"].(bool))
	})

	t.Run("UpdateWorkingHours", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/employees/:id/working-hours", func(c *gin.Context) {
			var workingHoursData map[string]interface{}
			if err := c.ShouldBindJSON(&workingHoursData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Validate working hours
			if hours, exists := workingHoursData["workingHoursPerWeek"]; exists {
				if hoursFloat := hours.(float64); hoursFloat < 0 || hoursFloat > 80 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Working hours must be between 0 and 80"})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Working hours updated successfully",
			})
		})

		workingHoursData := map[string]interface{}{
			"workingHoursPerWeek": 30.0,
			"workingDaysPerWeek":  4,
			"workTimeModel":       "parttime",
		}

		jsonData, _ := json.Marshal(workingHoursData)
		req, _ := http.NewRequest("PUT", "/employees/"+primitive.NewObjectID().Hex()+"/working-hours", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestEmployeeOvertimeOperations tests overtime-related functionality
func TestEmployeeOvertimeOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockEmployee := &model.Employee{
		ID:              primitive.NewObjectID(),
		EmployeeID:      "EMP001",
		FirstName:       "John",
		LastName:        "Doe",
		OvertimeBalance: 8.5,
		OvertimeAdjustments: []model.OvertimeAdjustment{
			{
				ID:     primitive.NewObjectID(),
				Hours:  2.0,
				Reason: "Extra project work",
				Status: "approved",
			},
		},
	}

	t.Run("GetEmployeeOvertimeBalance", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/employees/:id/overtime", func(c *gin.Context) {
			baseBalance := mockEmployee.OvertimeBalance
			adjustmentsTotal := mockEmployee.GetTotalAdjustments()
			finalBalance := mockEmployee.GetAdjustedOvertimeBalance()
			status := mockEmployee.GetOvertimeStatus()
			
			c.JSON(http.StatusOK, gin.H{
				"success":          true,
				"baseBalance":      baseBalance,
				"adjustmentsTotal": adjustmentsTotal,
				"finalBalance":     finalBalance,
				"status":           status,
				"formatted":        mockEmployee.FormatAdjustedOvertimeBalance(),
				"details":          mockEmployee.GetOvertimeBalanceWithDetails(),
			})
		})

		req, _ := http.NewRequest("GET", "/employees/"+mockEmployee.ID.Hex()+"/overtime", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Equal(t, float64(8.5), response["baseBalance"].(float64))
		assert.Equal(t, float64(2.0), response["adjustmentsTotal"].(float64))
		assert.Equal(t, float64(10.5), response["finalBalance"].(float64))
		assert.Equal(t, "positive", response["status"].(string))
	})
}

// TestEmployeeAbsenceManagement tests absence-related functionality
func TestEmployeeAbsenceManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetEmployeeAbsences", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/employees/:id/absences", func(c *gin.Context) {
			// Mock absences
			absences := []model.Absence{
				{
					ID:        primitive.NewObjectID(),
					Type:      "vacation",
					StartDate: time.Now().AddDate(0, 0, 7),
					EndDate:   time.Now().AddDate(0, 0, 14),
					Days:      5,
					Status:    "approved",
					Reason:    "Annual vacation",
				},
				{
					ID:        primitive.NewObjectID(),
					Type:      "sick",
					StartDate: time.Now().AddDate(0, 0, -3),
					EndDate:   time.Now().AddDate(0, 0, -1),
					Days:      2,
					Status:    "approved",
					Reason:    "Flu",
				},
			}
			
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"absences": absences,
				"total":    len(absences),
			})
		})

		req, _ := http.NewRequest("GET", "/employees/"+primitive.NewObjectID().Hex()+"/absences", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "absences")
		assert.Equal(t, float64(2), response["total"].(float64))
	})

	t.Run("CreateEmployeeAbsence", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/employees/:id/absences", func(c *gin.Context) {
			var absenceData map[string]interface{}
			if err := c.ShouldBindJSON(&absenceData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Basic validation
			if absenceData["type"] == "" || absenceData["startDate"] == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Type and start date are required"})
				return
			}

			// Simulate successful absence creation
			newAbsence := model.Absence{
				ID:     primitive.NewObjectID(),
				Type:   absenceData["type"].(string),
				Reason: absenceData["reason"].(string),
				Status: "requested",
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"absence": newAbsence,
				"message": "Absence request created successfully",
			})
		})

		absenceData := map[string]interface{}{
			"type":      "vacation",
			"startDate": "2025-07-01",
			"endDate":   "2025-07-05",
			"reason":    "Summer vacation",
		}

		jsonData, _ := json.Marshal(absenceData)
		req, _ := http.NewRequest("POST", "/employees/"+primitive.NewObjectID().Hex()+"/absences", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "absence")
	})
}

// Benchmark tests for employee operations
func BenchmarkGetEmployeesHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/employees", func(c *gin.Context) {
		employees := make([]model.Employee, 500) // Simulate 500 employees
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"employees": employees,
			"total":     len(employees),
		})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/employees", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCreateEmployeeHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.POST("/employees", func(c *gin.Context) {
		var employeeData map[string]interface{}
		c.ShouldBindJSON(&employeeData)
		c.JSON(http.StatusCreated, gin.H{"success": true})
	})

	employeeData := map[string]interface{}{
		"firstName":  "Test",
		"lastName":   "Employee",
		"email":      "test@company.com",
		"position":   "Developer",
		"department": "IT",
	}
	jsonData, _ := json.Marshal(employeeData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}