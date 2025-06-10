package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"PeopleFlow/backend/model"
	"PeopleFlow/backend/repository"
	"PeopleFlow/backend/utils"

	"github.com/gin-gonic/gin"
)

// Role middleware errors
var (
	ErrRoleMissing        = errors.New("user role not found in context")
	ErrUserIDMissing      = errors.New("user ID not found in context")
	ErrInvalidTargetUser  = errors.New("invalid target user")
	ErrUnauthorizedAccess = errors.New("unauthorized access to resource")
)

// ImprovedRoleMiddleware provides enhanced role-based access control with logging
func ImprovedRoleMiddleware(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		perf := utils.StartPerformanceLogging(ctx, "RoleMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "RoleMiddleware", c.Writer.Status() < 400, time.Since(start),
				"allowed_roles", allowedRoles,
				"status", c.Writer.Status(),
				"path", c.Request.URL.Path,
			)
			perf.End("allowed_roles", allowedRoles, "status", c.Writer.Status())
		}()

		// Get user role from context
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			utils.LogError(ctx, ErrRoleMissing, "Role check failed: no role in context",
				"path", c.Request.URL.Path,
			)
			handleRoleError(c, ErrRoleMissing, "Access denied: role information missing")
			perf.EndWithError(ErrRoleMissing)
			return
		}

		userRole := model.UserRole(userRoleInterface.(string))

		// Check if user has any of the required roles
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			userID := c.GetString("userId")
			utils.LogWarn(ctx, "Access denied: insufficient role",
				"user_id", userID,
				"user_role", userRole,
				"allowed_roles", allowedRoles,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
			)

			err := fmt.Errorf("access denied: user role %s not in allowed roles %v", userRole, allowedRoles)
			handleRoleError(c, err, fmt.Sprintf("Access denied: requires one of roles %v", allowedRoles))
			perf.EndWithError(err)
			return
		}

		utils.LogDebug(ctx, "Role check passed",
			"user_role", userRole,
			"allowed_roles", allowedRoles,
			"path", c.Request.URL.Path,
		)

		c.Next()
	}
}

// ImprovedHRMiddleware provides enhanced HR role protection with logging
func ImprovedHRMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		perf := utils.StartPerformanceLogging(ctx, "HRMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "HRMiddleware", c.Writer.Status() < 400, time.Since(start),
				"target_id", c.Param("id"),
				"status", c.Writer.Status(),
			)
			perf.End("target_id", c.Param("id"), "status", c.Writer.Status())
		}()

		// Get current user role
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			utils.LogError(ctx, ErrRoleMissing, "HR middleware: no role in context")
			handleRoleError(c, ErrRoleMissing, "Access denied: role information missing")
			perf.EndWithError(ErrRoleMissing)
			return
		}

		userRole := model.UserRole(userRoleInterface.(string))
		userID := c.GetString("userId")

		// Get target user ID from parameter
		targetID := c.Param("id")

		// If no target ID, allow the request to continue
		if targetID == "" {
			c.Next()
			return
		}

		// Only check restrictions for HR role
		if userRole == model.RoleHR {
			userRepo := repository.NewImprovedUserRepository()
			targetUser, err := userRepo.FindByID(targetID)

			if err != nil {
				utils.LogError(ctx, err, "HR middleware: failed to find target user",
					"target_id", targetID,
					"user_id", userID,
				)
				handleRoleError(c, ErrInvalidTargetUser, "Invalid target user")
				perf.EndWithError(err)
				return
			}

			// Check if target user is Admin or Manager
			if targetUser.Role == model.RoleAdmin || targetUser.Role == model.RoleManager {
				utils.LogWarn(ctx, "HR access denied: cannot modify admin/manager",
					"hr_user_id", userID,
					"target_user_id", targetID,
					"target_role", targetUser.Role,
					"path", c.Request.URL.Path,
				)

				err := fmt.Errorf("HR cannot modify user with role %s", targetUser.Role)
				handleRoleError(c, err, "Access denied: cannot modify administrators or managers")
				perf.EndWithError(err)
				return
			}

			utils.LogDebug(ctx, "HR middleware: access granted",
				"hr_user_id", userID,
				"target_user_id", targetID,
				"target_role", targetUser.Role,
			)
		}

		c.Next()
	}
}

