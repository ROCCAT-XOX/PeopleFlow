package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EmployeeStatus repräsentiert den Status eines Mitarbeiters
type EmployeeStatus string

// Department repräsentiert eine Abteilung
type Department string

const (
	// Mitarbeiterstatus
	EmployeeStatusActive   EmployeeStatus = "active"
	EmployeeStatusInactive EmployeeStatus = "inactive"
	EmployeeStatusOnLeave  EmployeeStatus = "onleave"
	EmployeeStatusRemote   EmployeeStatus = "remote"

	// Abteilungen
	DepartmentIT         Department = "IT"
	DepartmentSales      Department = "Sales"
	DepartmentHR         Department = "HR"
	DepartmentMarketing  Department = "Marketing"
	DepartmentFinance    Department = "Finance"
	DepartmentProduction Department = "Production"
)

// Employee repräsentiert einen Mitarbeiter im System
type Employee struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName      string             `bson:"firstName" json:"firstName"`
	LastName       string             `bson:"lastName" json:"lastName"`
	Email          string             `bson:"email" json:"email"`
	Phone          string             `bson:"phone" json:"phone"`
	Address        string             `bson:"address" json:"address"`
	DateOfBirth    time.Time          `bson:"dateOfBirth" json:"dateOfBirth"`
	HireDate       time.Time          `bson:"hireDate" json:"hireDate"`
	Position       string             `bson:"position" json:"position"`
	Department     Department         `bson:"department" json:"department"`
	ManagerID      primitive.ObjectID `bson:"managerId,omitempty" json:"managerId"`
	Status         EmployeeStatus     `bson:"status" json:"status"`
	Salary         float64            `bson:"salary" json:"salary"`
	BankAccount    string             `bson:"bankAccount" json:"bankAccount"`
	TaxID          string             `bson:"taxId" json:"taxId"`
	SocialSecID    string             `bson:"socialSecId" json:"socialSecId"`
	EmergencyName  string             `bson:"emergencyName" json:"emergencyName"`
	EmergencyPhone string             `bson:"emergencyPhone" json:"emergencyPhone"`
	ProfileImage   string             `bson:"profileImage" json:"profileImage"`
	Notes          string             `bson:"notes" json:"notes"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}
