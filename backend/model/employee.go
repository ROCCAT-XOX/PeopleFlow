package model

import (
	"fmt"
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

	// Worktime
	WorkTimeModelFullTime   WorkTimeModel = "fulltime"   // Vollzeit
	WorkTimeModelPartTime   WorkTimeModel = "parttime"   // Teilzeit
	WorkTimeModelFlexTime   WorkTimeModel = "flextime"   // Gleitzeit
	WorkTimeModelRemote     WorkTimeModel = "remote"     // Remote/Homeoffice
	WorkTimeModelShift      WorkTimeModel = "shift"      // Schichtarbeit
	WorkTimeModelContract   WorkTimeModel = "contract"   // Werkvertrag
	WorkTimeModelInternship WorkTimeModel = "internship" // Praktikum
)

// Employee repräsentiert einen Mitarbeiter im System
// Employee repräsentiert einen Mitarbeiter im System
type Employee struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EmployeeID        string             `bson:"employeeId" json:"employeeId"`
	FirstName         string             `bson:"firstName" json:"firstName"`
	LastName          string             `bson:"lastName" json:"lastName"`
	Email             string             `bson:"email" json:"email"`
	Phone             string             `bson:"phone" json:"phone"`
	InternalPhone     string             `bson:"internalPhone" json:"internalPhone"`
	InternalExtension string             `bson:"internalExtension" json:"internalExtension"`
	Address           string             `bson:"address" json:"address"`
	DateOfBirth       time.Time          `bson:"dateOfBirth" json:"dateOfBirth"`

	// Beschäftigungsdaten
	HireDate   time.Time          `bson:"hireDate" json:"hireDate"`
	Position   string             `bson:"position" json:"position"`
	Department Department         `bson:"department" json:"department"`
	ManagerID  primitive.ObjectID `bson:"managerId,omitempty" json:"managerId"`
	Status     EmployeeStatus     `bson:"status" json:"status"`

	// Neu: Arbeitszeit-Regelungen
	WorkingHoursPerWeek  float64       `bson:"workingHoursPerWeek" json:"workingHoursPerWeek"`
	WorkingDaysPerWeek   int           `bson:"workingDaysPerWeek" json:"workingDaysPerWeek"`
	WorkTimeModel        WorkTimeModel `bson:"workTimeModel" json:"workTimeModel"`
	FlexibleWorkingHours bool          `bson:"flexibleWorkingHours" json:"flexibleWorkingHours"`
	CoreWorkingTimeStart string        `bson:"coreWorkingTimeStart" json:"coreWorkingTimeStart"` // Format: "09:00"
	CoreWorkingTimeEnd   string        `bson:"coreWorkingTimeEnd" json:"coreWorkingTimeEnd"`     // Format: "15:00"

	// Zeitkonto-Verwaltung
	OvertimeBalance    float64           `bson:"overtimeBalance" json:"overtimeBalance"`       // Saldo Überstunden in Stunden
	LastTimeCalculated time.Time         `bson:"lastTimeCalculated" json:"lastTimeCalculated"` // Letztes Berechnungsdatum
	WeeklyTimeEntries  []WeeklyTimeEntry `bson:"weeklyTimeEntries" json:"weeklyTimeEntries"`   // Wöchentliche Zusammenfassungen

	// Finanzielle Daten
	Salary          float64 `bson:"salary" json:"salary"`
	BankAccount     string  `bson:"bankAccount" json:"bankAccount"`
	TaxID           string  `bson:"taxId" json:"taxId"`
	SocialSecID     string  `bson:"socialSecId" json:"socialSecId"`
	HealthInsurance string  `bson:"healthInsurance" json:"healthInsurance"`

	// Notfallkontakt
	EmergencyName  string `bson:"emergencyName" json:"emergencyName"`
	EmergencyPhone string `bson:"emergencyPhone" json:"emergencyPhone"`

	// Urlaub und Abwesenheiten
	VacationDays      int       `bson:"vacationDays" json:"vacationDays"`
	RemainingVacation int       `bson:"remainingVacation" json:"remainingVacation"`
	Absences          []Absence `bson:"absences" json:"absences"`

	// Dokumente und weitere Daten
	ProfileImage     string           `bson:"profileImage" json:"profileImage"`
	ProfileImageData primitive.Binary `bson:"profileImageData" json:"profileImageData"`
	Notes            string           `bson:"notes" json:"notes"`

	// Weitere bestehende Felder...
	Documents            []Document          `bson:"documents" json:"documents"`
	ApplicationDocuments []Document          `bson:"applicationDocuments" json:"applicationDocuments"`
	Trainings            []Training          `bson:"trainings" json:"trainings"`
	Evaluations          []Evaluation        `bson:"evaluations" json:"evaluations"`
	DevelopmentPlan      []DevelopmentItem   `bson:"developmentPlan" json:"developmentPlan"`
	Conversations        []Conversation      `bson:"conversations" json:"conversations"`
	ProjectAssignments   []ProjectAssignment `bson:"projectAssignments" json:"projectAssignments"`
	TimeEntries          []TimeEntry         `bson:"timeEntries" json:"timeEntries"`

	// Integration IDs
	TimebutlerUserID string `bson:"timebutlerUserId" json:"timebutlerUserId"`
	Erfasst123ID     string `bson:"erfasst123Id" json:"erfasst123Id"`

	// Timestamps
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// WeeklyTimeEntry repräsentiert die wöchentliche Zeiterfassung
type WeeklyTimeEntry struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WeekStartDate time.Time          `bson:"weekStartDate" json:"weekStartDate"` // Montag der Woche
	WeekEndDate   time.Time          `bson:"weekEndDate" json:"weekEndDate"`     // Sonntag der Woche
	Year          int                `bson:"year" json:"year"`                   // Jahr
	WeekNumber    int                `bson:"weekNumber" json:"weekNumber"`       // Kalenderwoche
	PlannedHours  float64            `bson:"plannedHours" json:"plannedHours"`   // Geplante Stunden der Woche
	ActualHours   float64            `bson:"actualHours" json:"actualHours"`     // Tatsächlich gearbeitete Stunden
	OvertimeHours float64            `bson:"overtimeHours" json:"overtimeHours"` // Überstunden (+/-)
	DaysWorked    int                `bson:"daysWorked" json:"daysWorked"`       // Anzahl gearbeiteter Tage
	IsComplete    bool               `bson:"isComplete" json:"isComplete"`       // Woche abgeschlossen
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// WorkTimeModel repräsentiert verschiedene Arbeitszeitmodelle
type WorkTimeModel string

