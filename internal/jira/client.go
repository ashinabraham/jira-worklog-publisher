package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"

	"jira-calendar/internal/models"
)

// Client wraps the go-jira client with our custom logic
type Client struct {
	client       *jira.Client
	httpClient   *http.Client
	baseURL      string
	username     string
	apiToken     string
	AccountID    string
	DisplayName  string
	EmailAddress string
	TimeZone     string
	AvatarURL    string
	Locale       string
	AccountType  string
	Active       bool
	SelfURL      string
}

// NewClient creates a new Jira client with authentication
func NewClient(config models.JiraConfig) (*Client, error) {
	log.Printf("[JIRA] Creating new Jira client")
	log.Printf("[JIRA] BaseURL: %s", config.BaseURL)
	log.Printf("[JIRA] Username: %s", config.Username)

	// Normalize base URL - remove trailing slash
	baseURL := strings.TrimSuffix(config.BaseURL, "/")

	// Setup authentication
	tp := jira.BasicAuthTransport{
		Username: config.Username,
		Password: config.APIToken,
	}

	// Create Jira client
	client, err := jira.NewClient(tp.Client(), baseURL)
	if err != nil {
		log.Printf("[JIRA] Error creating Jira client: %v", err)
		return nil, fmt.Errorf("failed to create Jira client: %v", err)
	}

	log.Printf("[JIRA] Jira client created, authenticating...")

	// Get current user's account ID
	user, _, err := client.User.GetSelf()
	if err != nil {
		log.Printf("[JIRA] Authentication failed: %v", err)
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}

	log.Printf("[JIRA] Authentication successful")
	log.Printf("[JIRA] User: %s (Account ID: %s)", user.DisplayName, user.AccountID)
	log.Printf("[JIRA] Email: %s, TimeZone: %s", user.EmailAddress, user.TimeZone)

	// Get largest avatar URL
	avatarURL := ""
	if user.AvatarUrls.Four8X48 != "" {
		avatarURL = user.AvatarUrls.Four8X48
	} else if user.AvatarUrls.Three2X32 != "" {
		avatarURL = user.AvatarUrls.Three2X32
	} else if user.AvatarUrls.Two4X24 != "" {
		avatarURL = user.AvatarUrls.Two4X24
	} else if user.AvatarUrls.One6X16 != "" {
		avatarURL = user.AvatarUrls.One6X16
	}

	log.Printf("[JIRA] Avatar URL: %s", avatarURL)
	log.Printf("[JIRA] Locale: %s, Account Type: %s, Active: %t", user.Locale, user.AccountType, user.Active)
	log.Printf("[JIRA] Self URL: %s", user.Self)

	return &Client{
		client:       client,
		httpClient:   tp.Client(),
		baseURL:      baseURL,
		username:     config.Username,
		apiToken:     config.APIToken,
		AccountID:    user.AccountID,
		DisplayName:  user.DisplayName,
		EmailAddress: user.EmailAddress,
		TimeZone:     user.TimeZone,
		AvatarURL:    avatarURL,
		Locale:       user.Locale,
		AccountType:  user.AccountType,
		Active:       user.Active,
		SelfURL:      user.Self,
	}, nil
}

