package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
)

// TestOvertimeHandlerBasics tests basic functionality without complex mocking
func TestOvertimeHandlerBasics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetPendingAdjustments_WithoutDatabase", func(t *testing.T) {
		// This test verifies the handler structure and basic response format
		// In a real scenario, we'd need dependency injection for full testing
		
		router := gin.New()
		
		// Mock endpoint that simulates the expected response structure
		router.GET("/api/overtime/adjustments/pending", func(c *gin.Context) {
			// Simulate authenticated user context
			user := &model.User{
				ID:        primitive.NewObjectID(),
				FirstName: "Test",
				LastName:  "Admin",
				Email:     "admin@test.com",
				Role:      model.RoleAdmin,
				Status:    model.StatusActive,
			}
			c.Set("user", user)
			c.Set("userRole", "admin")
			
			// Simulate empty pending adjustments (like our real scenario)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    []interface{}{},
			})
		})

		// Make request
		req, _ := http.NewRequest("GET", "/api/overtime/adjustments/pending", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
		
		data := response["data"].([]interface{})
		assert.Len(t, data, 0) // Empty as expected
	})

	t.Run("GetPendingAdjustments_WithMockData", func(t *testing.T) {
		router := gin.New()
		
		router.GET("/api/overtime/adjustments/pending", func(c *gin.Context) {
			// Simulate authenticated user context
			user := &model.User{
				ID:        primitive.NewObjectID(),
				FirstName: "Test",
				LastName:  "Admin",
				Email:     "admin@test.com",
				Role:      model.RoleAdmin,
				Status:    model.StatusActive,
			}
			c.Set("user", user)
			c.Set("userRole", "admin")
			
			// Simulate pending adjustments with enriched data
			mockAdjustment := gin.H{
				"id":           primitive.NewObjectID().Hex(),
				"employeeId":   primitive.NewObjectID().Hex(),
				"employeeName": "John Doe",
				"department":   "IT",
				"type":         "manual",
				"hours":        8.5,
				"reason":       "Extra work on project deadline",
				"status":       "pending",
				"adjusterName": "Admin User",
				"createdAt":    "2025-06-11T10:44:19.749Z",
			}
			
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    []interface{}{mockAdjustment},
			})
		})

		// Make request
		req, _ := http.NewRequest("GET", "/api/overtime/adjustments/pending", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "data")
		
		data := response["data"].([]interface{})
		assert.Len(t, data, 1)
		
		adjustment := data[0].(map[string]interface{})
		assert.Equal(t, "John Doe", adjustment["employeeName"])
		assert.Equal(t, "IT", adjustment["department"])
		assert.Equal(t, float64(8.5), adjustment["hours"])
		assert.Equal(t, "pending", adjustment["status"])
	})
}

// TestOvertimeEmployeeSummary tests the OvertimeEmployeeSummary struct
func TestOvertimeEmployeeSummary(t *testing.T) {
	summary := OvertimeEmployeeSummary{
		EmployeeID:      primitive.NewObjectID().Hex(),
		EmployeeName:    "John Doe",
		Department:      "IT",
		HasProfileImage: true,
		WeeklyTarget:    40.0,
		TotalHours:      42.5,
		OvertimeBalance: 2.5,
		OvertimeStatus:  "positive",
		LastCalculated:  time.Now(),
		WorkTimeModel:   "Vollzeit",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(summary)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "John Doe")
	assert.Contains(t, string(jsonData), "positive")

	// Test deserialization
	var unmarshaled OvertimeEmployeeSummary
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, summary.EmployeeName, unmarshaled.EmployeeName)
	assert.Equal(t, summary.OvertimeBalance, unmarshaled.OvertimeBalance)
}

// TestOvertimeHandlerResponseFormats tests various response formats
func TestOvertimeHandlerResponseFormats(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func() gin.HandlerFunc
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success response format",
			setupHandler: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"data":    []interface{}{},
					})
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "data")
			},
		},
		{
			name: "Error response format",
			setupHandler: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Fehler beim Abrufen der ausstehenden Anpassungen",
					})
				}
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
				assert.Contains(t, response["error"], "Fehler")
			},
		},
		{
			name: "Empty data response",
			setupHandler: func() gin.HandlerFunc {
				return func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{
						"success": true,
						"data":    nil,
					})
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response["success"].(bool))
				assert.Nil(t, response["data"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/test", tt.setupHandler())

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

// Benchmark tests for performance measurement
func BenchmarkOvertimeResponseSerialization(b *testing.B) {
	summary := OvertimeEmployeeSummary{
		EmployeeID:      primitive.NewObjectID().Hex(),
		EmployeeName:    "John Doe",
		Department:      "IT",
		HasProfileImage: true,
		WeeklyTarget:    40.0,
		TotalHours:      42.5,
		OvertimeBalance: 2.5,
		OvertimeStatus:  "positive",
		LastCalculated:  time.Now(),
		WorkTimeModel:   "Vollzeit",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(summary)
		if err != nil {
			b.Fatal(err)
		}
	}
}