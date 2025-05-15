// backend/model/erfasst123_project.go
package model

import (
	"time"
)

// Erfasst123Planning represents project planning data from 123erfasst
type Erfasst123Planning struct {
	Project   Erfasst123Project  `json:"project"`
	Persons   []Erfasst123Person `json:"persons"`
	DateStart string             `json:"dateStart"`
	DateEnd   string             `json:"dateEnd"`
	// Parsed dates for internal use
	DateStartParsed time.Time `json:"-"`
	DateEndParsed   time.Time `json:"-"`
}

// Erfasst123Project represents a project from 123erfasst
type Erfasst123Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Erfasst123PlanningResponse represents the GraphQL API response structure
type Erfasst123PlanningResponse struct {
	Data struct {
		Plannings struct {
			Nodes      []Erfasst123Planning `json:"nodes"`
			TotalCount int                  `json:"totalCount"`
		} `json:"plannings"`
	} `json:"data"`
}