// ImprovedSalaryViewMiddleware provides enhanced salary data access control
func ImprovedSalaryViewMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		perf := utils.StartPerformanceLogging(ctx, "SalaryViewMiddleware")
		defer func() {
			hideSalary := c.GetBool("hideSalary")
			utils.LogMiddleware(ctx, "SalaryViewMiddleware", true, time.Since(start),
				"salary_hidden", hideSalary,
			)
			perf.End("salary_hidden", hideSalary)
		}()

		// Get user role from context
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			utils.LogError(ctx, ErrRoleMissing, "Salary middleware: no role in context")
			c.Set("hideSalary", true) // Default to hiding salary
			c.Next()
			return
		}

		userRole := model.UserRole(userRoleInterface.(string))
		userID := c.GetString("userId")

		// Only Admin and Manager can view salary data
		canViewSalary := userRole == model.RoleAdmin || userRole == model.RoleManager

		c.Set("hideSalary", !canViewSalary)

		if canViewSalary {
			utils.LogDebug(ctx, "Salary access granted",
				"user_id", userID,
				"user_role", userRole,
			)
		} else {
			utils.LogDebug(ctx, "Salary access denied",
				"user_id", userID,
				"user_role", userRole,
			)
		}

		c.Next()
	}
}

// ImprovedSelfOrAdminMiddleware provides enhanced self-access control with logging
func ImprovedSelfOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		perf := utils.StartPerformanceLogging(ctx, "SelfOrAdminMiddleware")
		defer func() {
			utils.LogMiddleware(ctx, "SelfOrAdminMiddleware", c.Writer.Status() < 400, time.Since(start),
				"status", c.Writer.Status(),
				"self_access", c.GetBool("self_access"),
				"admin_access", c.GetBool("admin_access"),
			)
			perf.End("status", c.Writer.Status())
		}()

		// Get user ID from context
		userIDInterface, exists := c.Get("userId")
		if !exists {
			utils.LogError(ctx, ErrUserIDMissing, "Self/Admin middleware: no user ID in context")
			handleRoleError(c, ErrUserIDMissing, "Access denied: user information missing")
			perf.EndWithError(ErrUserIDMissing)
			return
		}

		userID := userIDInterface.(string)

		// Get user role from context
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			utils.LogError(ctx, ErrRoleMissing, "Self/Admin middleware: no role in context")
			handleRoleError(c, ErrRoleMissing, "Access denied: role information missing")
			perf.EndWithError(ErrRoleMissing)
			return
		}

		userRole := model.UserRole(userRoleInterface.(string))

		// Get requested ID from parameter or form
		requestedID := c.Param("id")
		if requestedID == "" {
			// Try to get from form data (for POST requests)
			requestedID = c.PostForm("id")
		}
		if requestedID == "" {
			// Try to get from JSON body for API requests
			if jsonData, exists := c.Get("json_data"); exists {
				if data, ok := jsonData.(map[string]interface{}); ok {
					if id, ok := data["id"].(string); ok {
						requestedID = id
					}
				}
			}
		}

		isAdmin := userRole == model.RoleAdmin
		isSelfAccess := userID == requestedID

		// Set access type flags for logging
		c.Set("admin_access", isAdmin)
		c.Set("self_access", isSelfAccess)

		// Check if user has access (admin or accessing own data)
		if isAdmin || isSelfAccess {
			if isAdmin {
				utils.LogDebug(ctx, "Admin access granted",
					"user_id", userID,
					"requested_id", requestedID,
					"path", c.Request.URL.Path,
				)
			} else {
				utils.LogDebug(ctx, "Self access granted",
					"user_id", userID,
					"requested_id", requestedID,
					"path", c.Request.URL.Path,
				)
			}
			c.Next()
			return
		}

		// Access denied
		utils.LogWarn(ctx, "Self/Admin access denied",
			"user_id", userID,
			"user_role", userRole,
			"requested_id", requestedID,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)

		err := fmt.Errorf("access denied: user %s cannot access resource for user %s", userID, requestedID)
		handleRoleError(c, err, "Access denied: you can only access your own data or must be an administrator")
		perf.EndWithError(err)
	}
}

