package model

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEmployeeWorkingHoursMethods(t *testing.T) {
	t.Run("GetWorkingHoursPerDay", func(t *testing.T) {
		tests := []struct {
			name               string
			workingHoursPerWeek float64
			workingDaysPerWeek  int
			expected           float64
		}{
			{
				name:               "full time 40h/5days",
				workingHoursPerWeek: 40.0,
				workingDaysPerWeek:  5,
				expected:           8.0,
			},
			{
				name:               "part time 20h/3days",
				workingHoursPerWeek: 20.0,
				workingDaysPerWeek:  3,
				expected:           6.666666666666667,
			},
			{
				name:               "zero working days",
				workingHoursPerWeek: 40.0,
				workingDaysPerWeek:  0,
				expected:           0.0,
			},
			{
				name:               "zero hours",
				workingHoursPerWeek: 0.0,
				workingDaysPerWeek:  5,
				expected:           0.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{
					WorkingHoursPerWeek: tt.workingHoursPerWeek,
					WorkingDaysPerWeek:  tt.workingDaysPerWeek,
				}
				result := employee.GetWorkingHoursPerDay()
				if result != tt.expected {
					t.Errorf("GetWorkingHoursPerDay() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	t.Run("IsFullTimeEmployee", func(t *testing.T) {
		tests := []struct {
			name               string
			workTimeModel      WorkTimeModel
			workingHoursPerWeek float64
			expected           bool
		}{
			{
				name:               "full time model",
				workTimeModel:      WorkTimeModelFullTime,
				workingHoursPerWeek: 30.0,
				expected:           true,
			},
			{
				name:               "part time model but 40h",
				workTimeModel:      WorkTimeModelPartTime,
				workingHoursPerWeek: 40.0,
				expected:           true,
			},
			{
				name:               "part time model and 20h",
				workTimeModel:      WorkTimeModelPartTime,
				workingHoursPerWeek: 20.0,
				expected:           false,
			},
			{
				name:               "35h threshold",
				workTimeModel:      WorkTimeModelFlexTime,
				workingHoursPerWeek: 35.0,
				expected:           true,
			},
			{
				name:               "below 35h threshold",
				workTimeModel:      WorkTimeModelFlexTime,
				workingHoursPerWeek: 34.0,
				expected:           false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{
					WorkTimeModel:       tt.workTimeModel,
					WorkingHoursPerWeek: tt.workingHoursPerWeek,
				}
				result := employee.IsFullTimeEmployee()
				if result != tt.expected {
					t.Errorf("IsFullTimeEmployee() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	t.Run("GetWorkingTimeDescription", func(t *testing.T) {
		tests := []struct {
			name               string
			employee           Employee
			expectedContains   []string
		}{
			{
				name: "full description",
				employee: Employee{
					WorkingHoursPerWeek: 40.0,
					WorkingDaysPerWeek:  5,
					WorkTimeModel:       WorkTimeModelFullTime,
				},
				expectedContains: []string{"40.0 Std/Woche", "8.0 Std/Tag", "Vollzeit"},
			},
			{
				name: "no working days",
				employee: Employee{
					WorkingHoursPerWeek: 30.0,
					WorkingDaysPerWeek:  0,
					WorkTimeModel:       WorkTimeModelPartTime,
				},
				expectedContains: []string{"30.0 Std/Woche", "Teilzeit"},
			},
			{
				name: "no hours set",
				employee: Employee{
					WorkingHoursPerWeek: 0.0,
				},
				expectedContains: []string{"Nicht festgelegt"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.employee.GetWorkingTimeDescription()
				for _, expected := range tt.expectedContains {
					if !containsString(result, expected) {
						t.Errorf("GetWorkingTimeDescription() = %q, should contain %q", result, expected)
					}
				}
			})
		}
	})

	t.Run("GetWeeklyTargetHours", func(t *testing.T) {
		tests := []struct {
			name               string
			workingHoursPerWeek float64
			expected           float64
		}{
			{
				name:               "set working hours",
				workingHoursPerWeek: 35.0,
				expected:           35.0,
			},
			{
				name:               "zero working hours (fallback)",
				workingHoursPerWeek: 0.0,
				expected:           40.0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{WorkingHoursPerWeek: tt.workingHoursPerWeek}
				result := employee.GetWeeklyTargetHours()
				if result != tt.expected {
					t.Errorf("GetWeeklyTargetHours() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})
}

func TestEmployeeOvertimeMethods(t *testing.T) {
	t.Run("FormatOvertimeBalance", func(t *testing.T) {
		tests := []struct {
			name            string
			overtimeBalance float64
			expected        string
		}{
			{
				name:            "positive overtime",
				overtimeBalance: 15.75,
				expected:        "+15.75 Std",
			},
			{
				name:            "negative overtime",
				overtimeBalance: -8.25,
				expected:        "-8.25 Std",
			},
			{
				name:            "zero overtime",
				overtimeBalance: 0.0,
				expected:        "+0.00 Std",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{OvertimeBalance: tt.overtimeBalance}
				result := employee.FormatOvertimeBalance()
				if result != tt.expected {
					t.Errorf("FormatOvertimeBalance() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("GetOvertimeStatus", func(t *testing.T) {
		tests := []struct {
			name            string
			overtimeBalance float64
			expected        string
		}{
			{
				name:            "positive overtime",
				overtimeBalance: 15.0,
				expected:        "positive",
			},
			{
				name:            "negative overtime",
				overtimeBalance: -8.0,
				expected:        "negative",
			},
			{
				name:            "zero overtime",
				overtimeBalance: 0.0,
				expected:        "neutral",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{OvertimeBalance: tt.overtimeBalance}
				result := employee.GetOvertimeStatus()
				if result != tt.expected {
					t.Errorf("GetOvertimeStatus() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("GetTotalAdjustments", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 5.0, Status: "approved"},
			{Hours: -2.0, Status: "approved"},
			{Hours: 3.0, Status: "pending"},
			{Hours: 1.0, Status: "rejected"},
			{Hours: 4.0, Status: "approved"},
		}

		employee := Employee{OvertimeAdjustments: adjustments}
		result := employee.GetTotalAdjustments()
		expected := 7.0 // 5.0 + (-2.0) + 4.0 = 7.0 (only approved)
		
		if result != expected {
			t.Errorf("GetTotalAdjustments() = %v, expected %v", result, expected)
		}
	})

	t.Run("GetApprovedAdjustments", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 5.0, Status: "approved"},
			{Hours: -2.0, Status: "approved"},
			{Hours: 3.0, Status: "pending"},
			{Hours: 1.0, Status: "rejected"},
		}

		employee := Employee{OvertimeAdjustments: adjustments}
		result := employee.GetApprovedAdjustments()
		
		if len(result) != 2 {
			t.Errorf("GetApprovedAdjustments() returned %d adjustments, expected 2", len(result))
		}

		// Check that only approved adjustments are returned
		for _, adj := range result {
			if adj.Status != "approved" {
				t.Errorf("GetApprovedAdjustments() returned non-approved adjustment with status %s", adj.Status)
			}
		}
	})

	t.Run("GetAdjustedOvertimeBalance", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 5.0, Status: "approved"},
			{Hours: -2.0, Status: "approved"},
			{Hours: 3.0, Status: "pending"}, // Should not be included
		}

		employee := Employee{
			OvertimeBalance:     10.0,
			OvertimeAdjustments: adjustments,
		}
		
		result := employee.GetAdjustedOvertimeBalance()
		expected := 13.0 // 10.0 + 5.0 + (-2.0) = 13.0
		
		if result != expected {
			t.Errorf("GetAdjustedOvertimeBalance() = %v, expected %v", result, expected)
		}
	})

	t.Run("FormatAdjustedOvertimeBalance", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 5.0, Status: "approved"},
		}

		tests := []struct {
			name            string
			baseBalance     float64
			expected        string
		}{
			{
				name:        "positive adjusted balance",
				baseBalance: 10.0,
				expected:    "+15.00 Std",
			},
			{
				name:        "negative adjusted balance",
				baseBalance: -10.0,
				expected:    "-5.00 Std",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				employee := Employee{
					OvertimeBalance:     tt.baseBalance,
					OvertimeAdjustments: adjustments,
				}
				result := employee.FormatAdjustedOvertimeBalance()
				if result != tt.expected {
					t.Errorf("FormatAdjustedOvertimeBalance() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("CalculateFinalOvertimeBalance", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 3.0, Status: "approved"},
			{Hours: -1.0, Status: "approved"},
			{Hours: 5.0, Status: "pending"}, // Should not be included
		}

		employee := Employee{
			OvertimeBalance:     12.5,
			OvertimeAdjustments: adjustments,
		}
		
		result := employee.CalculateFinalOvertimeBalance()
		expected := 14.5 // 12.5 + 3.0 + (-1.0) = 14.5
		
		if result != expected {
			t.Errorf("CalculateFinalOvertimeBalance() = %v, expected %v", result, expected)
		}
	})

	t.Run("GetOvertimeBalanceWithDetails", func(t *testing.T) {
		adjustments := []OvertimeAdjustment{
			{Hours: 5.0, Status: "approved"},
			{Hours: -2.0, Status: "approved"},
			{Hours: 3.0, Status: "pending"},
		}

		employee := Employee{
			OvertimeBalance:     10.0,
			OvertimeAdjustments: adjustments,
		}
		
		result := employee.GetOvertimeBalanceWithDetails()
		
		if result["baseBalance"] != 10.0 {
			t.Errorf("Expected baseBalance 10.0, got %v", result["baseBalance"])
		}
		if result["adjustmentsTotal"] != 3.0 {
			t.Errorf("Expected adjustmentsTotal 3.0, got %v", result["adjustmentsTotal"])
		}
		if result["finalBalance"] != 13.0 {
			t.Errorf("Expected finalBalance 13.0, got %v", result["finalBalance"])
		}
		if result["adjustmentCount"] != 2 {
			t.Errorf("Expected adjustmentCount 2, got %v", result["adjustmentCount"])
		}
	})

	t.Run("UpdateOvertimeBalance", func(t *testing.T) {
		employee := Employee{}
		newBalance := 25.5
		
		beforeUpdate := time.Now()
		employee.UpdateOvertimeBalance(newBalance)
		afterUpdate := time.Now()
		
		if employee.OvertimeBalance != newBalance {
			t.Errorf("OvertimeBalance = %v, expected %v", employee.OvertimeBalance, newBalance)
		}
		
		if employee.LastTimeCalculated.Before(beforeUpdate) || employee.LastTimeCalculated.After(afterUpdate) {
			t.Error("LastTimeCalculated should be set to current time")
		}
	})
}

func TestWorkTimeModelMethods(t *testing.T) {
	t.Run("GetDisplayName", func(t *testing.T) {
		tests := []struct {
			model    WorkTimeModel
			expected string
		}{
			{WorkTimeModelFullTime, "Vollzeit"},
			{WorkTimeModelPartTime, "Teilzeit"},
			{WorkTimeModelFlexTime, "Gleitzeit"},
			{WorkTimeModelRemote, "Remote/Homeoffice"},
			{WorkTimeModelShift, "Schichtarbeit"},
			{WorkTimeModelContract, "Werkvertrag"},
			{WorkTimeModelInternship, "Praktikum"},
			{WorkTimeModel("unknown"), "unknown"},
		}

		for _, tt := range tests {
			t.Run(string(tt.model), func(t *testing.T) {
				result := tt.model.GetDisplayName()
				if result != tt.expected {
					t.Errorf("GetDisplayName() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})
}

func TestEmployeeConstants(t *testing.T) {
	t.Run("EmployeeStatus constants", func(t *testing.T) {
		// Test that constants are defined
		statuses := []EmployeeStatus{
			EmployeeStatusActive,
			EmployeeStatusInactive,
			EmployeeStatusOnLeave,
			EmployeeStatusRemote,
		}

		for _, status := range statuses {
			if string(status) == "" {
				t.Errorf("Employee status constant should not be empty: %v", status)
			}
		}
	})

	t.Run("Department constants", func(t *testing.T) {
		// Test that constants are defined
		departments := []Department{
			DepartmentIT,
			DepartmentSales,
			DepartmentHR,
			DepartmentMarketing,
			DepartmentFinance,
			DepartmentProduction,
		}

		for _, dept := range departments {
			if string(dept) == "" {
				t.Errorf("Department constant should not be empty: %v", dept)
			}
		}
	})

	t.Run("WorkTimeModel constants", func(t *testing.T) {
		// Test that constants are defined
		models := []WorkTimeModel{
			WorkTimeModelFullTime,
			WorkTimeModelPartTime,
			WorkTimeModelFlexTime,
			WorkTimeModelRemote,
			WorkTimeModelShift,
			WorkTimeModelContract,
			WorkTimeModelInternship,
		}

		for _, model := range models {
			if string(model) == "" {
				t.Errorf("WorkTimeModel constant should not be empty: %v", model)
			}
		}
	})
}

func TestEmployeeStructFields(t *testing.T) {
	t.Run("Employee struct initialization", func(t *testing.T) {
		now := time.Now()
		employeeID := "EMP001"
		managerID := primitive.NewObjectID()

		employee := Employee{
			ID:                   primitive.NewObjectID(),
			EmployeeID:           employeeID,
			FirstName:            "John",
			LastName:             "Doe",
			Email:                "john.doe@company.com",
			Phone:                "+1234567890",
			InternalPhone:        "1234",
			InternalExtension:    "ext123",
			Address:              "123 Main St",
			DateOfBirth:          now.AddDate(-30, 0, 0),
			HireDate:             now.AddDate(-2, 0, 0),
			Position:             "Software Developer",
			Department:           DepartmentIT,
			ManagerID:            managerID,
			Status:               EmployeeStatusActive,
			WorkingHoursPerWeek:  40.0,
			WorkingDaysPerWeek:   5,
			WorkTimeModel:        WorkTimeModelFullTime,
			FlexibleWorkingHours: true,
			CoreWorkingTimeStart: "09:00",
			CoreWorkingTimeEnd:   "15:00",
			OvertimeBalance:      10.5,
			LastTimeCalculated:   now,
			Salary:               75000.0,
			BankAccount:          "DE89370400440532013000",
			TaxID:                "123456789",
			SocialSecID:          "987654321",
			HealthInsurance:      "AOK",
			EmergencyName:        "Jane Doe",
			EmergencyPhone:       "+0987654321",
			VacationDays:         30,
			RemainingVacation:    25,
			ProfileImage:         "/images/john_doe.jpg",
			Notes:                "Excellent performer",
			TimebutlerUserID:     "tb_123",
			Erfasst123ID:         "erf_456",
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		// Test that all fields are properly set
		if employee.EmployeeID != employeeID {
			t.Errorf("EmployeeID = %q, expected %q", employee.EmployeeID, employeeID)
		}
		if employee.Department != DepartmentIT {
			t.Errorf("Department = %v, expected %v", employee.Department, DepartmentIT)
		}
		if employee.ManagerID != managerID {
			t.Errorf("ManagerID = %v, expected %v", employee.ManagerID, managerID)
		}
		if employee.WorkTimeModel != WorkTimeModelFullTime {
			t.Errorf("WorkTimeModel = %v, expected %v", employee.WorkTimeModel, WorkTimeModelFullTime)
		}
		if !employee.FlexibleWorkingHours {
			t.Error("FlexibleWorkingHours should be true")
		}
		if employee.OvertimeBalance != 10.5 {
			t.Errorf("OvertimeBalance = %v, expected %v", employee.OvertimeBalance, 10.5)
		}
	})
}

func TestEmployeeNestedStructs(t *testing.T) {
	t.Run("WeeklyTimeEntry", func(t *testing.T) {
		now := time.Now()
		entry := WeeklyTimeEntry{
			ID:            primitive.NewObjectID(),
			WeekStartDate: now,
			WeekEndDate:   now.AddDate(0, 0, 6),
			Year:          2024,
			WeekNumber:    15,
			PlannedHours:  40.0,
			ActualHours:   42.5,
			OvertimeHours: 2.5,
			DaysWorked:    5,
			IsComplete:    true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if entry.OvertimeHours != 2.5 {
			t.Errorf("OvertimeHours = %v, expected %v", entry.OvertimeHours, 2.5)
		}
		if entry.WeekNumber != 15 {
			t.Errorf("WeekNumber = %v, expected %v", entry.WeekNumber, 15)
		}
		if !entry.IsComplete {
			t.Error("IsComplete should be true")
		}
	})

	t.Run("Document", func(t *testing.T) {
		now := time.Now()
		uploaderID := primitive.NewObjectID()
		
		doc := Document{
			ID:          primitive.NewObjectID(),
			Name:        "Employee Contract",
			FileName:    "contract.pdf",
			FileType:    "application/pdf",
			Description: "Employment contract for John Doe",
			Category:    "contracts",
			FilePath:    "/uploads/contracts/contract.pdf",
			FileSize:    1024000,
			UploadDate:  now,
			UploadedBy:  uploaderID,
		}

		if doc.FileSize != 1024000 {
			t.Errorf("FileSize = %v, expected %v", doc.FileSize, 1024000)
		}
		if doc.Category != "contracts" {
			t.Errorf("Category = %q, expected %q", doc.Category, "contracts")
		}
		if doc.UploadedBy != uploaderID {
			t.Errorf("UploadedBy = %v, expected %v", doc.UploadedBy, uploaderID)
		}
	})

	t.Run("Training", func(t *testing.T) {
		now := time.Now()
		training := Training{
			ID:          primitive.NewObjectID(),
			Title:       "Go Programming",
			Description: "Advanced Go programming course",
			StartDate:   now,
			EndDate:     now.AddDate(0, 0, 5),
			Provider:    "Tech Academy",
			Certificate: "GO-ADV-2024",
			Status:      "completed",
			Notes:       "Excellent participation",
		}

		if training.Status != "completed" {
			t.Errorf("Training Status = %q, expected %q", training.Status, "completed")
		}
		if training.Provider != "Tech Academy" {
			t.Errorf("Training Provider = %q, expected %q", training.Provider, "Tech Academy")
		}
	})

	t.Run("Absence", func(t *testing.T) {
		now := time.Now()
		approverID := primitive.NewObjectID()
		
		absence := Absence{
			ID:           primitive.NewObjectID(),
			Type:         "vacation",
			StartDate:    now,
			EndDate:      now.AddDate(0, 0, 5),
			Days:         5.0,
			Status:       "approved",
			ApprovedBy:   approverID,
			ApproverName: "Manager Smith",
			Reason:       "Annual vacation",
			Notes:        "Enjoy your time off!",
		}

		if absence.Type != "vacation" {
			t.Errorf("Absence Type = %q, expected %q", absence.Type, "vacation")
		}
		if absence.Days != 5.0 {
			t.Errorf("Absence Days = %v, expected %v", absence.Days, 5.0)
		}
		if absence.Status != "approved" {
			t.Errorf("Absence Status = %q, expected %q", absence.Status, "approved")
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || 
		len(str) > len(substr) && 
		(str[:len(substr)] == substr || 
		 str[len(str)-len(substr):] == substr ||
		 findSubstring(str, substr)))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkEmployeeOvertimeCalculations(b *testing.B) {
	adjustments := []OvertimeAdjustment{
		{Hours: 5.0, Status: "approved"},
		{Hours: -2.0, Status: "approved"},
		{Hours: 3.0, Status: "pending"},
		{Hours: 1.0, Status: "approved"},
	}

	employee := Employee{
		OvertimeBalance:     15.0,
		OvertimeAdjustments: adjustments,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = employee.GetTotalAdjustments()
		_ = employee.GetAdjustedOvertimeBalance()
		_ = employee.CalculateFinalOvertimeBalance()
	}
}

func BenchmarkEmployeeWorkingTimeCalculations(b *testing.B) {
	employee := Employee{
		WorkingHoursPerWeek: 40.0,
		WorkingDaysPerWeek:  5,
		WorkTimeModel:       WorkTimeModelFullTime,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = employee.GetWorkingHoursPerDay()
		_ = employee.IsFullTimeEmployee()
		_ = employee.GetWorkingTimeDescription()
		_ = employee.GetWeeklyTargetHours()
	}
}