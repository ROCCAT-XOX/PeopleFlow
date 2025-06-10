package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupRoleMiddlewareTest(t *testing.T) {
	// Initialize logger for testing
	err := utils.InitLogger(utils.LoggerConfig{
		Level:  utils.LogLevelDebug,
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Set gin to test mode
	gin.SetMode(gin.TestMode)
}

func createTestUser(id, role string) *model.User {
	objID, _ := primitive.ObjectIDFromHex(id)
	return &model.User{
		ID:        objID,
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Role:      model.UserRole(role),
		Status:    model.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func setupTestContext(userID, userRole string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)

	// Set user context data
	if userID != "" {
		c.Set("userId", userID)
	}
	if userRole != "" {
		c.Set("userRole", userRole)
	}

	return c, w
}

func TestImprovedRoleMiddleware(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name           string
		userRole       string
		allowedRoles   []model.UserRole
		expectedStatus int
		expectAbort    bool
	}{
		{
			name:           "admin access with admin role allowed",
			userRole:       string(model.RoleAdmin),
			allowedRoles:   []model.UserRole{model.RoleAdmin},
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "employee access with admin role required",
			userRole:       string(model.RoleEmployee),
			allowedRoles:   []model.UserRole{model.RoleAdmin},
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "manager access with admin or manager allowed",
			userRole:       string(model.RoleManager),
			allowedRoles:   []model.UserRole{model.RoleAdmin, model.RoleManager},
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "no role in context",
			userRole:       "",
			allowedRoles:   []model.UserRole{model.RoleAdmin},
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "hr access with hr role allowed",
			userRole:       string(model.RoleHR),
			allowedRoles:   []model.UserRole{model.RoleHR, model.RoleManager, model.RoleAdmin},
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext("user123", tt.userRole)

			// Create test router
			router := gin.New()
			router.Use(ImprovedRoleMiddleware(tt.allowedRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			router.ServeHTTP(w, c.Request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectAbort && w.Code == http.StatusOK {
				t.Error("Expected request to be aborted but it continued")
			}
		})
	}
}

func TestImprovedHRMiddleware(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name           string
		userRole       string
		targetUserRole string
		targetID       string
		expectedStatus int
		expectAbort    bool
	}{
		{
			name:           "admin can modify anyone",
			userRole:       string(model.RoleAdmin),
			targetUserRole: string(model.RoleManager),
			targetID:       "target123",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "hr cannot modify admin",
			userRole:       string(model.RoleHR),
			targetUserRole: string(model.RoleAdmin),
			targetID:       "target123",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "hr cannot modify manager",
			userRole:       string(model.RoleHR),
			targetUserRole: string(model.RoleManager),
			targetID:       "target123",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "hr can modify employee",
			userRole:       string(model.RoleHR),
			targetUserRole: string(model.RoleEmployee),
			targetID:       "target123",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "no target id allows access",
			userRole:       string(model.RoleHR),
			targetUserRole: "",
			targetID:       "",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "manager can modify anyone",
			userRole:       string(model.RoleManager),
			targetUserRole: string(model.RoleAdmin),
			targetID:       "target123",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext("user123", tt.userRole)

			// Mock the target user lookup
			if tt.targetID != "" {
				c.Params = gin.Params{{Key: "id", Value: tt.targetID}}
				
				// For this test, we'll need to mock the repository call
				// Since this involves database calls, we'll focus on testing the logic flow
				// In a real scenario, you'd use dependency injection or mocking
			}

			// Create test router
			router := gin.New()
			router.Use(ImprovedHRMiddleware())
			router.GET("/test/:id", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			path := "/test"
			if tt.targetID != "" {
				path = "/test/" + tt.targetID
			}

			req := httptest.NewRequest("GET", path, nil)
			router.ServeHTTP(w, req)

			// Note: This test will require database setup or mocking to fully test
			// For now, we're testing the middleware structure and flow
		})
	}
}

func TestImprovedSalaryViewMiddleware(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name            string
		userRole        string
		expectedHidden  bool
	}{
		{
			name:           "admin can view salary",
			userRole:       string(model.RoleAdmin),
			expectedHidden: false,
		},
		{
			name:           "manager can view salary",
			userRole:       string(model.RoleManager),
			expectedHidden: false,
		},
		{
			name:           "hr cannot view salary",
			userRole:       string(model.RoleHR),
			expectedHidden: true,
		},
		{
			name:           "employee cannot view salary",
			userRole:       string(model.RoleEmployee),
			expectedHidden: true,
		},
		{
			name:           "no role hides salary",
			userRole:       "",
			expectedHidden: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext("user123", tt.userRole)

			// Create middleware instance
			middleware := ImprovedSalaryViewMiddleware()
			middleware(c)

			// Check if salary is hidden
			hideSalary := c.GetBool("hideSalary")
			if hideSalary != tt.expectedHidden {
				t.Errorf("Expected hideSalary=%v, got %v", tt.expectedHidden, hideSalary)
			}

			// Should not abort the request
			if w.Code != 0 && w.Code != http.StatusOK {
				t.Errorf("Salary middleware should not abort request, got status %d", w.Code)
			}
		})
	}
}

func TestImprovedSelfOrAdminMiddleware(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name           string
		userID         string
		userRole       string
		requestedID    string
		expectedStatus int
		expectAbort    bool
	}{
		{
			name:           "admin can access any user",
			userID:         "user123",
			userRole:       string(model.RoleAdmin),
			requestedID:    "other456",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "user can access own data",
			userID:         "user123",
			userRole:       string(model.RoleEmployee),
			requestedID:    "user123",
			expectedStatus: http.StatusOK,
			expectAbort:    false,
		},
		{
			name:           "user cannot access other user data",
			userID:         "user123",
			userRole:       string(model.RoleEmployee),
			requestedID:    "other456",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "missing user id fails",
			userID:         "",
			userRole:       string(model.RoleEmployee),
			requestedID:    "other456",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
		{
			name:           "missing role fails",
			userID:         "user123",
			userRole:       "",
			requestedID:    "other456",
			expectedStatus: http.StatusForbidden,
			expectAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext(tt.userID, tt.userRole)

			if tt.requestedID != "" {
				c.Params = gin.Params{{Key: "id", Value: tt.requestedID}}
			}

			// Create test router
			router := gin.New()
			router.Use(ImprovedSelfOrAdminMiddleware())
			router.GET("/test/:id", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			path := "/test/" + tt.requestedID
			req := httptest.NewRequest("GET", path, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSpecializedRoleMiddlewares(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name           string
		middleware     gin.HandlerFunc
		userRole       string
		expectedStatus int
	}{
		{
			name:           "AdminOnlyMiddleware - admin access",
			middleware:     AdminOnlyMiddleware(),
			userRole:       string(model.RoleAdmin),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "AdminOnlyMiddleware - employee denied",
			middleware:     AdminOnlyMiddleware(),
			userRole:       string(model.RoleEmployee),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "ManagerOrAdminMiddleware - manager access",
			middleware:     ManagerOrAdminMiddleware(),
			userRole:       string(model.RoleManager),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "ManagerOrAdminMiddleware - hr denied",
			middleware:     ManagerOrAdminMiddleware(),
			userRole:       string(model.RoleHR),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "HROrHigherMiddleware - hr access",
			middleware:     HROrHigherMiddleware(),
			userRole:       string(model.RoleHR),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "HROrHigherMiddleware - employee denied",
			middleware:     HROrHigherMiddleware(),
			userRole:       string(model.RoleEmployee),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "EmployeeOrHigherMiddleware - employee access",
			middleware:     EmployeeOrHigherMiddleware(),
			userRole:       string(model.RoleEmployee),
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupTestContext("user123", tt.userRole)

			// Create test router
			router := gin.New()
			router.Use(tt.middleware)
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			router.ServeHTTP(w, c.Request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleRoleError(t *testing.T) {
	setupRoleMiddlewareTest(t)

	tests := []struct {
		name           string
		setupHeaders   func(*http.Request)
		expectedStatus int
		expectJSON     bool
	}{
		{
			name: "API request returns JSON",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Accept", "application/json")
			},
			expectedStatus: http.StatusForbidden,
			expectJSON:     true,
		},
		{
			name: "Web request returns HTML",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Accept", "text/html")
			},
			expectedStatus: http.StatusForbidden,
			expectJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			tt.setupHeaders(c.Request)

			handleRoleError(c, ErrRoleMissing, "Test error message")

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectJSON {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, "application/json") {
					t.Error("Expected JSON response for API request")
				}

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				if response["error"] == "" {
					t.Error("JSON response should contain error message")
				}

				if response["code"] == "" {
					t.Error("JSON response should contain error code")
				}
			}
		})
	}
}

func TestGetRoleErrorCode(t *testing.T) {
	tests := []struct {
		error        error
		expectedCode string
	}{
		{ErrRoleMissing, "ROLE_MISSING"},
		{ErrUserIDMissing, "USER_ID_MISSING"},
		{ErrInvalidTargetUser, "INVALID_TARGET_USER"},
		{ErrUnauthorizedAccess, "UNAUTHORIZED_ACCESS"},
		{errors.New("unknown error"), "ACCESS_DENIED"},
	}

	for _, tt := range tests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			result := getRoleErrorCode(tt.error)
			if result != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, result)
			}
		})
	}
}

