package models

import "time"

// JiraConfig holds the Jira instance configuration
type JiraConfig struct {
	BaseURL  string `json:"baseURL"`
	Username string `json:"username"`
	APIToken string `json:"apiToken"`
}

// WorkLog represents a single work log entry
type WorkLog struct {
	ID               string
	IssueKey         string
	IssueSummary     string
	TimeSpent        string
	TimeSpentSeconds int
	Started          time.Time
	Author           WorkLogAuthor
	Comment          string
}

// WorkLogAuthor represents the author of a work log
type WorkLogAuthor struct {
	DisplayName  string
	AccountID    string
	EmailAddress string
}

// DayWorkSummary aggregates work logs for a single day
type DayWorkSummary struct {
	Date       time.Time
	Tickets    []TicketWork
	TotalHours float64
}

// TicketWork represents work done on a specific ticket
type TicketWork struct {
	IssueKey string
	Summary  string
	Hours    float64
	Comments []string
}
