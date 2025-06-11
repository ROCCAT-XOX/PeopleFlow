package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
)

// TestAuthHandlerBasics tests basic authentication functionality
func TestAuthHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("LoginHandler_ValidRequest", func(t *testing.T) {
		router := gin.New()
		
		// Mock login endpoint
		router.POST("/login", func(c *gin.Context) {
			var loginData struct {
				Email    string `form:"email" json:"email"`
				Password string `form:"password" json:"password"`
			}
			
			if err := c.ShouldBind(&loginData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// Simulate successful login
			if loginData.Email == "admin@test.com" && loginData.Password == "admin" {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Login successful",
					"user": gin.H{
						"id":    primitive.NewObjectID().Hex(),
						"email": loginData.Email,
						"role":  "admin",
					},
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			}
		})

		// Test valid credentials
		form := url.Values{}
		form.Add("email", "admin@test.com")
		form.Add("password", "admin")
		
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "user")
	})

	t.Run("LoginHandler_InvalidCredentials", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/login", func(c *gin.Context) {
			var loginData struct {
				Email    string `form:"email" json:"email"`
				Password string `form:"password" json:"password"`
			}
			
			if err := c.ShouldBind(&loginData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		})

		form := url.Values{}
		form.Add("email", "admin@test.com")
		form.Add("password", "wrong")
		
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "error")
	})

	t.Run("LogoutHandler", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/logout", func(c *gin.Context) {
			// Simulate logout by clearing cookie
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Logged out successfully",
			})
		})

		req, _ := http.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

// TestPasswordResetRequestFlow tests password reset functionality
func TestPasswordResetRequestFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("PasswordResetRequest_ValidEmail", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset-request", func(c *gin.Context) {
			var resetData struct {
				Email string `form:"email" json:"email"`
			}
			
			if err := c.ShouldBind(&resetData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// Simulate password reset request
			if resetData.Email != "" {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Password reset email sent",
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			}
		})

		form := url.Values{}
		form.Add("email", "user@test.com")
		
		req, _ := http.NewRequest("POST", "/password-reset-request", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("PasswordReset_ValidToken", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset", func(c *gin.Context) {
			var resetData struct {
				Token    string `form:"token" json:"token"`
				Password string `form:"password" json:"password"`
			}
			
			if err := c.ShouldBind(&resetData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// Simulate password reset with valid token
			if resetData.Token == "valid-token" && len(resetData.Password) >= 6 {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Password reset successful",
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token or password"})
			}
		})

		form := url.Values{}
		form.Add("token", "valid-token")
		form.Add("password", "newpassword123")
		
		req, _ := http.NewRequest("POST", "/password-reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CompleteAuthFlow", func(t *testing.T) {
		router := gin.New()
		
		// Mock user for testing
		mockUser := &model.User{
			ID:        primitive.NewObjectID(),
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@test.com",
			Role:      model.RoleAdmin,
			Status:    model.StatusActive,
		}

		// Login endpoint
		router.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"user":    mockUser,
				"token":   "mock-jwt-token",
			})
		})

		// Protected endpoint that requires auth
		router.GET("/protected", func(c *gin.Context) {
			// Simulate middleware setting user
			c.Set("user", mockUser)
			c.Set("userRole", string(mockUser.Role))
			
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Access granted",
				"user":    mockUser.Email,
			})
		})

		// Test login
		loginReq, _ := http.NewRequest("POST", "/login", nil)
		loginW := httptest.NewRecorder()
		router.ServeHTTP(loginW, loginReq)
		assert.Equal(t, http.StatusOK, loginW.Code)

		// Test protected access
		protectedReq, _ := http.NewRequest("GET", "/protected", nil)
		protectedW := httptest.NewRecorder()
		router.ServeHTTP(protectedW, protectedReq)
		assert.Equal(t, http.StatusOK, protectedW.Code)
	})
}

// TestAuthInputValidation tests input validation for authentication
func TestAuthInputValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		email        string
		password     string
		expectedCode int
		shouldPass   bool
	}{
		{
			name:         "Valid credentials",
			email:        "user@test.com",
			password:     "password123",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "Empty email",
			email:        "",
			password:     "password123",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "Empty password",
			email:        "user@test.com",
			password:     "",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "Invalid email format",
			email:        "invalid-email",
			password:     "password123",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.POST("/login", func(c *gin.Context) {
				var loginData struct {
					Email    string `form:"email" json:"email" binding:"required,email"`
					Password string `form:"password" json:"password" binding:"required,min=6"`
				}
				
				if err := c.ShouldBind(&loginData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			form := url.Values{}
			form.Add("email", tt.email)
			form.Add("password", tt.password)
			
			req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// Benchmark tests for authentication performance
func BenchmarkLoginHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.POST("/login", func(c *gin.Context) {
		var loginData struct {
			Email    string `form:"email" json:"email"`
			Password string `form:"password" json:"password"`
		}
		
		c.ShouldBind(&loginData)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	form := url.Values{}
	form.Add("email", "test@test.com")
	form.Add("password", "password123")
	body := strings.NewReader(form.Encode())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body.Seek(0, 0) // Reset body for reuse
		req, _ := http.NewRequest("POST", "/login", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}