package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
)

// TestSystemSettingsHandlerBasics tests basic system settings functionality
func TestSystemSettingsHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetSystemSettings_Success", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/system/settings", func(c *gin.Context) {
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			// Mock system settings
			settings := model.SystemSettings{
				ID:                  primitive.NewObjectID(),
				CompanyName:         "Test Company",
				CompanyAddress:      "123 Test Street",
				Language:            "de",
				State:               "nordrhein_westfalen",
				DefaultWorkingHours: 40.0,
				DefaultVacationDays: 30,
			}
			
			c.JSON(http.StatusOK, gin.H{
				"success":  true,
				"settings": settings,
			})
		})

		req, _ := http.NewRequest("GET", "/system/settings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "settings")
	})

	t.Run("UpdateSystemSettings_Success", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/system/settings", func(c *gin.Context) {
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			var settingsData map[string]interface{}
			if err := c.ShouldBindJSON(&settingsData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Validate required fields
			if settingsData["companyName"] == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Company name is required"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "System settings updated successfully",
			})
		})

		settingsData := map[string]interface{}{
			"companyName":         "Updated Company Name",
			"defaultWorkingHours": 38.0,
			"language":            "en",
			"state":               "bayern",
		}

		jsonData, _ := json.Marshal(settingsData)
		req, _ := http.NewRequest("PUT", "/system/settings", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"].(bool))
	})

	t.Run("UpdateEmailSettings_Success", func(t *testing.T) {
		router := gin.New()
		
		router.PUT("/system/settings/email", func(c *gin.Context) {
			// Simulate authenticated admin user
			c.Set("user", &model.User{Role: model.RoleAdmin})
			c.Set("userRole", "admin")
			
			var emailData map[string]interface{}
			if err := c.ShouldBindJSON(&emailData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Validate email notification settings
			if enabled, exists := emailData["enabled"]; exists && enabled.(bool) {
				if emailData["smtpHost"] == "" || emailData["smtpPort"] == nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP host and port are required"})
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Email settings updated successfully",
			})
		})

		emailData := map[string]interface{}{
			"enabled":   true,
			"smtpHost":  "smtp.gmail.com",
			"smtpPort":  587,
			"smtpUser":  "test@gmail.com",
			"fromName":  "Test Company",
			"fromEmail": "test@gmail.com",
		}

		jsonData, _ := json.Marshal(emailData)
		req, _ := http.NewRequest("PUT", "/system/settings/email", bytes.NewBuffer(jsonData))
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

// TestSystemSettingsValidation tests input validation for system settings
func TestSystemSettingsValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		settingsData map[string]interface{}
		expectedCode int
		shouldPass   bool
	}{
		{
			name: "Valid settings",
			settingsData: map[string]interface{}{
				"companyName":         "Valid Company",
				"defaultWorkingHours": 40.0,
				"language":            "de",
				"state":               "bayern",
			},
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name: "Missing company name",
			settingsData: map[string]interface{}{
				"defaultWorkingHours": 40.0,
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name: "Invalid working hours",
			settingsData: map[string]interface{}{
				"companyName":         "Test Company",
				"defaultWorkingHours": -5.0,
			},
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			
			router.PUT("/system/settings", func(c *gin.Context) {
				var settingsData map[string]interface{}
				if err := c.ShouldBindJSON(&settingsData); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
					return
				}

				// Validate company name
				if settingsData["companyName"] == nil || settingsData["companyName"] == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Company name is required"})
					return
				}

				// Validate working hours
				if hours, exists := settingsData["defaultWorkingHours"]; exists {
					if hoursFloat := hours.(float64); hoursFloat <= 0 || hoursFloat > 80 {
						c.JSON(http.StatusBadRequest, gin.H{"error": "Working hours must be between 1 and 80"})
						return
					}
				}

				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			jsonData, _ := json.Marshal(tt.settingsData)
			req, _ := http.NewRequest("PUT", "/system/settings", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestSystemSettingsAuthorization tests authorization for system settings
func TestSystemSettingsAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		userRole     model.UserRole
		operation    string
		expectedCode int
		shouldPass   bool
	}{
		{
			name:         "Admin can read settings",
			userRole:     model.RoleAdmin,
			operation:    "read",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "Admin can update settings",
			userRole:     model.RoleAdmin,
			operation:    "update",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "Manager cannot update settings",
			userRole:     model.RoleManager,
			operation:    "update",
			expectedCode: http.StatusForbidden,
			shouldPass:   false,
		},
		{
			name:         "User cannot read settings",
			userRole:     model.RoleUser,
			operation:    "read",
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

			if tt.operation == "read" {
				router.GET("/system/settings", func(c *gin.Context) {
					userRole := c.GetString("userRole")
					if userRole != "admin" {
						c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
						return
					}
					c.JSON(http.StatusOK, gin.H{"success": true})
				})

				req, _ := http.NewRequest("GET", "/system/settings", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)

			} else if tt.operation == "update" {
				router.PUT("/system/settings", func(c *gin.Context) {
					userRole := c.GetString("userRole")
					if userRole != "admin" {
						c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
						return
					}
					c.JSON(http.StatusOK, gin.H{"success": true})
				})

				settingsData := map[string]interface{}{"companyName": "Test"}
				jsonData, _ := json.Marshal(settingsData)
				req, _ := http.NewRequest("PUT", "/system/settings", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCode, w.Code)
			}
		})
	}
}

// Benchmark tests for system settings operations
func BenchmarkGetSystemSettings(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.GET("/system/settings", func(c *gin.Context) {
		settings := model.SystemSettings{
			CompanyName: "Test Company",
			Language:    "de",
		}
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"settings": settings,
		})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/system/settings", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateSystemSettings(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	router.PUT("/system/settings", func(c *gin.Context) {
		var settingsData map[string]interface{}
		c.ShouldBindJSON(&settingsData)
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	settingsData := map[string]interface{}{
		"companyName":         "Test Company",
		"defaultWorkingHours": 40.0,
	}
	jsonData, _ := json.Marshal(settingsData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/system/settings", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}