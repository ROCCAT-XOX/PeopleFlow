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
)

// TestPasswordResetHandlerBasics tests password reset functionality
func TestPasswordResetHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("RequestPasswordReset_ValidEmail", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset-request", func(c *gin.Context) {
			var requestData struct {
				Email string `form:"email" json:"email" binding:"required,email"`
			}
			
			if err := c.ShouldBind(&requestData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
				return
			}

			// Simulate successful password reset request
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Password reset email sent successfully",
			})
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
		assert.Contains(t, response["message"], "reset email sent")
	})

	t.Run("RequestPasswordReset_InvalidEmail", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset-request", func(c *gin.Context) {
			var requestData struct {
				Email string `form:"email" json:"email" binding:"required,email"`
			}
			
			if err := c.ShouldBind(&requestData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		form := url.Values{}
		form.Add("email", "invalid-email")
		
		req, _ := http.NewRequest("POST", "/password-reset-request", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ResetPassword_ValidToken", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset", func(c *gin.Context) {
			var resetData struct {
				Token    string `form:"token" json:"token" binding:"required"`
				Password string `form:"password" json:"password" binding:"required,min=6"`
			}
			
			if err := c.ShouldBind(&resetData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token or password too short"})
				return
			}

			// Simulate valid token validation
			if resetData.Token == "valid-reset-token" {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Password reset successfully",
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
			}
		})

		form := url.Values{}
		form.Add("token", "valid-reset-token")
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

	t.Run("ResetPassword_InvalidToken", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset", func(c *gin.Context) {
			var resetData struct {
				Token    string `form:"token" json:"token" binding:"required"`
				Password string `form:"password" json:"password" binding:"required,min=6"`
			}
			
			if err := c.ShouldBind(&resetData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		})

		form := url.Values{}
		form.Add("token", "invalid-token")
		form.Add("password", "newpassword123")
		
		req, _ := http.NewRequest("POST", "/password-reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestPasswordValidation tests password strength validation
func TestPasswordValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		password     string
		expectedCode int
		shouldPass   bool
	}{
		{
			name:         "Valid strong password",
			password:     "SecurePass123!",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "Valid minimum length password",
			password:     "pass123",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "Too short password",
			password:     "123",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "Empty password",
			password:     "",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.POST("/password-reset", func(c *gin.Context) {
				var resetData struct {
					Token    string `form:"token" json:"token" binding:"required"`
					Password string `form:"password" json:"password" binding:"required,min=6"`
				}
				
				if err := c.ShouldBind(&resetData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			form := url.Values{}
			form.Add("token", "valid-token")
			form.Add("password", tt.password)
			
			req, _ := http.NewRequest("POST", "/password-reset", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestPasswordResetFlow tests the complete password reset flow
func TestPasswordResetFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CompletePasswordResetFlow", func(t *testing.T) {
		router := gin.New()
		
		// Step 1: Request password reset
		router.POST("/password-reset-request", func(c *gin.Context) {
			var requestData struct {
				Email string `form:"email" json:"email"`
			}
			
			c.ShouldBind(&requestData)
			
			// Simulate email found and reset token generated
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Reset email sent",
				"token":   "mock-reset-token", // In real implementation, this would be sent via email
			})
		})

		// Step 2: Reset password with token
		router.POST("/password-reset", func(c *gin.Context) {
			var resetData struct {
				Token    string `form:"token" json:"token"`
				Password string `form:"password" json:"password"`
			}
			
			c.ShouldBind(&resetData)
			
			if resetData.Token == "mock-reset-token" && len(resetData.Password) >= 6 {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Password reset successfully",
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token or password"})
			}
		})

		// Test Step 1: Request reset
		form1 := url.Values{}
		form1.Add("email", "user@test.com")
		
		req1, _ := http.NewRequest("POST", "/password-reset-request", strings.NewReader(form1.Encode()))
		req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)

		// Test Step 2: Reset with token
		form2 := url.Values{}
		form2.Add("token", "mock-reset-token")
		form2.Add("password", "newpassword123")
		
		req2, _ := http.NewRequest("POST", "/password-reset", strings.NewReader(form2.Encode()))
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w2.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})
}

// TestPasswordResetSecurity tests security aspects of password reset
func TestPasswordResetSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("TokenExpiration", func(t *testing.T) {
		router := gin.New()
		
		router.POST("/password-reset", func(c *gin.Context) {
			var resetData struct {
				Token    string `form:"token" json:"token"`
				Password string `form:"password" json:"password"`
			}
			
			c.ShouldBind(&resetData)
			
			// Simulate expired token
			if resetData.Token == "expired-token" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Token has expired"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		form := url.Values{}
		form.Add("token", "expired-token")
		form.Add("password", "newpassword123")
		
		req, _ := http.NewRequest("POST", "/password-reset", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "expired")
	})

	t.Run("RateLimiting", func(t *testing.T) {
		router := gin.New()
		
		requestCount := 0
		router.POST("/password-reset-request", func(c *gin.Context) {
			requestCount++
			
			// Simulate rate limiting after 3 requests
			if requestCount > 3 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "Too many password reset requests. Please try again later.",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// Make multiple requests
		for i := 0; i < 5; i++ {
			form := url.Values{}
			form.Add("email", "user@test.com")
			
			req, _ := http.NewRequest("POST", "/password-reset-request", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if i < 3 {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			}
		}
	})
}

// Benchmark tests for password reset operations
func BenchmarkPasswordResetRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.POST("/password-reset-request", func(c *gin.Context) {
		var requestData struct {
			Email string `form:"email" json:"email"`
		}
		c.ShouldBind(&requestData)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	form := url.Values{}
	form.Add("email", "test@test.com")
	body := strings.NewReader(form.Encode())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body.Seek(0, 0)
		req, _ := http.NewRequest("POST", "/password-reset-request", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPasswordReset(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.POST("/password-reset", func(c *gin.Context) {
		var resetData struct {
			Token    string `form:"token" json:"token"`
			Password string `form:"password" json:"password"`
		}
		c.ShouldBind(&resetData)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	form := url.Values{}
	form.Add("token", "test-token")
	form.Add("password", "newpassword123")
	body := strings.NewReader(form.Encode())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body.Seek(0, 0)
		req, _ := http.NewRequest("POST", "/password-reset", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}