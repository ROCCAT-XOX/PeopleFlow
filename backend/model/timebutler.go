// backend/model/timebutler.go
package model

import (
	"time"
)

// TimebutlerUser repräsentiert einen Benutzer aus der Timebutler-API
type TimebutlerUser struct {
	UserID                      string    `json:"userId"`
	LastName                    string    `json:"lastName"`
	FirstName                   string    `json:"firstName"`
	EmployeeNumber              string    `json:"employeeNumber"`
	EmailAddress                string    `json:"emailAddress"`
	Phone                       string    `json:"phone"`
	MobilePhone                 string    `json:"mobilePhone"`
	CostCenter                  string    `json:"costCenter"`
	BranchOffice                string    `json:"branchOffice"`
	Department                  string    `json:"department"`
	UserType                    string    `json:"userType"`
	Language                    string    `json:"language"`
	ManagerIDs                  []string  `json:"managerIds"`
	UserAccountLocked           bool      `json:"userAccountLocked"`
	AdditionalInformation       string    `json:"additionalInformation"`
	DateOfEntry                 time.Time `json:"dateOfEntry"`
	DateOfSeparationFromCompany time.Time `json:"dateOfSeparationFromCompany"`
	DateOfBirth                 time.Time `json:"dateOfBirth"`
}

// TimebutlerAbsence repräsentiert eine Abwesenheit aus der Timebutler-API
type TimebutlerAbsence struct {
	UserID       string    `json:"userId"`
	EmailAddress string    `json:"emailAddress"`
	StartDate    time.Time `json:"startDate"`
	EndDate      time.Time `json:"endDate"`
	AbsenceType  string    `json:"absenceType"`
	Status       string    `json:"status"`
	Comment      string    `json:"comment"`
}
