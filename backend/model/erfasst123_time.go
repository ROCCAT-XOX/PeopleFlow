package model

import (
	"time"
)

// Erfasst123Time represents a time entry from the 123erfasst API
type Erfasst123Time struct {
	FID       string              `json:"fid"`
	Person    Erfasst123Person    `json:"person"`
	Project   Erfasst123Project   `json:"project"`
	Date      string              `json:"date"`
	TimeStart string              `json:"timeStart"`
	TimeEnd   string              `json:"timeEnd"`
	Activity  Erfasst123Activity  `json:"activity"`
	WageType  *Erfasst123WageType `json:"wageType"`

	// Parsed dates for internal use
	DateParsed      time.Time `json:"-"`
	TimeStartParsed time.Time `json:"-"`
	TimeEndParsed   time.Time `json:"-"`
	Duration        float64   `json:"-"` // Duration in hours
}

// Erfasst123Activity represents an activity type from 123erfasst
type Erfasst123Activity struct {
	Ident string `json:"ident"`
	Name  string `json:"name"`
}

// Erfasst123WageType represents a wage type from 123erfasst
type Erfasst123WageType struct {
	Ident string `json:"ident"`
	Name  string `json:"name"`
}

// Erfasst123TimeResponse represents the GraphQL API response structure for time entries
type Erfasst123TimeResponse struct {
	Data struct {
		Times struct {
			Nodes      []Erfasst123Time `json:"nodes"`
			TotalCount int              `json:"totalCount"`
		} `json:"times"`
	} `json:"data"`
}
