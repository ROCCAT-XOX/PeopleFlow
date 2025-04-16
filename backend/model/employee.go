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
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName       string             `bson:"firstName" json:"firstName"`
	LastName        string             `bson:"lastName" json:"lastName"`
	Email           string             `bson:"email" json:"email"`
	Phone           string             `bson:"phone" json:"phone"`
	Address         string             `bson:"address" json:"address"`
	DateOfBirth     time.Time          `bson:"dateOfBirth" json:"dateOfBirth"`
	HireDate        time.Time          `bson:"hireDate" json:"hireDate"`
	Position        string             `bson:"position" json:"position"`
	Department      Department         `bson:"department" json:"department"`
	ManagerID       primitive.ObjectID `bson:"managerId,omitempty" json:"managerId"`
	Status          EmployeeStatus     `bson:"status" json:"status"`
	Salary          float64            `bson:"salary" json:"salary"`
	BankAccount     string             `bson:"bankAccount" json:"bankAccount"`
	TaxID           string             `bson:"taxId" json:"taxId"`
	SocialSecID     string             `bson:"socialSecId" json:"socialSecId"`
	HealthInsurance string             `bson:"healthInsurance" json:"healthInsurance"`
	EmergencyName   string             `bson:"emergencyName" json:"emergencyName"`
	EmergencyPhone  string             `bson:"emergencyPhone" json:"emergencyPhone"`
	ProfileImage    string             `bson:"profileImage" json:"profileImage"`
	Notes           string             `bson:"notes" json:"notes"`

	// Neue Felder für erweiterte Mitarbeiterinformationen

	// Bewerbungs- und Einstellungsunterlagen
	ApplicationDocuments []Document `bson:"applicationDocuments,omitempty" json:"applicationDocuments,omitempty"`

	// Weiterbildung und Entwicklung
	Trainings       []Training        `bson:"trainings,omitempty" json:"trainings,omitempty"`
	DevelopmentPlan []DevelopmentItem `bson:"developmentPlan,omitempty" json:"developmentPlan,omitempty"`
	Evaluations     []Evaluation      `bson:"evaluations,omitempty" json:"evaluations,omitempty"`

	// Abwesenheiten und Urlaub
	Absences          []Absence `bson:"absences,omitempty" json:"absences,omitempty"`
	VacationDays      int       `bson:"vacationDays" json:"vacationDays"`
	RemainingVacation int       `bson:"remainingVacation" json:"remainingVacation"`

	// Allgemeine Dokumente
	Documents []Document `bson:"documents,omitempty" json:"documents,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Document repräsentiert ein Dokument oder eine Datei im System
type Document struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	FileName    string             `bson:"fileName" json:"fileName"`
	FileType    string             `bson:"fileType" json:"fileType"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	FilePath    string             `bson:"filePath" json:"filePath"`
	FileSize    int64              `bson:"fileSize" json:"fileSize"`
	UploadDate  time.Time          `bson:"uploadDate" json:"uploadDate"`
	UploadedBy  primitive.ObjectID `bson:"uploadedBy,omitempty" json:"uploadedBy"`
}

// Training repräsentiert eine Weiterbildung oder ein Training
type Training struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	StartDate   time.Time          `bson:"startDate" json:"startDate"`
	EndDate     time.Time          `bson:"endDate" json:"endDate"`
	Provider    string             `bson:"provider" json:"provider"`
	Certificate string             `bson:"certificate" json:"certificate"`
	Status      string             `bson:"status" json:"status"` // planned, ongoing, completed
	Notes       string             `bson:"notes" json:"notes"`
	Documents   []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
}

// DevelopmentItem repräsentiert einen Eintrag im Entwicklungsplan eines Mitarbeiters
type DevelopmentItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Type        string             `bson:"type" json:"type"` // skill, knowledge, certification
	TargetDate  time.Time          `bson:"targetDate" json:"targetDate"`
	Status      string             `bson:"status" json:"status"` // not started, in progress, completed
	Notes       string             `bson:"notes" json:"notes"`
}

// Evaluation repräsentiert eine Leistungsbeurteilung
type Evaluation struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title            string             `bson:"title" json:"title"`
	Date             time.Time          `bson:"date" json:"date"`
	EvaluatorID      primitive.ObjectID `bson:"evaluatorId,omitempty" json:"evaluatorId"`
	EvaluatorName    string             `bson:"evaluatorName" json:"evaluatorName"`
	OverallRating    int                `bson:"overallRating" json:"overallRating"` // 1-5
	Strengths        string             `bson:"strengths" json:"strengths"`
	AreasToImprove   string             `bson:"areasToImprove" json:"areasToImprove"`
	Comments         string             `bson:"comments" json:"comments"`
	EmployeeComments string             `bson:"employeeComments" json:"employeeComments"`
	Documents        []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
}

// Absence repräsentiert eine Abwesenheit (Urlaub, Krankheit, etc.)
type Absence struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type         string             `bson:"type" json:"type"` // vacation, sick, special
	StartDate    time.Time          `bson:"startDate" json:"startDate"`
	EndDate      time.Time          `bson:"endDate" json:"endDate"`
	Days         float64            `bson:"days" json:"days"`
	Status       string             `bson:"status" json:"status"` // requested, approved, rejected, cancelled
	ApprovedBy   primitive.ObjectID `bson:"approvedBy,omitempty" json:"approvedBy"`
	ApproverName string             `bson:"approverName" json:"approverName"`
	Reason       string             `bson:"reason" json:"reason"`
	Notes        string             `bson:"notes" json:"notes"`
	Documents    []Document         `bson:"documents,omitempty" json:"documents,omitempty"`
}
