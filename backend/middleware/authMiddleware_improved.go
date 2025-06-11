package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuthMiddleware errors
var (
	ErrNoToken           = errors.New("no authentication token provided")
	ErrInvalidToken      = errors.New("invalid authentication token")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserInactive      = errors.New("user account is inactive")
	ErrTokenExpired      = errors.New("authentication token has expired")
	ErrInsufficientRole  = errors.New("insufficient role permissions")
)

// ImprovedAuthMiddleware provides enhanced authentication with comprehensive logging
func ImprovedAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := context.WithValue(c.Request.Context(), "requestID", generateRequestID())
		c.Request = c.Request.WithContext(ctx)

		// Start performance monitoring
		perf := utils.StartPerformanceLogging(ctx, "AuthMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "AuthMiddleware", c.Writer.Status() < 400, time.Since(start),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"user_agent", c.GetHeader("User-Agent"),
				"ip", c.ClientIP(),
			)
			perf.End("status", c.Writer.Status())
		}()

		// Extract token
		tokenString, err := improvedExtractToken(c)
		if err != nil {
			utils.LogWarn(ctx, "Authentication failed: no token",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			handleAuthError(c, ErrNoToken, "Authentication required")
			perf.EndWithError(err)
			return
		}

		// Validate token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			utils.LogWarn(ctx, "Authentication failed: invalid token",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			handleAuthError(c, ErrInvalidToken, "Invalid authentication token")
			perf.EndWithError(err)
			return
		}

		// Check token expiration
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			utils.LogWarn(ctx, "Authentication failed: token expired",
				"user_id", claims.UserID,
				"expires_at", claims.ExpiresAt,
				"path", c.Request.URL.Path,
			)
			handleAuthError(c, ErrTokenExpired, "Authentication token has expired")
			perf.EndWithError(ErrTokenExpired)
			return
		}

		// Get user from database
		userRepo := repository.NewUserRepository()
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			utils.LogError(ctx, err, "Authentication failed: user lookup failed",
				"user_id", claims.UserID,
				"path", c.Request.URL.Path,
			)
			handleAuthError(c, ErrUserNotFound, "User account not found")
			perf.EndWithError(err)
			return
		}

		// Check if user is active
		if user.Status != model.StatusActive {
			utils.LogWarn(ctx, "Authentication failed: user inactive",
				"user_id", user.ID.Hex(),
				"email", user.Email,
				"status", user.Status,
				"path", c.Request.URL.Path,
			)
			handleAuthError(c, ErrUserInactive, "User account is inactive")
			perf.EndWithError(ErrUserInactive)
			return
		}

		// Update last activity could be implemented here
		// go func() {
		//     userRepo := repository.NewUserRepository()
		//     userRepo.UpdateLastLogin(user.ID.Hex())
		// }()

		// Add user information to context
		c.Set("user", user)
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("requestID", ctx.Value("requestID"))

		// Add user info to request context for logging
		userCtx := context.WithValue(ctx, "userID", user.ID.Hex())
		c.Request = c.Request.WithContext(userCtx)

		utils.LogDebug(ctx, "Authentication successful",
			"user_id", user.ID.Hex(),
			"email", user.Email,
			"role", user.Role,
			"path", c.Request.URL.Path,
		)

		c.Next()
	}
}

