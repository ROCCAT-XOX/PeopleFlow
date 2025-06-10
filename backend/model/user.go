package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserRole repräsentiert die Rolle eines Benutzers
type UserRole string

// UserStatus repräsentiert den Status eines Benutzers
type UserStatus string

const (
	// Benutzerrollen
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleHR       UserRole = "hr"       // Personalverwaltung
	RoleEmployee UserRole = "employee" // Standard Mitarbeiter
	RoleUser     UserRole = "user"     // Legacy - wird zu RoleEmployee migriert

	// Benutzerstatus
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

// User validation errors
var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrWeakPassword     = errors.New("password is too weak")
	ErrEmptyName        = errors.New("name cannot be empty")
	ErrInvalidRole      = errors.New("invalid user role")
	ErrInvalidStatus    = errors.New("invalid user status")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
)

// User repräsentiert einen Benutzer im System
type User struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	FirstName    string              `bson:"firstName" json:"firstName"`
	LastName     string              `bson:"lastName" json:"lastName"`
	Email        string              `bson:"email" json:"email"`
	Password     string              `bson:"password,omitempty" json:"-"`     // Legacy field for backward compatibility + input
	PasswordHash string              `bson:"passwordHash,omitempty" json:"-"` // New field for password hashes
	Role         UserRole            `bson:"role" json:"role"`
	Status       UserStatus          `bson:"status" json:"status"`
	EmployeeID   *primitive.ObjectID `bson:"employeeId,omitempty" json:"employeeId,omitempty"` // Link to Employee
	LastLogin    *time.Time          `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
	DeletedAt    *time.Time          `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	CreatedAt    time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time           `bson:"updatedAt" json:"updatedAt"`
}

