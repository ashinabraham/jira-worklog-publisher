// Package models defines data structures for Jira configuration, work logs,
// and API request/response types (including Atlassian Document Format for comments).
package models

import (
	"encoding/json"
	"strings"
	"time"
)

// JiraConfig holds the Jira instance configuration (base URL, username, API token).
type JiraConfig struct {
	BaseURL  string `json:"baseURL"`
	Username string `json:"username"`
	APIToken string `json:"apiToken"`
}

// WorkLog represents a single work log entry as returned by the Jira API.
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

// WorkLogAuthor holds author details for a work log.
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

// WorklogRequest is the payload for creating a work log (Jira REST API v3).
type WorklogRequest struct {
	Comment          string `json:"comment"`
	Started          string `json:"started"`          // ISO 8601 format: "2024-01-15T10:00:00.000+0000"
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

// ADFComment represents an Atlassian Document Format (ADF) comment structure from Jira API v3.
type ADFComment struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Content []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"content"`
}

// WorklogResponse is the response from the create work log API. Comment may be string or ADF object.
type WorklogResponse struct {
	ID               string      `json:"id"`
	Self             string      `json:"self"`
	Comment          interface{} `json:"comment"` // Can be string or ADFComment
	Created          string      `json:"created"`
	Updated          string      `json:"updated"`
	Started          string      `json:"started"`
	TimeSpent        string      `json:"timeSpent"`
	TimeSpentSeconds int         `json:"timeSpentSeconds"`
}

// GetCommentText returns the plain text of the comment, whether it was stored as a string or ADF.
func (wr *WorklogResponse) GetCommentText() string {
	if wr.Comment == nil {
		return ""
	}
	
	// If it's already a string, return it
	if str, ok := wr.Comment.(string); ok {
		return str
	}
	
	// Try to parse as ADF format (map[string]interface{})
	commentMap, ok := wr.Comment.(map[string]interface{})
	if !ok {
		// Try to unmarshal from JSON bytes if it's not a map
		commentBytes, err := json.Marshal(wr.Comment)
		if err != nil {
			return ""
		}
		if err := json.Unmarshal(commentBytes, &commentMap); err != nil {
			return ""
		}
	}
	
	// Extract text from ADF structure
	var textParts []string
	if content, ok := commentMap["content"].([]interface{}); ok {
		for _, para := range content {
			if paraMap, ok := para.(map[string]interface{}); ok {
				if paraType, _ := paraMap["type"].(string); paraType == "paragraph" {
					if paraContent, ok := paraMap["content"].([]interface{}); ok {
						for _, textElem := range paraContent {
							if textMap, ok := textElem.(map[string]interface{}); ok {
								if textType, _ := textMap["type"].(string); textType == "text" {
									if text, ok := textMap["text"].(string); ok {
										textParts = append(textParts, text)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	return strings.Join(textParts, " ")
}