// APIAuthMiddleware provides JSON-based authentication for API endpoints
func APIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := context.WithValue(c.Request.Context(), "requestID", generateRequestID())
		c.Request = c.Request.WithContext(ctx)

		perf := utils.StartPerformanceLogging(ctx, "APIAuthMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "APIAuthMiddleware", c.Writer.Status() < 400, time.Since(start),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
			)
			perf.End("status", c.Writer.Status())
		}()

		// Extract token from Authorization header
		bearerToken := c.GetHeader("Authorization")
		if bearerToken == "" || !strings.HasPrefix(bearerToken, "Bearer ") {
			utils.LogWarn(ctx, "API authentication failed: missing bearer token",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header with Bearer token required",
				"code":  "MISSING_TOKEN",
			})
			c.Abort()
			perf.EndWithError(ErrNoToken)
			return
		}

		tokenString := strings.TrimPrefix(bearerToken, "Bearer ")

		// Validate token
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			utils.LogWarn(ctx, "API authentication failed: invalid token",
				"error", err.Error(),
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authentication token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			perf.EndWithError(err)
			return
		}

		// Get user
		userRepo := repository.NewUserRepository()
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil {
			utils.LogError(ctx, err, "API authentication failed: user lookup failed",
				"user_id", claims.UserID,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User account not found",
				"code":  "USER_NOT_FOUND",
			})
			c.Abort()
			perf.EndWithError(err)
			return
		}

		// Check if user is active
		if user.Status != model.StatusActive {
			utils.LogWarn(ctx, "API authentication failed: user inactive",
				"user_id", user.ID.Hex(),
				"status", user.Status,
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User account is inactive",
				"code":  "USER_INACTIVE",
			})
			c.Abort()
			perf.EndWithError(ErrUserInactive)
			return
		}

		// Add user information to context
		c.Set("user", user)
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("requestID", ctx.Value("requestID"))

		utils.LogDebug(ctx, "API authentication successful",
			"user_id", user.ID.Hex(),
			"email", user.Email,
			"role", user.Role,
		)

		c.Next()
	}
}

// ImprovedRoleMiddlewareFromAuth creates middleware that requires specific roles
func ImprovedRoleMiddlewareFromAuth(requiredRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		perf := utils.StartPerformanceLogging(ctx, "RoleMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "RoleMiddleware", c.Writer.Status() < 400, time.Since(start),
				"required_roles", requiredRoles,
				"status", c.Writer.Status(),
			)
			perf.End("required_roles", requiredRoles, "status", c.Writer.Status())
		}()

		// Get user role from context
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			utils.LogError(ctx, errors.New("user role not found in context"), "Role check failed",
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: role information missing",
				"code":  "ROLE_MISSING",
			})
			c.Abort()
			perf.EndWithError(errors.New("role missing"))
			return
		}

		userRole := model.UserRole(userRoleInterface.(string))

		// Check if user has any of the required roles
		hasPermission := false
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			user, _ := c.Get("user")
			userModel := user.(*model.User)

			utils.LogWarn(ctx, "Access denied: insufficient role",
				"user_id", userModel.ID.Hex(),
				"user_role", userRole,
				"required_roles", requiredRoles,
				"path", c.Request.URL.Path,
			)

			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Access denied: requires one of roles %v", requiredRoles),
				"code":  "INSUFFICIENT_ROLE",
			})
			c.Abort()
			perf.EndWithError(ErrInsufficientRole)
			return
		}

		utils.LogDebug(ctx, "Role check passed",
			"user_role", userRole,
			"required_roles", requiredRoles,
		)

		c.Next()
	}
}

// ImprovedAdminMiddleware requires admin role
func ImprovedAdminMiddleware() gin.HandlerFunc {
	return ImprovedRoleMiddlewareFromAuth(model.RoleAdmin)
}

// OptionalAuthMiddleware provides optional authentication (doesn't redirect if no token)
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := context.WithValue(c.Request.Context(), "requestID", generateRequestID())
		c.Request = c.Request.WithContext(ctx)

		perf := utils.StartPerformanceLogging(ctx, "OptionalAuthMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "OptionalAuthMiddleware", true, time.Since(start),
				"authenticated", c.GetString("userId") != "",
			)
			perf.End("authenticated", c.GetString("userId") != "")
		}()

		// Try to extract token
		tokenString, err := improvedExtractToken(c)
		if err != nil {
			// No token is okay for optional auth
			c.Set("requestID", ctx.Value("requestID"))
			c.Next()
			return
		}

		// Validate token if present
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			// Invalid token - continue without auth
			utils.LogDebug(ctx, "Optional auth: invalid token ignored",
				"error", err.Error(),
			)
			c.Set("requestID", ctx.Value("requestID"))
			c.Next()
			return
		}

		// Get user if token is valid
		userRepo := repository.NewUserRepository()
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil || user.Status != model.StatusActive {
			// User issues - continue without auth
			utils.LogDebug(ctx, "Optional auth: user issues ignored",
				"user_id", claims.UserID,
				"error", err,
			)
			c.Set("requestID", ctx.Value("requestID"))
			c.Next()
			return
		}

		// Set user info if everything is valid
		c.Set("user", user)
		c.Set("userId", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("requestID", ctx.Value("requestID"))

		utils.LogDebug(ctx, "Optional auth: user authenticated",
			"user_id", user.ID.Hex(),
		)

		c.Next()
	}
}