// GetWorkTimeModelDisplayName gibt den deutschen Anzeigenamen für das Arbeitszeitmodell zurück
func (w WorkTimeModel) GetDisplayName() string {
	switch w {
	case WorkTimeModelFullTime:
		return "Vollzeit"
	case WorkTimeModelPartTime:
		return "Teilzeit"
	case WorkTimeModelFlexTime:
		return "Gleitzeit"
	case WorkTimeModelRemote:
		return "Remote/Homeoffice"
	case WorkTimeModelShift:
		return "Schichtarbeit"
	case WorkTimeModelContract:
		return "Werkvertrag"
	case WorkTimeModelInternship:
		return "Praktikum"
	default:
		return string(w)
	}
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

// Conversation repräsentiert ein Mitarbeitergespräch
type Conversation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Date        time.Time          `bson:"date" json:"date"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Status      string             `bson:"status" json:"status"` // planned, completed
	Notes       string             `bson:"notes" json:"notes"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
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

// ProjectAssignment represents an employee's assignment to a project
type ProjectAssignment struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID   string             `bson:"projectId" json:"projectId"`
	ProjectName string             `bson:"projectName" json:"projectName"`
	StartDate   time.Time          `bson:"startDate" json:"startDate"`
	EndDate     time.Time          `bson:"endDate" json:"endDate"`
	Role        string             `bson:"role,omitempty" json:"role,omitempty"`
	Notes       string             `bson:"notes,omitempty" json:"notes,omitempty"`
	Source      string             `bson:"source" json:"source"` // e.g., "123erfasst"
}

// TimeEntry represents a time entry logged by an employee
type TimeEntry struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Date        time.Time          `bson:"date" json:"date"`
	StartTime   time.Time          `bson:"startTime" json:"startTime"`
	EndTime     time.Time          `bson:"endTime" json:"endTime"`
	Duration    float64            `bson:"duration" json:"duration"` // Duration in hours
	ProjectID   string             `bson:"projectId" json:"projectId"`
	ProjectName string             `bson:"projectName" json:"projectName"`
	Activity    string             `bson:"activity" json:"activity"`
	WageType    string             `bson:"wageType,omitempty" json:"wageType,omitempty"`
	Source      string             `bson:"source" json:"source"` // e.g., "123erfasst"
}

// GetWorkingHoursPerDay berechnet die durchschnittlichen Arbeitsstunden pro Tag
func (e *Employee) GetWorkingHoursPerDay() float64 {
	if e.WorkingDaysPerWeek == 0 {
		return 0
	}
	return e.WorkingHoursPerWeek / float64(e.WorkingDaysPerWeek)
}

// IsFullTimeEmployee prüft, ob es sich um einen Vollzeit-Mitarbeiter handelt
func (e *Employee) IsFullTimeEmployee() bool {
	return e.WorkTimeModel == WorkTimeModelFullTime || e.WorkingHoursPerWeek >= 35
}

// GetWorkingTimeDescription gibt eine textuelle Beschreibung der Arbeitszeit zurück
func (e *Employee) GetWorkingTimeDescription() string {
	if e.WorkingHoursPerWeek == 0 {
		return "Nicht festgelegt"
	}

	description := fmt.Sprintf("%.1f Std/Woche", e.WorkingHoursPerWeek)

	if e.WorkingDaysPerWeek > 0 {
		description += fmt.Sprintf(" (%.1f Std/Tag)", e.GetWorkingHoursPerDay())
	}

	if e.WorkTimeModel != "" {
		description += " - " + e.WorkTimeModel.GetDisplayName()
	}

	return description
}

// GetWeeklyTargetHours berechnet die Ziel-Arbeitsstunden für eine bestimmte Woche
func (e *Employee) GetWeeklyTargetHours() float64 {
	if e.WorkingHoursPerWeek == 0 {
		return 40.0 // Standard-Vollzeit als Fallback
	}
	return e.WorkingHoursPerWeek
}

// FormatOvertimeBalance formatiert das Überstunden-Saldo zur Anzeige
func (e *Employee) FormatOvertimeBalance() string {
	if e.OvertimeBalance >= 0 {
		return fmt.Sprintf("+%.2f Std", e.OvertimeBalance)
	}
	return fmt.Sprintf("%.2f Std", e.OvertimeBalance)
}

// GetOvertimeStatus gibt den Status des Überstunden-Saldos zurück
func (e *Employee) GetOvertimeStatus() string {
	if e.OvertimeBalance > 0 {
		return "positive"
	} else if e.OvertimeBalance < 0 {
		return "negative"
	}
	return "neutral"
}
