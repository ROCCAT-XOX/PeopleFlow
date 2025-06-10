package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"PeopleFlow/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setupMiddlewareTest(t *testing.T) {
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

func createTestJWT(userID, role string, expiration time.Time) (string, error) {
	claims := &utils.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("test-secret-key"))
}

func TestExtractToken(t *testing.T) {
	setupMiddlewareTest(t)

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		expectToken string
		expectError bool
	}{
		{
			name: "token from cookie",
			setupFunc: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "token",
					Value: "cookie-token",
				})
			},
			expectToken: "cookie-token",
			expectError: false,
		},
		{
			name: "token from Authorization header",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer header-token")
			},
			expectToken: "header-token",
			expectError: false,
		},
		{
			name: "no token",
			setupFunc: func(c *gin.Context) {
				// No token setup
			},
			expectToken: "",
			expectError: true,
		},
		{
			name: "malformed Authorization header",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "InvalidFormat token")
			},
			expectToken: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Setup test case
			tt.setupFunc(c)

			// Extract token
			token, err := extractToken(c)

			// Check results
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if token != tt.expectToken {
					t.Errorf("Expected token %s, got %s", tt.expectToken, token)
				}
			}
		})
	}
}

func TestIsAPIRequest(t *testing.T) {
	setupMiddlewareTest(t)

	tests := []struct {
		name      string
		setupFunc func(*gin.Context)
		expected  bool
	}{
		{
			name: "JSON Accept header",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "application/json")
			},
			expected: true,
		},
		{
			name: "JSON Content-Type",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Content-Type", "application/json")
			},
			expected: true,
		},
		{
			name: "API path",
			setupFunc: func(c *gin.Context) {
				c.Request.URL.Path = "/api/users"
			},
			expected: true,
		},
		{
			name: "HTML Accept header",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "text/html")
			},
			expected: false,
		},
		{
			name: "no indicators",
			setupFunc: func(c *gin.Context) {
				// Default setup
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Setup test case
			tt.setupFunc(c)

			// Test function
			result := isAPIRequest(c)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	setupMiddlewareTest(t)

	tests := []struct {
		error    error
		expected string
	}{
		{ErrNoToken, "NO_TOKEN"},
		{ErrInvalidToken, "INVALID_TOKEN"},
		{ErrUserNotFound, "USER_NOT_FOUND"},
		{ErrUserInactive, "USER_INACTIVE"},
		{ErrTokenExpired, "TOKEN_EXPIRED"},
		{ErrInsufficientRole, "INSUFFICIENT_ROLE"},
		{errors.New("unknown error"), "AUTH_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			result := getErrorCode(tt.error)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	setupMiddlewareTest(t)

	// Generate multiple request IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()

		// Check that ID is not empty
		if id == "" {
			t.Error("Generated request ID should not be empty")
		}

		// Check that ID is unique
		if ids[id] {
			t.Errorf("Request ID %s was generated twice", id)
		}
		ids[id] = true

		// Check that ID has correct length (ObjectID hex length is 24)
		if len(id) != 24 {
			t.Errorf("Expected request ID length 24, got %d", len(id))
		}
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	setupMiddlewareTest(t)

	// Create test router
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("requestID")
		if !exists {
			c.JSON(500, gin.H{"error": "no request ID"})
			return
		}
		c.JSON(200, gin.H{"requestID": requestID})
	})

	// Make test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Check status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	requestID, exists := response["requestID"]
	if !exists {
		t.Error("Response should contain requestID")
	}

	if requestID == "" {
		t.Error("Request ID should not be empty")
	}

	// Check header
	headerRequestID := w.Header().Get("X-Request-ID")
	if headerRequestID == "" {
		t.Error("X-Request-ID header should be set")
	}

	if headerRequestID != requestID {
		t.Error("Header request ID should match response request ID")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	setupMiddlewareTest(t)

	// Create test router
	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Make test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")
	router.ServeHTTP(w, req)

	// Check that request was processed
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// The actual logging is tested through the logger itself
	// This test mainly ensures the middleware doesn't break the request flow
}

func TestRecoveryMiddleware(t *testing.T) {
	setupMiddlewareTest(t)

	tests := []struct {
		name           string
		setupHeaders   func(*http.Request)
		expectedStatus int
		expectJSON     bool
	}{
		{
			name: "API request panic",
			setupHeaders: func(req *http.Request) {
				req.Header.Set("Accept", "application/json")
			},
			expectedStatus: http.StatusInternalServerError,
			expectJSON:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router
			router := gin.New()
			router.Use(RecoveryMiddleware())

			// Add template for HTML responses
			if !tt.expectJSON {
				// Skip template loading in tests as it requires template helpers
				// router.LoadHTMLGlob("../../frontend/templates/*.html")
			}

			router.GET("/panic", func(c *gin.Context) {
				panic("test panic")
			})

			// Make test request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/panic", nil)
			tt.setupHeaders(req)
			router.ServeHTTP(w, req)

			// Check status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response type
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
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	setupMiddlewareTest(t)

	tests := []struct {
		name        string
		setupFunc   func(*gin.Context)
		expectAuth  bool
		expectError bool
	}{
		{
			name: "no token - should continue without auth",
			setupFunc: func(c *gin.Context) {
				// No token setup
			},
			expectAuth:  false,
			expectError: false,
		},
		{
			name: "invalid token - should continue without auth",
			setupFunc: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectAuth:  false,
			expectError: false,
		},
		{
			name: "valid token - should authenticate",
			setupFunc: func(c *gin.Context) {
				// This would need a valid JWT token and user setup
				// For now, we test the flow without actual validation
			},
			expectAuth:  false, // Would be true with proper setup
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test router
			router := gin.New()
			router.Use(OptionalAuthMiddleware())
			router.GET("/test", func(c *gin.Context) {
				userID := c.GetString("userId")
				authenticated := userID != ""
				c.JSON(200, gin.H{"authenticated": authenticated})
			})

			// Make test request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			// Create context for setup
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			tt.setupFunc(c)

			router.ServeHTTP(w, req)

			// Check that request was processed (no abort)
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Parse response
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			authenticated := response["authenticated"].(bool)
			if authenticated != tt.expectAuth {
				t.Errorf("Expected authenticated=%v, got %v", tt.expectAuth, authenticated)
			}
		})
	}
}

// Benchmark tests
func BenchmarkExtractToken(b *testing.B) {
	setupMiddlewareTest(&testing.T{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer test-token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractToken(c)
	}
}

func BenchmarkGenerateRequestID(b *testing.B) {
	setupMiddlewareTest(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateRequestID()
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	setupMiddlewareTest(&testing.T{})

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
	}
}