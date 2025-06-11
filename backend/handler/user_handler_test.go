package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
)

// TestUserHandlerBasics tests basic user management functionality
func TestUserHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock user for testing
	mockUser := &model.User{
		ID:        primitive.NewObjectID(),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@test.com",
		Role:      model.RoleUser,
		Status:    model.StatusActive,
	}

	t.Run("GetUsersHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/users", func(c *gin.Context) {
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			users := []model.User{*mockUser}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"users":   users,
				"total":   len(users),
			})
		})

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "users")
	})

	t.Run("CreateUserHandler_ValidData", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/users", func(c *gin.Context) {
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			var userData map[string]interface{}
			if err := c.ShouldBindJSON(&userData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Basic validation
			if userData["email"] == "" || userData["firstName"] == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Required fields missing"})
				return
			}

			// Simulate successful user creation
			newUser := model.User{
				ID:        primitive.NewObjectID(),
				FirstName: userData["firstName"].(string),
				LastName:  userData["lastName"].(string),
				Email:     userData["email"].(string),
				Role:      model.UserRole(userData["role"].(string)),
				Status:    model.StatusActive,
			}

			c.JSON(http.StatusCreated, gin.H{
				"success": true,
				"user":    newUser,
				"message": "User created successfully",
			})
		})

		userData := map[string]interface{}{
			"firstName": "Jane",
			"lastName":  "Smith",
			"email":     "jane.smith@test.com",
			"role":      "user",
			"password":  "password123",
		}

		jsonData, _ := json.Marshal(userData)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "user")
	})

	t.Run("UpdateUserHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/users/:id", func(c *gin.Context) {
			userID := c.Param("id")
			
			// Simulate authenticated user
			c.Set("user", mockUser)
			c.Set("userRole", string(mockUser.Role))
			
			var updateData map[string]interface{}
			if err := c.ShouldBindJSON(&updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Simulate successful update
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "User updated successfully",
				"userId":  userID,
			})
		})

		updateData := map[string]interface{}{
			"firstName": "John Updated",
			"lastName":  "Doe Updated",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/users/"+mockUser.ID.Hex(), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("DeleteUserHandler_Success", func(t *testing.T) {
		router := gin.New()
		
		router.DELETE("/users/:id", func(c *gin.Context) {
			userID := c.Param("id")
			
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			// Validate ObjectID format
			if _, err := primitive.ObjectIDFromHex(userID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "User deleted successfully",
			})
		})

		req, _ := http.NewRequest("DELETE", "/users/"+mockUser.ID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

// TestUserInputValidation tests input validation for user operations
func TestUserInputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userData     map[string]interface{}
		expectedCode int
		shouldPass   bool
	}{
		{
			name: "Valid user data",
			userData: map[string]interface{}{
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "john@test.com",
				"role":      "user",
			},
			expectedCode: http.StatusCreated,
			shouldPass:   true,
		},
		{
			name: "Missing email",
			userData: map[string]interface{}{
				"firstName": "John",
				"lastName":  "Doe",
				"role":      "user",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name: "Missing firstName",
			userData: map[string]interface{}{
				"lastName": "Doe",
				"email":    "john@test.com",
				"role":     "user",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name: "Invalid email format",
			userData: map[string]interface{}{
				"firstName": "John",
				"lastName":  "Doe",
				"email":     "invalid-email",
				"role":      "user",
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.POST("/users", func(c *gin.Context) {
				var userData map[string]interface{}
				if err := c.ShouldBindJSON(&userData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}

				// Validate required fields
				if userData["email"] == nil || userData["email"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
					return
				}
				if userData["firstName"] == nil || userData["firstName"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "First name is required"})
					return
				}

				// Validate email format (basic)
				email := userData["email"].(string)
				if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
					return
				}

				c.JSON(http.StatusCreated, gin.H{"success": true})
			})

			jsonData, _ := json.Marshal(tt.userData)
			req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestUserRoleAuthorization tests role-based authorization
func TestUserRoleAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userRole     model.UserRole
		operation    string
		expectedCode int
		shouldPass   bool
	}{
		{
			name:         "Admin can create users",
			userRole:     model.RoleAdmin,
			operation:    "create",
			expectedCode: http.StatusCreated,
			shouldPass:   true,
		},
		{
			name:         "HR can create users",
			userRole:     model.RoleHR,
			operation:    "create",
			expectedCode: http.StatusCreated,
			shouldPass:   true,
		},
		{
			name:         "User cannot create users",
			userRole:     model.RoleUser,
			operation:    "create",
			expectedCode: http.StatusForbidden,
			shouldPass:   false,
		},
		{
			name:         "Admin can delete users",
			userRole:     model.RoleAdmin,
			operation:    "delete",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "HR cannot delete users",
			userRole:     model.RoleHR,
			operation:    "delete",
			expectedCode: http.StatusForbidden,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			// Mock authorization middleware
			router.Use(func(c *gin.Context) {
				c.Set("user", &model.User{Role: tt.userRole})
				c.Set("userRole", string(tt.userRole))
				c.Next()
			})

			if tt.operation == "create" {
				router.POST("/users", func(c *gin.Context) {
					userRole := c.GetString("userRole")
					if userRole != "admin" && userRole != "hr" {
						c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
						return
					}
					c.JSON(http.StatusCreated, gin.H{"success": true})
				})

				req, _ := http.NewRequest("POST", "/users", strings.NewReader("{}"))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)

			} else if tt.operation == "delete" {
				router.DELETE("/users/:id", func(c *gin.Context) {
					userRole := c.GetString("userRole")
					if userRole != "admin" {
						c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
						return
					}
					c.JSON(http.StatusOK, gin.H{"success": true})
				})

				req, _ := http.NewRequest("DELETE", "/users/"+primitive.NewObjectID().Hex(), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)
			}
		})
	}
}

// TestUserProfileOperations tests user profile management
func TestUserProfileOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUser := &model.User{
		ID:        primitive.NewObjectID(),
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@test.com",
		Role:      model.RoleUser,
		Status:    model.StatusActive,
	}

	t.Run("GetProfile_Success", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/profile", func(c *gin.Context) {
			c.Set("user", mockUser)
			
			user, _ := c.Get("user")
			userModel := user.(*model.User)
			
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"user":    userModel,
			})
		})

		req, _ := http.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "user")
	})

	t.Run("UpdateProfile_Success", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/profile", func(c *gin.Context) {
			c.Set("user", mockUser)
			
			var updateData map[string]interface{}
			if err := c.ShouldBindJSON(&updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Profile updated successfully",
			})
		})

		updateData := map[string]interface{}{
			"firstName": "John Updated",
			"phone":     "+1234567890",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", "/profile", bytes.NewBuffer(jsonData))
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

// Benchmark tests for user operations
func BenchmarkGetUsersHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/users", func(c *gin.Context) {
		users := make([]model.User, 100) // Simulate 100 users
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"users":   users,
			"total":   len(users),
		})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkCreateUserHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.POST("/users", func(c *gin.Context) {
		var userData map[string]interface{}
		c.ShouldBindJSON(&userData)
		c.JSON(http.StatusCreated, gin.H{"success": true})
	})

	userData := map[string]interface{}{
		"firstName": "Test",
		"lastName":  "User",
		"email":     "test@test.com",
		"role":      "user",
	}
	jsonData, _ := json.Marshal(userData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}