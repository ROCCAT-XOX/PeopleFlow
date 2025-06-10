package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserRole definiert die Rolle eines Benutzers
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleEmployee UserRole = "employee"
)

// UserStatus definiert den Status eines Benutzers
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

// User repräsentiert einen Systembenutzer
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	Password     string             `bson:"-" json:"password,omitempty"` // Never stored, only used for input
	PasswordHash string             `bson:"passwordHash" json:"-"`       // Never exposed in JSON
	FirstName    string             `bson:"firstName" json:"firstName"`
	LastName     string             `bson:"lastName" json:"lastName"`
	Role         UserRole           `bson:"role" json:"role"`
	Status       UserStatus         `bson:"status" json:"status"`
	EmployeeID   primitive.ObjectID `bson:"employeeId,omitempty" json:"employeeId,omitempty"`
	LastLogin    *time.Time         `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	DeletedAt    *time.Time         `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	if u.Password == "" {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hashedPassword)
	u.Password = "" // Clear the plain text password
	return nil
}

// CheckPassword checks if the provided password matches the user's password hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// FullName returns the full name of the user
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// IsActive returns whether the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive && u.DeletedAt == nil
}

// IsAdmin returns whether the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsEmployee returns whether the user has employee role
func (u *User) IsEmployee() bool {
	return u.Role == RoleEmployee
}

// CanManageEmployees returns whether the user can manage employees
func (u *User) CanManageEmployees() bool {
	return u.IsAdmin() && u.IsActive()
}

// CanApproveRequests returns whether the user can approve requests
func (u *User) CanApproveRequests() bool {
	return u.IsAdmin() && u.IsActive()
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
	u.UpdatedAt = now
}