// AdminOnlyMiddleware provides admin-only access control
func AdminOnlyMiddleware() gin.HandlerFunc {
	return ImprovedRoleMiddleware(model.RoleAdmin)
}

// ManagerOrAdminMiddleware provides manager or admin access control
func ManagerOrAdminMiddleware() gin.HandlerFunc {
	return ImprovedRoleMiddleware(model.RoleAdmin, model.RoleManager)
}

// HROrHigherMiddleware provides HR, manager, or admin access control
func HROrHigherMiddleware() gin.HandlerFunc {
	return ImprovedRoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR)
}

// EmployeeOrHigherMiddleware provides access for all authenticated users
func EmployeeOrHigherMiddleware() gin.HandlerFunc {
	return ImprovedRoleMiddleware(model.RoleAdmin, model.RoleManager, model.RoleHR, model.RoleEmployee)
}

// handleRoleError handles role-based errors consistently
func handleRoleError(c *gin.Context, err error, message string) {
	// Check if this is an API request (JSON expected)
	if isAPIRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": message,
			"code":  getRoleErrorCode(err),
		})
	} else {
		// For web requests, try to render HTML or fallback to JSON if templates not available
		if c.Value("gin.template") != nil {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"title":   "Access Denied",
				"message": message,
				"year":    time.Now().Year(),
			})
		} else {
			// Fallback to JSON if templates not available (e.g., in tests)
			c.JSON(http.StatusForbidden, gin.H{
				"error": message,
				"code":  getRoleErrorCode(err),
			})
		}
	}
	c.Abort()
}

// getRoleErrorCode maps role errors to error codes
func getRoleErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrRoleMissing):
		return "ROLE_MISSING"
	case errors.Is(err, ErrUserIDMissing):
		return "USER_ID_MISSING"
	case errors.Is(err, ErrInvalidTargetUser):
		return "INVALID_TARGET_USER"
	case errors.Is(err, ErrUnauthorizedAccess):
		return "UNAUTHORIZED_ACCESS"
	default:
		return "ACCESS_DENIED"
	}
}

// RolePermissionChecker provides utility methods for checking role permissions
type RolePermissionChecker struct{}

// NewRolePermissionChecker creates a new role permission checker
func NewRolePermissionChecker() *RolePermissionChecker {
	return &RolePermissionChecker{}
}

// CanAccessUser checks if a user can access another user's data
func (rpc *RolePermissionChecker) CanAccessUser(userRole model.UserRole, userID, targetID string) bool {
	// Admin can access anyone
	if userRole == model.RoleAdmin {
		return true
	}

	// Users can access their own data
	if userID == targetID {
		return true
	}

	return false
}

// CanModifyUser checks if a user can modify another user's data
func (rpc *RolePermissionChecker) CanModifyUser(userRole model.UserRole, userID, targetID string, targetRole model.UserRole) bool {
	// Admin can modify anyone
	if userRole == model.RoleAdmin {
		return true
	}

	// Manager can modify HR and Employee roles
	if userRole == model.RoleManager && (targetRole == model.RoleHR || targetRole == model.RoleEmployee) {
		return true
	}

	// HR can modify Employee roles (but not Admin/Manager)
	if userRole == model.RoleHR && targetRole == model.RoleEmployee {
		return true
	}

	// Users can modify their own data
	if userID == targetID {
		return true
	}

	return false
}

// CanViewSalary checks if a user can view salary information
func (rpc *RolePermissionChecker) CanViewSalary(userRole model.UserRole) bool {
	return userRole == model.RoleAdmin || userRole == model.RoleManager
}

// HasMinimumRole checks if a user has at least the specified role level
func (rpc *RolePermissionChecker) HasMinimumRole(userRole, minimumRole model.UserRole) bool {
	roleHierarchy := map[model.UserRole]int{
		model.RoleEmployee: 1,
		model.RoleHR:       2,
		model.RoleManager:  3,
		model.RoleAdmin:    4,
	}

	userLevel := roleHierarchy[userRole]
	minLevel := roleHierarchy[minimumRole]

	return userLevel >= minLevel
}