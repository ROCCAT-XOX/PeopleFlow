package model

import (
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
	RoleAdmin   UserRole = "admin"
	RoleHR      UserRole = "hr"
	RoleUser    UserRole = "user"
	RoleManager UserRole = "manager"

	// Benutzerstatus
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
)

// User repräsentiert einen Benutzer im System
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName string             `bson:"firstName" json:"firstName"`
	LastName  string             `bson:"lastName" json:"lastName"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"` // "-" verhindert, dass das Passwort in JSON-Antworten erscheint
	Role      UserRole           `bson:"role" json:"role"`
	Status    UserStatus         `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// HashPassword verschlüsselt das Passwort mit bcrypt
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword überprüft, ob das eingegebene Passwort mit dem gespeicherten Hash übereinstimmt
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
