package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/utils"
)

// MockUserRepository is a mock implementation of the UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByPasswordResetToken(token string) (*model.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]string
		setupMock      func(*MockUserRepository)
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "Successful login",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:        primitive.NewObjectID(),
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Role:      model.RoleUser,
					Active:    true,
				}
				// Set the password hash
				user.SetPassword("password123")
				m.On("FindByEmail", "test@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp map[string]interface{}) {
				assert.Contains(t, resp, "token")
				assert.Contains(t, resp, "user")
				user := resp["user"].(map[string]interface{})
				assert.Equal(t, "test@example.com", user["email"])
			},
		},
		{
			name: "Invalid email format",
			payload: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
			setupMock:      func(m *MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User not found",
			payload: map[string]string{
				"email":    "notfound@example.com",
				"password": "password123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", "notfound@example.com").Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Wrong password",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:        primitive.NewObjectID(),
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Role:      model.RoleUser,
					Active:    true,
				}
				user.SetPassword("password123")
				m.On("FindByEmail", "test@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Inactive user",
			payload: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:        primitive.NewObjectID(),
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Role:      model.RoleUser,
					Active:    false,
				}
				user.SetPassword("password123")
				m.On("FindByEmail", "test@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			router := setupTestRouter()
			authHandler := &AuthHandler{
				userRepo: mockRepo,
			}
			router.POST("/login", authHandler.Login)

			// Make request
			jsonBody, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.checkResponse != nil && w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				tt.checkResponse(t, response)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPasswordResetRequest(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]string
		setupMock      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name: "Successful password reset request",
			payload: map[string]string{
				"email": "test@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:        primitive.NewObjectID(),
					Email:     "test@example.com",
					FirstName: "Test",
					LastName:  "User",
					Active:    true,
				}
				m.On("FindByEmail", "test@example.com").Return(user, nil)
				m.On("Update", mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User not found - still returns OK for security",
			payload: map[string]string{
				"email": "notfound@example.com",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindByEmail", "notfound@example.com").Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid email format",
			payload: map[string]string{
				"email": "invalid-email",
			},
			setupMock:      func(m *MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			router := setupTestRouter()
			authHandler := &AuthHandler{
				userRepo: mockRepo,
			}
			router.POST("/password-reset-request", authHandler.PasswordResetRequest)

			// Make request
			jsonBody, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/password-reset-request", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPasswordReset(t *testing.T) {
	validToken := utils.GenerateResetToken()
	
	tests := []struct {
		name           string
		payload        map[string]string
		setupMock      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name: "Successful password reset",
			payload: map[string]string{
				"token":    validToken,
				"password": "newpassword123",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:                   primitive.NewObjectID(),
					Email:                "test@example.com",
					PasswordResetToken:   validToken,
					PasswordResetExpires: primitive.NewDateTimeFromTime(time.Now().Add(time.Hour)),
					Active:               true,
				}
				m.On("FindByPasswordResetToken", validToken).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid token",
			payload: map[string]string{
				"token":    "invalid-token",
				"password": "newpassword123",
			},
			setupMock: func(m *MockUserRepository) {
				m.On("FindByPasswordResetToken", "invalid-token").Return(nil, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Expired token",
			payload: map[string]string{
				"token":    validToken,
				"password": "newpassword123",
			},
			setupMock: func(m *MockUserRepository) {
				user := &model.User{
					ID:                   primitive.NewObjectID(),
					Email:                "test@example.com",
					PasswordResetToken:   validToken,
					PasswordResetExpires: primitive.NewDateTimeFromTime(time.Now().Add(-time.Hour)),
					Active:               true,
				}
				m.On("FindByPasswordResetToken", validToken).Return(user, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Weak password",
			payload: map[string]string{
				"token":    validToken,
				"password": "123",
			},
			setupMock:      func(m *MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)
			
			router := setupTestRouter()
			authHandler := &AuthHandler{
				userRepo: mockRepo,
			}
			router.POST("/password-reset", authHandler.PasswordReset)

			// Make request
			jsonBody, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/password-reset", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}