// improvedExtractToken extracts JWT token from cookie or header with enhanced error reporting
func improvedExtractToken(c *gin.Context) (string, error) {
	// First try cookie
	token, err := c.Cookie("token")
	if err == nil && token != "" {
		return token, nil
	}

	// Then try Authorization header
	bearerToken := c.GetHeader("Authorization")
	if bearerToken != "" && strings.HasPrefix(bearerToken, "Bearer ") {
		return strings.TrimPrefix(bearerToken, "Bearer "), nil
	}

	// Try X-Auth-Token header (alternative)
	authToken := c.GetHeader("X-Auth-Token")
	if authToken != "" {
		return authToken, nil
	}

	return "", ErrNoToken
}

// handleAuthError handles authentication errors consistently
func handleAuthError(c *gin.Context, err error, message string) {
	// Check if this is an API request (JSON expected)
	if isAPIRequest(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": message,
			"code":  getErrorCode(err),
		})
	} else {
		// Redirect to login for web requests
		c.Redirect(http.StatusFound, "/login?error=auth_required")
	}
	c.Abort()
}

// isAPIRequest determines if the request expects JSON response
func isAPIRequest(c *gin.Context) bool {
	// Check Accept header
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}

	// Check Content-Type
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// Check if path starts with /api/
	if strings.HasPrefix(c.Request.URL.Path, "/api/") {
		return true
	}

	return false
}

// getErrorCode maps errors to error codes
func getErrorCode(err error) string {
	switch err {
	case ErrNoToken:
		return "NO_TOKEN"
	case ErrInvalidToken:
		return "INVALID_TOKEN"
	case ErrUserNotFound:
		return "USER_NOT_FOUND"
	case ErrUserInactive:
		return "USER_INACTIVE"
	case ErrTokenExpired:
		return "TOKEN_EXPIRED"
	case ErrInsufficientRole:
		return "INSUFFICIENT_ROLE"
	default:
		return "AUTH_ERROR"
	}
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	return primitive.NewObjectID().Hex()
}

// RequestIDMiddleware adds request ID to all requests
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		ctx := context.WithValue(c.Request.Context(), "requestID", requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Build log entry
		ctx := c.Request.Context()
		if raw != "" {
			path = path + "?" + raw
		}

		utils.LogHTTPRequest(ctx, c.Request.Method, path, c.Writer.Status(), duration,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"referer", c.Request.Referer(),
			"size", c.Writer.Size(),
		)
	}
}

// RecoveryMiddleware provides panic recovery with logging
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		ctx := c.Request.Context()
		
		utils.LogError(ctx, fmt.Errorf("panic recovered: %v", recovered), "Request panic recovered",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
		)

		if isAPIRequest(c) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		} else {
			// For web requests, try to render HTML or fallback to JSON if templates not available
			if c.Value("gin.template") != nil {
				c.HTML(http.StatusInternalServerError, "error.html", gin.H{
					"title":   "Server Error",
					"message": "An internal server error occurred",
					"year":    time.Now().Year(),
				})
			} else {
				// Fallback to JSON if templates not available (e.g., in tests)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
					"code":  "INTERNAL_ERROR",
				})
			}
		}
	})
}