// Validate validates all user fields
func (u *User) Validate(isUpdate bool) error {
	// Basic field validation for new users
	if !isUpdate {
		if err := u.ValidateName(); err != nil {
			return err
		}
		if err := u.ValidateEmail(); err != nil {
			return err
		}
		if err := u.ValidatePassword(); err != nil {
			return err
		}
	}

	// Always validate role and status if provided
	if u.Role != "" {
		if err := u.ValidateRole(); err != nil {
			return err
		}
	}
	if u.Status != "" {
		if err := u.ValidateStatus(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateName validates first and last name
func (u *User) ValidateName() error {
	if strings.TrimSpace(u.FirstName) == "" {
		return fmt.Errorf("%w: first name", ErrEmptyName)
	}
	if strings.TrimSpace(u.LastName) == "" {
		return fmt.Errorf("%w: last name", ErrEmptyName)
	}
	return nil
}

// ValidateEmail validates email format
func (u *User) ValidateEmail() error {
	if strings.TrimSpace(u.Email) == "" {
		return fmt.Errorf("%w: email cannot be empty", ErrInvalidEmail)
	}
	
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword validates password strength
func (u *User) ValidatePassword() error {
	if u.Password == "" {
		return ErrPasswordTooShort
	}
	if len(u.Password) < 6 {
		return ErrPasswordTooShort
	}
	// Add more password strength requirements as needed
	return nil
}

// ValidateRole validates user role
func (u *User) ValidateRole() error {
	switch u.Role {
	case RoleAdmin, RoleManager, RoleHR, RoleEmployee, RoleUser:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidRole, u.Role)
	}
}

// ValidateStatus validates user status
func (u *User) ValidateStatus() error {
	switch u.Status {
	case StatusActive, StatusInactive:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidStatus, u.Status)
	}
}

// HashPassword verschlüsselt das Passwort mit bcrypt
func (u *User) HashPassword() error {
	if u.Password == "" {
		return ErrPasswordTooShort
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	u.PasswordHash = string(hashedPassword)
	u.Password = "" // Clear plain password
	return nil
}

// CheckPassword überprüft, ob das eingegebene Passwort mit dem gespeicherten Hash übereinstimmt
func (u *User) CheckPassword(password string) bool {
	if password == "" {
		return false
	}
	
	// Try new PasswordHash field first
	if u.PasswordHash != "" {
		err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
		return err == nil
	}
	
	// Fallback to old Password field for backward compatibility
	if u.Password != "" {
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
		return err == nil
	}
	
	return false
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

// GetDisplayName returns a display name for the user
func (u *User) GetDisplayName() string {
	fullName := u.GetFullName()
	if fullName != "" {
		return fullName
	}
	return u.Email
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsManager returns true if the user has manager role
func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

// IsHR returns true if the user has HR role
func (u *User) IsHR() bool {
	return u.Role == RoleHR
}

// IsEmployee returns true if the user has employee role
func (u *User) IsEmployee() bool {
	return u.Role == RoleEmployee || u.Role == RoleUser // Support legacy role
}

// HasRole checks if the user has any of the specified roles
func (u *User) HasRole(roles ...UserRole) bool {
	for _, role := range roles {
		if u.Role == role {
			return true
		}
	}
	return false
}

// CanModifyUser checks if this user can modify another user
func (u *User) CanModifyUser(targetUser *User) bool {
	// Admin can modify anyone
	if u.IsAdmin() {
		return true
	}
	
	// Manager can modify HR and Employee roles
	if u.IsManager() && (targetUser.IsHR() || targetUser.IsEmployee()) {
		return true
	}
	
	// HR can modify Employee roles (but not Admin/Manager)
	if u.IsHR() && targetUser.IsEmployee() {
		return true
	}
	
	// Users can only modify themselves
	return u.ID == targetUser.ID
}

// CanViewSalary checks if this user can view salary information
func (u *User) CanViewSalary() bool {
	return u.IsAdmin() || u.IsManager()
}

// NormalizeEmail normalizes the email address
func (u *User) NormalizeEmail() {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
}

// SetDefaults sets default values for the user
func (u *User) SetDefaults() {
	if u.Status == "" {
		u.Status = StatusActive
	}
	if u.Role == "" {
		u.Role = RoleEmployee
	}
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	u.UpdatedAt = now
}

// PrepareForCreate prepares the user for creation
func (u *User) PrepareForCreate() error {
	u.NormalizeEmail()
	u.SetDefaults()
	
	if err := u.Validate(false); err != nil {
		return err
	}
	
	if err := u.HashPassword(); err != nil {
		return err
	}
	
	return nil
}

// PrepareForUpdate prepares the user for update
func (u *User) PrepareForUpdate() error {
	u.NormalizeEmail()
	u.UpdatedAt = time.Now()
	
	if err := u.Validate(true); err != nil {
		return err
	}
	
	// Only hash password if it's provided
	if u.Password != "" {
		if err := u.HashPassword(); err != nil {
			return err
		}
	}
	
	return nil
}

// GetRoleDisplayName returns the German display name for the role
func (u *User) GetRoleDisplayName() string {
	switch u.Role {
	case RoleAdmin:
		return "Administrator"
	case RoleManager:
		return "Manager"
	case RoleHR:
		return "Personalverwaltung"
	case RoleEmployee, RoleUser:
		return "Mitarbeiter"
	default:
		return string(u.Role)
	}
}

// GetStatusDisplayName returns the German display name for the status
func (u *User) GetStatusDisplayName() string {
	switch u.Status {
	case StatusActive:
		return "Aktiv"
	case StatusInactive:
		return "Inaktiv"
	default:
		return string(u.Status)
	}
}

// ToJSON returns a JSON-safe representation of the user (without password)
func (u *User) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"id":        u.ID,
		"firstName": u.FirstName,
		"lastName":  u.LastName,
		"email":     u.Email,
		"role":      u.Role,
		"status":    u.Status,
		"fullName":  u.GetFullName(),
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
	}
	
	if u.EmployeeID != nil {
		result["employeeId"] = u.EmployeeID
	}
	
	if u.LastLogin != nil {
		result["lastLogin"] = u.LastLogin
	}
	
	return result
}