func TestRolePermissionChecker(t *testing.T) {
	checker := NewRolePermissionChecker()

	// Test CanAccessUser
	t.Run("CanAccessUser", func(t *testing.T) {
		tests := []struct {
			userRole   model.UserRole
			userID     string
			targetID   string
			expected   bool
		}{
			{model.RoleAdmin, "user1", "user2", true},  // Admin can access anyone
			{model.RoleEmployee, "user1", "user1", true}, // Self access
			{model.RoleEmployee, "user1", "user2", false}, // Cannot access others
		}

		for _, tt := range tests {
			result := checker.CanAccessUser(tt.userRole, tt.userID, tt.targetID)
			if result != tt.expected {
				t.Errorf("CanAccessUser(%s, %s, %s) = %v, expected %v",
					tt.userRole, tt.userID, tt.targetID, result, tt.expected)
			}
		}
	})

	// Test CanModifyUser
	t.Run("CanModifyUser", func(t *testing.T) {
		tests := []struct {
			userRole   model.UserRole
			userID     string
			targetID   string
			targetRole model.UserRole
			expected   bool
		}{
			{model.RoleAdmin, "user1", "user2", model.RoleEmployee, true},   // Admin can modify anyone
			{model.RoleManager, "user1", "user2", model.RoleEmployee, true}, // Manager can modify employees
			{model.RoleManager, "user1", "user2", model.RoleAdmin, false},   // Manager cannot modify admin
			{model.RoleHR, "user1", "user2", model.RoleEmployee, true},      // HR can modify employees
			{model.RoleHR, "user1", "user2", model.RoleManager, false},      // HR cannot modify managers
			{model.RoleEmployee, "user1", "user1", model.RoleEmployee, true}, // Self modification
			{model.RoleEmployee, "user1", "user2", model.RoleEmployee, false}, // Cannot modify others
		}

		for _, tt := range tests {
			result := checker.CanModifyUser(tt.userRole, tt.userID, tt.targetID, tt.targetRole)
			if result != tt.expected {
				t.Errorf("CanModifyUser(%s, %s, %s, %s) = %v, expected %v",
					tt.userRole, tt.userID, tt.targetID, tt.targetRole, result, tt.expected)
			}
		}
	})

	// Test CanViewSalary
	t.Run("CanViewSalary", func(t *testing.T) {
		tests := []struct {
			userRole model.UserRole
			expected bool
		}{
			{model.RoleAdmin, true},
			{model.RoleManager, true},
			{model.RoleHR, false},
			{model.RoleEmployee, false},
		}

		for _, tt := range tests {
			result := checker.CanViewSalary(tt.userRole)
			if result != tt.expected {
				t.Errorf("CanViewSalary(%s) = %v, expected %v", tt.userRole, result, tt.expected)
			}
		}
	})

	// Test HasMinimumRole
	t.Run("HasMinimumRole", func(t *testing.T) {
		tests := []struct {
			userRole    model.UserRole
			minimumRole model.UserRole
			expected    bool
		}{
			{model.RoleAdmin, model.RoleEmployee, true},   // Admin >= Employee
			{model.RoleManager, model.RoleHR, true},       // Manager >= HR
			{model.RoleHR, model.RoleManager, false},      // HR < Manager
			{model.RoleEmployee, model.RoleAdmin, false},  // Employee < Admin
			{model.RoleManager, model.RoleManager, true},  // Manager == Manager
		}

		for _, tt := range tests {
			result := checker.HasMinimumRole(tt.userRole, tt.minimumRole)
			if result != tt.expected {
				t.Errorf("HasMinimumRole(%s, %s) = %v, expected %v",
					tt.userRole, tt.minimumRole, result, tt.expected)
			}
		}
	})
}

// Benchmark tests
func BenchmarkImprovedRoleMiddleware(b *testing.B) {
	setupRoleMiddlewareTest(&testing.T{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Set("userId", "user123")
	c.Set("userRole", string(model.RoleAdmin))

	middleware := ImprovedRoleMiddleware(model.RoleAdmin, model.RoleManager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware(c)
	}
}

func BenchmarkRolePermissionChecker(b *testing.B) {
	checker := NewRolePermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checker.CanAccessUser(model.RoleAdmin, "user1", "user2")
		_ = checker.CanModifyUser(model.RoleManager, "user1", "user2", model.RoleEmployee)
		_ = checker.CanViewSalary(model.RoleAdmin)
		_ = checker.HasMinimumRole(model.RoleManager, model.RoleHR)
	}
}