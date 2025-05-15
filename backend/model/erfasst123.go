// backend/model/erfasst123.go
package model

import (
	"time"
)

// Erfasst123Person repräsentiert einen Benutzer aus der 123erfasst-API
type Erfasst123Person struct {
	Ident     string             `json:"ident"`
	Firstname string             `json:"firstname"`
	Lastname  string             `json:"lastname"`
	Mail      string             `json:"mail"`
	Employee  Erfasst123Employee `json:"employee"`
}

// Erfasst123Employee enthält die Mitarbeiterdaten aus 123erfasst
type Erfasst123Employee struct {
	IsActive       bool       `json:"isActive"`
	HireDate       string     `json:"hireDate"` // Datum als String von der API
	ExitDate       *string    `json:"exitDate"` // Kann null sein
	HireDateParsed time.Time  `json:"-"`        // Intern geparst
	ExitDateParsed *time.Time `json:"-"`        // Intern geparst
}

type Erfasst123Response struct {
	Data struct {
		Persons struct {
			Nodes      []Erfasst123Person `json:"nodes"`
			TotalCount int                `json:"totalCount"`
		} `json:"persons"`
	} `json:"data"`
}
