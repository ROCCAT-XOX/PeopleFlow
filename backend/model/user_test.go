package model

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserValidation(t *testing.T) {
	t.Run("ValidateName", func(t *testing.T) {
		tests := []struct {
			name      string
			user      User
			expectErr bool
		}{
			{
				name: "valid names",
				user: User{FirstName: "John", LastName: "Doe"},
				expectErr: false,
			},
			{
				name: "empty first name",
				user: User{FirstName: "", LastName: "Doe"},
				expectErr: true,
			},
			{
				name: "empty last name",
				user: User{FirstName: "John", LastName: ""},
				expectErr: true,
			},
			{
				name: "whitespace only first name",
				user: User{FirstName: "   ", LastName: "Doe"},
				expectErr: true,
			},
			{
				name: "whitespace only last name",
				user: User{FirstName: "John", LastName: "   "},
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.user.ValidateName()
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("ValidateEmail", func(t *testing.T) {
		tests := []struct {
			name      string
			email     string
			expectErr bool
		}{
			{
				name: "valid email",
				email: "test@example.com",
				expectErr: false,
			},
			{
				name: "valid email with subdomain",
				email: "user@mail.example.com",
				expectErr: false,
			},
			{
				name: "empty email",
				email: "",
				expectErr: true,
			},
			{
				name: "whitespace only email",
				email: "   ",
				expectErr: true,
			},
			{
				name: "invalid email no @",
				email: "testexample.com",
				expectErr: true,
			},
			{
				name: "invalid email no domain",
				email: "test@",
				expectErr: true,
			},
			{
				name: "invalid email no user",
				email: "@example.com",
				expectErr: true,
			},
			{
				name: "invalid email no TLD",
				email: "test@example",
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user := User{Email: tt.email}
				err := user.ValidateEmail()
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("ValidatePassword", func(t *testing.T) {
		tests := []struct {
			name      string
			password  string
			expectErr bool
		}{
			{
				name: "valid password",
				password: "password123",
				expectErr: false,
			},
			{
				name: "minimum length password",
				password: "123456",
				expectErr: false,
			},
			{
				name: "empty password",
				password: "",
				expectErr: true,
			},
			{
				name: "too short password",
				password: "12345",
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user := User{Password: tt.password}
				err := user.ValidatePassword()
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("ValidateRole", func(t *testing.T) {
		tests := []struct {
			name      string
			role      UserRole
			expectErr bool
		}{
			{
				name: "admin role",
				role: RoleAdmin,
				expectErr: false,
			},
			{
				name: "manager role",
				role: RoleManager,
				expectErr: false,
			},
			{
				name: "hr role",
				role: RoleHR,
				expectErr: false,
			},
			{
				name: "employee role",
				role: RoleEmployee,
				expectErr: false,
			},
			{
				name: "user role (legacy)",
				role: RoleUser,
				expectErr: false,
			},
			{
				name: "invalid role",
				role: UserRole("invalid"),
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user := User{Role: tt.role}
				err := user.ValidateRole()
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("ValidateStatus", func(t *testing.T) {
		tests := []struct {
			name      string
			status    UserStatus
			expectErr bool
		}{
			{
				name: "active status",
				status: StatusActive,
				expectErr: false,
			},
			{
				name: "inactive status",
				status: StatusInactive,
				expectErr: false,
			},
			{
				name: "invalid status",
				status: UserStatus("invalid"),
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				user := User{Status: tt.status}
				err := user.ValidateStatus()
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})

	t.Run("Validate", func(t *testing.T) {
		tests := []struct {
			name      string
			user      User
			isUpdate  bool
			expectErr bool
		}{
			{
				name: "valid new user",
				user: User{
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Password:  "password123",
					Role:      RoleEmployee,
					Status:    StatusActive,
				},
				isUpdate: false,
				expectErr: false,
			},
			{
				name: "valid update (minimal fields)",
				user: User{
					Role:   RoleManager,
					Status: StatusActive,
				},
				isUpdate: true,
				expectErr: false,
			},
			{
				name: "invalid new user (missing email)",
				user: User{
					FirstName: "John",
					LastName:  "Doe",
					Password:  "password123",
				},
				isUpdate: false,
				expectErr: true,
			},
			{
				name: "invalid role",
				user: User{
					Role: UserRole("invalid"),
				},
				isUpdate: true,
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.user.Validate(tt.isUpdate)
				if tt.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tt.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			})
		}
	})
}

func TestUserPasswordHandling(t *testing.T) {
	t.Run("HashPassword", func(t *testing.T) {
		user := User{Password: "testpassword"}
		
		err := user.HashPassword()
		if err != nil {
			t.Fatalf("HashPassword failed: %v", err)
		}

		// Check that password hash is set
		if user.PasswordHash == "" {
			t.Error("PasswordHash should be set after hashing")
		}

		// Check that plain password is cleared
		if user.Password != "" {
			t.Error("Password should be cleared after hashing")
		}

		// Check that hash is not the same as plain password
		if user.PasswordHash == "testpassword" {
			t.Error("PasswordHash should not be the same as plain password")
		}
	})

	t.Run("HashPassword with empty password", func(t *testing.T) {
		user := User{Password: ""}
		
		err := user.HashPassword()
		if err == nil {
			t.Error("Expected error for empty password")
		}
	})

	t.Run("CheckPassword", func(t *testing.T) {
		user := User{Password: "testpassword"}
		err := user.HashPassword()
		if err != nil {
			t.Fatalf("HashPassword failed: %v", err)
		}

		// Check correct password
		if !user.CheckPassword("testpassword") {
			t.Error("CheckPassword should return true for correct password")
		}

		// Check incorrect password
		if user.CheckPassword("wrongpassword") {
			t.Error("CheckPassword should return false for incorrect password")
		}

		// Check empty password
		if user.CheckPassword("") {
			t.Error("CheckPassword should return false for empty password")
		}
	})

	t.Run("CheckPassword with empty hash", func(t *testing.T) {
		user := User{PasswordHash: ""}
		
		if user.CheckPassword("anypassword") {
			t.Error("CheckPassword should return false when hash is empty")
		}
	})
}

func TestUserRoleMethods(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		isAdmin  bool
		isManager bool
		isHR     bool
		isEmployee bool
	}{
		{
			name: "admin user",
			role: RoleAdmin,
			isAdmin: true,
			isManager: false,
			isHR: false,
			isEmployee: false,
		},
		{
			name: "manager user",
			role: RoleManager,
			isAdmin: false,
			isManager: true,
			isHR: false,
			isEmployee: false,
		},
		{
			name: "hr user",
			role: RoleHR,
			isAdmin: false,
			isManager: false,
			isHR: true,
			isEmployee: false,
		},
		{
			name: "employee user",
			role: RoleEmployee,
			isAdmin: false,
			isManager: false,
			isHR: false,
			isEmployee: true,
		},
		{
			name: "legacy user role",
			role: RoleUser,
			isAdmin: false,
			isManager: false,
			isHR: false,
			isEmployee: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Role: tt.role}

			if user.IsAdmin() != tt.isAdmin {
				t.Errorf("IsAdmin() = %v, expected %v", user.IsAdmin(), tt.isAdmin)
			}
			if user.IsManager() != tt.isManager {
				t.Errorf("IsManager() = %v, expected %v", user.IsManager(), tt.isManager)
			}
			if user.IsHR() != tt.isHR {
				t.Errorf("IsHR() = %v, expected %v", user.IsHR(), tt.isHR)
			}
			if user.IsEmployee() != tt.isEmployee {
				t.Errorf("IsEmployee() = %v, expected %v", user.IsEmployee(), tt.isEmployee)
			}
		})
	}
}

func TestUserHasRole(t *testing.T) {
	user := User{Role: RoleManager}

	// Should have manager role
	if !user.HasRole(RoleManager) {
		t.Error("User should have manager role")
	}

	// Should have manager role when checking multiple roles
	if !user.HasRole(RoleAdmin, RoleManager, RoleHR) {
		t.Error("User should have one of the specified roles")
	}

	// Should not have admin role
	if user.HasRole(RoleAdmin) {
		t.Error("User should not have admin role")
	}

	// Should not have any of the specified roles
	if user.HasRole(RoleAdmin, RoleHR, RoleEmployee) {
		t.Error("User should not have any of the specified roles")
	}
}

func TestUserCanModifyUser(t *testing.T) {
	adminUser := User{ID: primitive.NewObjectID(), Role: RoleAdmin}
	managerUser := User{ID: primitive.NewObjectID(), Role: RoleManager}
	hrUser := User{ID: primitive.NewObjectID(), Role: RoleHR}
	employeeUser := User{ID: primitive.NewObjectID(), Role: RoleEmployee}
	otherEmployeeUser := User{ID: primitive.NewObjectID(), Role: RoleEmployee}

	tests := []struct {
		name       string
		user       User
		targetUser User
		canModify  bool
	}{
		{
			name: "admin can modify anyone",
			user: adminUser,
			targetUser: employeeUser,
			canModify: true,
		},
		{
			name: "admin can modify manager",
			user: adminUser,
			targetUser: managerUser,
			canModify: true,
		},
		{
			name: "manager can modify hr",
			user: managerUser,
			targetUser: hrUser,
			canModify: true,
		},
		{
			name: "manager can modify employee",
			user: managerUser,
			targetUser: employeeUser,
			canModify: true,
		},
		{
			name: "manager cannot modify admin",
			user: managerUser,
			targetUser: adminUser,
			canModify: false,
		},
		{
			name: "hr can modify employee",
			user: hrUser,
			targetUser: employeeUser,
			canModify: true,
		},
		{
			name: "hr cannot modify manager",
			user: hrUser,
			targetUser: managerUser,
			canModify: false,
		},
		{
			name: "hr cannot modify admin",
			user: hrUser,
			targetUser: adminUser,
			canModify: false,
		},
		{
			name: "employee can modify self",
			user: employeeUser,
			targetUser: employeeUser,
			canModify: true,
		},
		{
			name: "employee cannot modify other employee",
			user: employeeUser,
			targetUser: otherEmployeeUser,
			canModify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.CanModifyUser(&tt.targetUser)
			if result != tt.canModify {
				t.Errorf("CanModifyUser() = %v, expected %v", result, tt.canModify)
			}
		})
	}
}

func TestUserCanViewSalary(t *testing.T) {
	tests := []struct {
		name      string
		role      UserRole
		canView   bool
	}{
		{
			name: "admin can view salary",
			role: RoleAdmin,
			canView: true,
		},
		{
			name: "manager can view salary",
			role: RoleManager,
			canView: true,
		},
		{
			name: "hr cannot view salary",
			role: RoleHR,
			canView: false,
		},
		{
			name: "employee cannot view salary",
			role: RoleEmployee,
			canView: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Role: tt.role}
			result := user.CanViewSalary()
			if result != tt.canView {
				t.Errorf("CanViewSalary() = %v, expected %v", result, tt.canView)
			}
		})
	}
}

func TestUserUtilityMethods(t *testing.T) {
	t.Run("GetFullName", func(t *testing.T) {
		tests := []struct {
			name      string
			user      User
			expected  string
		}{
			{
				name: "both names provided",
				user: User{FirstName: "John", LastName: "Doe"},
				expected: "John Doe",
			},
			{
				name: "names with extra whitespace",
				user: User{FirstName: "  John  ", LastName: "  Doe  "},
				expected: "John     Doe",
			},
			{
				name: "empty names",
				user: User{FirstName: "", LastName: ""},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.user.GetFullName()
				if result != tt.expected {
					t.Errorf("GetFullName() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("GetDisplayName", func(t *testing.T) {
		tests := []struct {
			name      string
			user      User
			expected  string
		}{
			{
				name: "full name available",
				user: User{FirstName: "John", LastName: "Doe", Email: "john@example.com"},
				expected: "John Doe",
			},
			{
				name: "no name, use email",
				user: User{FirstName: "", LastName: "", Email: "john@example.com"},
				expected: "john@example.com",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.user.GetDisplayName()
				if result != tt.expected {
					t.Errorf("GetDisplayName() = %q, expected %q", result, tt.expected)
				}
			})
		}
	})

	t.Run("IsActive", func(t *testing.T) {
		now := time.Now()
		tests := []struct {
			name      string
			user      User
			expected  bool
		}{
			{
				name: "active user",
				user: User{Status: StatusActive},
				expected: true,
			},
			{
				name: "inactive user",
				user: User{Status: StatusInactive},
				expected: false,
			},
			{
				name: "deleted user",
				user: User{Status: StatusActive, DeletedAt: &now},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.user.IsActive()
				if result != tt.expected {
					t.Errorf("IsActive() = %v, expected %v", result, tt.expected)
				}
			})
		}
	})

	t.Run("NormalizeEmail", func(t *testing.T) {
		user := User{Email: "  John.Doe@EXAMPLE.COM  "}
		user.NormalizeEmail()
		expected := "john.doe@example.com"
		if user.Email != expected {
			t.Errorf("NormalizeEmail() result = %q, expected %q", user.Email, expected)
		}
	})

	t.Run("SetDefaults", func(t *testing.T) {
		user := User{}
		user.SetDefaults()

		if user.Status != StatusActive {
			t.Errorf("Default status should be active, got %v", user.Status)
		}
		if user.Role != RoleEmployee {
			t.Errorf("Default role should be employee, got %v", user.Role)
		}
		if user.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set")
		}
		if user.UpdatedAt.IsZero() {
			t.Error("UpdatedAt should be set")
		}
	})
}

func TestUserDisplayNames(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		status   UserStatus
		roleDisplay   string
		statusDisplay string
	}{
		{
			name: "admin role",
			role: RoleAdmin,
			status: StatusActive,
			roleDisplay: "Administrator",
			statusDisplay: "Aktiv",
		},
		{
			name: "manager role",
			role: RoleManager,
			status: StatusInactive,
			roleDisplay: "Manager",
			statusDisplay: "Inaktiv",
		},
		{
			name: "hr role",
			role: RoleHR,
			status: StatusActive,
			roleDisplay: "Personalverwaltung",
			statusDisplay: "Aktiv",
		},
		{
			name: "employee role",
			role: RoleEmployee,
			status: StatusActive,
			roleDisplay: "Mitarbeiter",
			statusDisplay: "Aktiv",
		},
		{
			name: "legacy user role",
			role: RoleUser,
			status: StatusActive,
			roleDisplay: "Mitarbeiter",
			statusDisplay: "Aktiv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Role: tt.role, Status: tt.status}

			roleDisplay := user.GetRoleDisplayName()
			if roleDisplay != tt.roleDisplay {
				t.Errorf("GetRoleDisplayName() = %q, expected %q", roleDisplay, tt.roleDisplay)
			}

			statusDisplay := user.GetStatusDisplayName()
			if statusDisplay != tt.statusDisplay {
				t.Errorf("GetStatusDisplayName() = %q, expected %q", statusDisplay, tt.statusDisplay)
			}
		})
	}
}

func TestUserPrepareForCreate(t *testing.T) {
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "  JOHN.DOE@EXAMPLE.COM  ",
		Password:  "password123",
	}

	err := user.PrepareForCreate()
	if err != nil {
		t.Fatalf("PrepareForCreate failed: %v", err)
	}

	// Check that email was normalized
	if user.Email != "john.doe@example.com" {
		t.Errorf("Email should be normalized, got %s", user.Email)
	}

	// Check that defaults were set
	if user.Status != StatusActive {
		t.Errorf("Status should be set to active, got %v", user.Status)
	}
	if user.Role != RoleEmployee {
		t.Errorf("Role should be set to employee, got %v", user.Role)
	}

	// Check that password was hashed
	if user.PasswordHash == "" {
		t.Error("PasswordHash should be set")
	}
	if user.Password != "" {
		t.Error("Password should be cleared")
	}

	// Check timestamps
	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUserPrepareForUpdate(t *testing.T) {
	user := User{
		Email:     "  UPDATED@EXAMPLE.COM  ",
		Password:  "newpassword",
		Role:      RoleManager,
		Status:    StatusActive,
	}

	err := user.PrepareForUpdate()
	if err != nil {
		t.Fatalf("PrepareForUpdate failed: %v", err)
	}

	// Check that email was normalized
	if user.Email != "updated@example.com" {
		t.Errorf("Email should be normalized, got %s", user.Email)
	}

	// Check that password was hashed
	if user.PasswordHash == "" {
		t.Error("PasswordHash should be set when password is provided")
	}
	if user.Password != "" {
		t.Error("Password should be cleared")
	}

	// Check that UpdatedAt was set
	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUserToJSON(t *testing.T) {
	employeeID := primitive.NewObjectID()
	lastLogin := time.Now()
	
	user := User{
		ID:         primitive.NewObjectID(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Role:       RoleEmployee,
		Status:     StatusActive,
		EmployeeID: &employeeID,
		LastLogin:  &lastLogin,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	json := user.ToJSON()

	// Check required fields
	if json["id"] != user.ID {
		t.Error("JSON should contain user ID")
	}
	if json["firstName"] != user.FirstName {
		t.Error("JSON should contain firstName")
	}
	if json["lastName"] != user.LastName {
		t.Error("JSON should contain lastName")
	}
	if json["email"] != user.Email {
		t.Error("JSON should contain email")
	}
	if json["role"] != user.Role {
		t.Error("JSON should contain role")
	}
	if json["status"] != user.Status {
		t.Error("JSON should contain status")
	}
	if json["fullName"] != user.GetFullName() {
		t.Error("JSON should contain fullName")
	}

	// Check optional fields
	if json["employeeId"] != user.EmployeeID {
		t.Error("JSON should contain employeeId when set")
	}
	if json["lastLogin"] != user.LastLogin {
		t.Error("JSON should contain lastLogin when set")
	}

	// Check that password is not included
	if _, exists := json["password"]; exists {
		t.Error("JSON should not contain password")
	}
	if _, exists := json["passwordHash"]; exists {
		t.Error("JSON should not contain passwordHash")
	}
}

// Benchmark tests
func BenchmarkUserValidation(b *testing.B) {
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Password:  "password123",
		Role:      RoleEmployee,
		Status:    StatusActive,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.Validate(false)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := User{Password: "testpassword"}
		_ = user.HashPassword()
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	user := User{Password: "testpassword"}
	_ = user.HashPassword()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.CheckPassword("testpassword")
	}
}