// searchIssuesWithJQL performs a JQL search using the correct API endpoint
func (jc *Client) searchIssuesWithJQL(jql string) ([]jira.Issue, error) {
	type SearchRequest struct {
		JQL           string   `json:"jql"`
		MaxResults    int      `json:"maxResults"`
		Fields        []string `json:"fields"`
		NextPageToken string   `json:"nextPageToken,omitempty"`
	}

	type SearchResponse struct {
		Issues        []jira.Issue `json:"issues"`
		IsLast        bool         `json:"isLast"`
		NextPageToken string       `json:"nextPageToken,omitempty"`
	}

	var allIssues []jira.Issue
	nextPageToken := ""

	for {
		reqBody := SearchRequest{
			JQL:           jql,
			MaxResults:    100,
			Fields:        []string{"id", "key", "summary"},
			NextPageToken: nextPageToken,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal search request: %v", err)
		}

		url := fmt.Sprintf("%s/rest/api/3/search/jql", jc.baseURL)
		log.Printf("[JIRA] POST %s", url)
		log.Printf("[JIRA] Request body: %s", string(jsonBody))

		req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		req.SetBasicAuth(jc.username, jc.apiToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := jc.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		log.Printf("[JIRA] Response status: %d", resp.StatusCode)

		if resp.StatusCode != 200 {
			log.Printf("[JIRA] Response body: %s", string(body))
			return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
		}

		var searchResp SearchResponse
		if err := json.Unmarshal(body, &searchResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}

		allIssues = append(allIssues, searchResp.Issues...)
		log.Printf("[JIRA] Fetched %d issues (total so far: %d, isLast: %v)", len(searchResp.Issues), len(allIssues), searchResp.IsLast)

		// Check if this is the last page
		if searchResp.IsLast || len(searchResp.Issues) == 0 {
			break
		}

		// Use the next page token for the next iteration
		nextPageToken = searchResp.NextPageToken
	}

	return allIssues, nil
}

// GetWorkLogs fetches work logs for the given date range using JQL search
func (jc *Client) GetWorkLogs(startDate, endDate time.Time) ([]models.WorkLog, error) {
	// Build JQL query for issues with worklogs in date range
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	jql := fmt.Sprintf("worklogDate >= %s AND worklogDate <= %s AND project = \"ARTC Service\" ORDER BY updated DESC", startStr, endStr)

	log.Printf("[JIRA] Fetching worklogs from %s to %s", startStr, endStr)
	log.Printf("[JIRA] JQL Query: %s", jql)

	// Search for issues using custom implementation with correct API endpoint
	allIssues, err := jc.searchIssuesWithJQL(jql)
	if err != nil {
		log.Printf("[JIRA] Error searching issues: %v", err)
		return nil, fmt.Errorf("failed to search issues: %v", err)
	}

	log.Printf("[JIRA] Found %d issues with worklogs", len(allIssues))

	// Get worklogs for each issue in parallel
	type result struct {
		worklogs []models.WorkLog
		err      error
	}

	resultChan := make(chan result, len(allIssues))

	for _, issue := range allIssues {
		go func(issueID, issueKey, issueSummary string) {
			log.Printf("[JIRA] Fetching worklogs for issue: %s (ID: %s)", issueKey, issueID)
			worklogs, err := jc.getWorkLogsForIssue(issueID, issueKey, issueSummary, startDate, endDate)
			if err != nil {
				log.Printf("[JIRA] Error fetching worklogs for %s: %v", issueKey, err)
			} else {
				log.Printf("[JIRA] Found %d worklogs for issue %s", len(worklogs), issueKey)
			}
			resultChan <- result{worklogs: worklogs, err: err}
		}(issue.ID, issue.Key, issue.Fields.Summary)
	}

	// Collect results
	var allWorkLogs []models.WorkLog
	for i := 0; i < len(allIssues); i++ {
		res := <-resultChan
		if res.err == nil {
			allWorkLogs = append(allWorkLogs, res.worklogs...)
		}
	}

	log.Printf("[JIRA] Total worklogs fetched: %d", len(allWorkLogs))
	return allWorkLogs, nil
}

// getWorkLogsForIssue fetches work logs for a specific issue
func (jc *Client) getWorkLogsForIssue(issueID, issueKey, issueSummary string, startDate, endDate time.Time) ([]models.WorkLog, error) {
	// Get worklogs for the issue using issue ID (more reliable than key)
	worklogRecord, _, err := jc.client.Issue.GetWorklogs(issueID)
	if err != nil {
		log.Printf("[JIRA] Error getting worklogs for issue %s (ID: %s): %v", issueKey, issueID, err)
		return nil, err
	}

	log.Printf("[JIRA] Issue %s has %d total worklogs", issueKey, len(worklogRecord.Worklogs))

	// Filter and convert worklogs
	var filtered []models.WorkLog
	for _, wl := range worklogRecord.Worklogs {
		// wl.Started is a *jira.Time type which embeds time.Time
		if wl.Started == nil {
			log.Printf("[JIRA] Skipping worklog with nil start time in issue %s", issueKey)
			continue
		}
		started := time.Time(*wl.Started)

		// Check if worklog is in date range
		worklogDate := time.Date(started.Year(), started.Month(), started.Day(), 0, 0, 0, 0, time.UTC)
		startDateOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDateOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

		if !worklogDate.Before(startDateOnly) && !worklogDate.After(endDateOnly) {
			// Check if worklog is by current user
			if wl.Author != nil && wl.Author.AccountID == jc.AccountID {
				log.Printf("[JIRA] Including worklog: Issue=%s, Date=%s, Hours=%.2f, Author=%s",
					issueKey, started.Format("2006-01-02"), float64(wl.TimeSpentSeconds)/3600.0, wl.Author.DisplayName)
				filtered = append(filtered, mapJiraWorklogToWorkLog(&wl, issueKey, issueSummary, started))
			} else if wl.Author != nil {
				log.Printf("[JIRA] Skipping worklog (different author): Issue=%s, Author=%s (AccountID: %s)",
					issueKey, wl.Author.DisplayName, wl.Author.AccountID)
			}
		} else {
			log.Printf("[JIRA] Skipping worklog (out of date range): Issue=%s, Date=%s",
				issueKey, started.Format("2006-01-02"))
		}
	}

	log.Printf("[JIRA] Filtered %d worklogs for issue %s (current user, date range)", len(filtered), issueKey)
	return filtered, nil
}

// mapJiraWorklogToWorkLog converts a jira.WorklogRecord to our WorkLog
func mapJiraWorklogToWorkLog(jiraWL *jira.WorklogRecord, issueKey, issueSummary string, started time.Time) models.WorkLog {
	comment := jiraWL.Comment

	return models.WorkLog{
		ID:               jiraWL.ID,
		IssueKey:         issueKey,
		IssueSummary:     issueSummary,
		TimeSpent:        jiraWL.TimeSpent,
		TimeSpentSeconds: jiraWL.TimeSpentSeconds,
		Started:          started,
		Author: models.WorkLogAuthor{
			DisplayName:  jiraWL.Author.DisplayName,
			AccountID:    jiraWL.Author.AccountID,
			EmailAddress: jiraWL.Author.EmailAddress,
		},
		Comment: comment,
	}
}

// AggregateWorkLogs groups work logs by date and calculates totals
func AggregateWorkLogs(workLogs []models.WorkLog) map[string]*models.DayWorkSummary {
	summary := make(map[string]*models.DayWorkSummary)

	for _, wl := range workLogs {
		dateKey := wl.Started.Format("2006-01-02")
		if summary[dateKey] == nil {
			summary[dateKey] = &models.DayWorkSummary{
				Date:       time.Date(wl.Started.Year(), wl.Started.Month(), wl.Started.Day(), 0, 0, 0, 0, wl.Started.Location()),
				Tickets:    []models.TicketWork{},
				TotalHours: 0,
			}
		}

		hours := float64(wl.TimeSpentSeconds) / 3600.0
		summary[dateKey].TotalHours += hours

		// Find if ticket already exists in this day's summary
		ticketFound := false
		for i := range summary[dateKey].Tickets {
			if summary[dateKey].Tickets[i].IssueKey == wl.IssueKey {
				// Add hours to existing ticket
				summary[dateKey].Tickets[i].Hours += hours
				// Add comment if not empty
				if wl.Comment != "" {
					summary[dateKey].Tickets[i].Comments = append(summary[dateKey].Tickets[i].Comments, wl.Comment)
				}
				ticketFound = true
				break
			}
		}

		// If ticket not found, add new ticket
		if !ticketFound {
			comments := []string{}
			if wl.Comment != "" {
				comments = append(comments, wl.Comment)
			}
			summary[dateKey].Tickets = append(summary[dateKey].Tickets, models.TicketWork{
				IssueKey: wl.IssueKey,
				Summary:  wl.IssueSummary,
				Hours:    hours,
				Comments: comments,
			})
		}
	}

	return summary
}
