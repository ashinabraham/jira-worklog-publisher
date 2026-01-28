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
	ID               string        `json:"id"`
	IssueKey         string        `json:"issueKey"`
	IssueSummary     string        `json:"issueSummary"`
	TimeSpent        string        `json:"timeSpent"`
	TimeSpentSeconds int           `json:"timeSpentSeconds"`
	Started          time.Time     `json:"started"`
	Author           WorkLogAuthor `json:"author"`
	Comment          string        `json:"comment"`
}

// WorkLogAuthor represents the author of a work log
type WorkLogAuthor struct {
	DisplayName  string `json:"displayName"`
	AccountID    string `json:"accountID"`
	EmailAddress string `json:"emailAddress"`
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
