// Package jira provides a Jira REST API v3 client with pluggable authentication.
// It supports fetching user info, searching issues via JQL, retrieving work logs,
// and creating work logs (with ADF comment format).
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

	"jira-calendar/internal/models"
)

// Authenticator is the interface for Jira authentication (e.g. BasicAuth, future OAuth).
type Authenticator interface {
	// AuthenticateRequest adds authentication headers/credentials to the HTTP request
	AuthenticateRequest(req *http.Request) error
	// GetHTTPClient returns an HTTP client configured for this auth method
	GetHTTPClient() *http.Client
}

// BasicAuth implements Authenticator using Jira username (email) and API token.
type BasicAuth struct {
	username string
	apiToken string
	client   *http.Client
}

// NewBasicAuth returns a BasicAuth authenticator for the given username and API token.
func NewBasicAuth(username, apiToken string) Authenticator {
	return &BasicAuth{
		username: username,
		apiToken: apiToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AuthenticateRequest sets Basic Auth header on the request.
func (b *BasicAuth) AuthenticateRequest(req *http.Request) error {
	req.SetBasicAuth(b.username, b.apiToken)
	return nil
}

// GetHTTPClient returns the HTTP client used by this authenticator.
func (b *BasicAuth) GetHTTPClient() *http.Client {
	return b.client
}

// Client is the Jira REST API v3 client. AccountID is set after successful auth for filtering work logs.
type Client struct {
	httpClient *http.Client
	baseURL    string
	auth       Authenticator
	AccountID  string // Needed for filtering worklogs by current user
}

// Jira API response types
type jiraUser struct {
	AccountID    string            `json:"accountId"`
	DisplayName  string            `json:"displayName"`
	EmailAddress string            `json:"emailAddress"`
	TimeZone     string            `json:"timeZone"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
	Locale       string            `json:"locale"`
	AccountType  string            `json:"accountType"`
	Active       bool              `json:"active"`
	Self         string            `json:"self"`
}

type jiraIssue struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Fields jiraFields   `json:"fields"`
}

type jiraFields struct {
	Summary string `json:"summary"`
}

// jiraComment can be a string or an object (rich text)
type jiraComment struct {
	raw json.RawMessage
}

// UnmarshalJSON handles both string and object comment formats
func (c *jiraComment) UnmarshalJSON(data []byte) error {
	c.raw = data
	return nil
}

// String extracts the comment text from either format
func (c *jiraComment) String() string {
	if len(c.raw) == 0 {
		return ""
	}
	
	// Try as string first
	var str string
	if err := json.Unmarshal(c.raw, &str); err == nil {
		return str
	}
	
	// Try as rich text object (Atlassian Document Format)
	var richText struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	}
	if err := json.Unmarshal(c.raw, &richText); err == nil {
		// Extract text from content array
		var text strings.Builder
		for _, item := range richText.Content {
			if item.Type == "text" {
				text.WriteString(item.Text)
			} else if item.Type == "paragraph" {
				// Handle nested content in paragraphs
				for _, nested := range item.Content {
					if nested.Type == "text" {
						text.WriteString(nested.Text)
					}
				}
			}
		}
		return text.String()
	}
	
	return ""
}

type jiraWorklogRecord struct {
	ID               string      `json:"id"`
	TimeSpent        string      `json:"timeSpent"`
	TimeSpentSeconds int         `json:"timeSpentSeconds"`
	Started          string      `json:"started"`
	Author           jiraUser    `json:"author"`
	Comment          jiraComment `json:"comment"`
}

type jiraWorklogResponse struct {
	Worklogs []jiraWorklogRecord `json:"worklogs"`
	MaxResults int               `json:"maxResults"`
	StartAt    int               `json:"startAt"`
	Total      int               `json:"total"`
}

type jiraSearchRequest struct {
	JQL           string   `json:"jql"`
	MaxResults    int      `json:"maxResults"`
	Fields        []string `json:"fields"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

type jiraSearchResponse struct {
	Issues        []jiraIssue `json:"issues"`
	IsLast        bool        `json:"isLast"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

// NewClient creates a new Jira client, normalizes baseURL, and verifies auth by fetching the current user.
func NewClient(baseURL string, auth Authenticator) (*Client, error) {
	log.Printf("[JIRA] Creating new Jira client")
	log.Printf("[JIRA] BaseURL: %s", baseURL)

	// Normalize base URL - remove trailing slash
	normalizedURL := strings.TrimSuffix(baseURL, "/")

	client := &Client{
		httpClient: auth.GetHTTPClient(),
		baseURL:    normalizedURL,
		auth:       auth,
	}

	// Authenticate and get user info to verify connection and get AccountID
	user, err := client.getCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}

	// Store AccountID for filtering worklogs
	client.AccountID = user.AccountID

	log.Printf("[JIRA] Authentication successful")
	log.Printf("[JIRA] User: %s (Account ID: %s)", user.DisplayName, client.AccountID)
	log.Printf("[JIRA] Email: %s", user.EmailAddress)

	return client, nil
}

// getCurrentUser calls /rest/api/3/myself to fetch the current user.
func (jc *Client) getCurrentUser() (*jiraUser, error) {
	url := fmt.Sprintf("%s/rest/api/3/myself", jc.baseURL)
	log.Printf("[JIRA] GET %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if err := jc.auth.AuthenticateRequest(req); err != nil {
		return nil, fmt.Errorf("failed to authenticate request: %v", err)
	}
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
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var user jiraUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user response: %v", err)
	}

	return &user, nil
}

// GetUserInfo returns the current user's profile as a map (displayName, avatarURL, timeZone, etc.).
func (jc *Client) GetUserInfo() (map[string]interface{}, error) {
	user, err := jc.getCurrentUser()
	if err != nil {
		return nil, err
	}

	// Get largest available avatar URL
	avatarURL := ""
	if avatar48, ok := user.AvatarUrls["48x48"]; ok && avatar48 != "" {
		avatarURL = avatar48
	} else if avatar32, ok := user.AvatarUrls["32x32"]; ok && avatar32 != "" {
		avatarURL = avatar32
	} else if avatar24, ok := user.AvatarUrls["24x24"]; ok && avatar24 != "" {
		avatarURL = avatar24
	} else if avatar16, ok := user.AvatarUrls["16x16"]; ok && avatar16 != "" {
		avatarURL = avatar16
	}

	return map[string]interface{}{
		"accountID":    user.AccountID,
		"displayName":  user.DisplayName,
		"emailAddress": user.EmailAddress,
		"avatarURL":    avatarURL,
		"timeZone":     user.TimeZone,
		"locale":       user.Locale,
		"accountType":  user.AccountType,
		"active":       user.Active,
	}, nil
}

// searchIssuesWithJQL runs a JQL search with pagination (POST /rest/api/3/search/jql).
func (jc *Client) searchIssuesWithJQL(jql string) ([]jiraIssue, error) {
	var allIssues []jiraIssue
	nextPageToken := ""

	for {
		reqBody := jiraSearchRequest{
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

		if err := jc.auth.AuthenticateRequest(req); err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %v", err)
		}
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

		var searchResp jiraSearchResponse
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

// GetWorkLogs fetches work logs in the date range via JQL search and worklog API, filtered by current user.
func (jc *Client) GetWorkLogs(startDate, endDate time.Time) ([]models.WorkLog, error) {
	// Build JQL query for issues with worklogs in date range
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")
	jql := fmt.Sprintf("worklogDate >= %s AND worklogDate <= %s AND project = \"ARTC Service\" ORDER BY updated DESC", startStr, endStr)

	log.Printf("[JIRA] Fetching worklogs from %s to %s", startStr, endStr)
	log.Printf("[JIRA] JQL Query: %s", jql)

	// Search for issues
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
	var allWorklogs []jiraWorklogRecord
	startAt := 0
	maxResults := 1000

	for {
		url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog?startAt=%d&maxResults=%d", jc.baseURL, issueID, startAt, maxResults)
		log.Printf("[JIRA] GET %s", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		if err := jc.auth.AuthenticateRequest(req); err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %v", err)
		}
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
			return nil, fmt.Errorf("failed to get worklogs with status %d: %s", resp.StatusCode, string(body))
		}

		var worklogResp jiraWorklogResponse
		if err := json.Unmarshal(body, &worklogResp); err != nil {
			return nil, fmt.Errorf("failed to parse worklog response: %v", err)
		}

		allWorklogs = append(allWorklogs, worklogResp.Worklogs...)
		log.Printf("[JIRA] Fetched %d worklogs (total: %d/%d)", len(worklogResp.Worklogs), len(allWorklogs), worklogResp.Total)

		// Check if we've fetched all worklogs
		if startAt+len(worklogResp.Worklogs) >= worklogResp.Total {
			break
		}

		startAt += len(worklogResp.Worklogs)
	}

	log.Printf("[JIRA] Issue %s has %d total worklogs", issueKey, len(allWorklogs))

	// Filter and convert worklogs
	var filtered []models.WorkLog
	for _, wl := range allWorklogs {
		// Parse started time
		started, err := time.Parse("2006-01-02T15:04:05.000-0700", wl.Started)
		if err != nil {
			// Try alternative format
			started, err = time.Parse("2006-01-02T15:04:05.000Z", wl.Started)
			if err != nil {
				log.Printf("[JIRA] Skipping worklog with invalid start time in issue %s: %s", issueKey, wl.Started)
				continue
			}
		}

		// Check if worklog is in date range
		worklogDate := time.Date(started.Year(), started.Month(), started.Day(), 0, 0, 0, 0, time.UTC)
		startDateOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDateOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

		if !worklogDate.Before(startDateOnly) && !worklogDate.After(endDateOnly) {
			// Check if worklog is by current user
			if wl.Author.AccountID == jc.AccountID {
				log.Printf("[JIRA] Including worklog: Issue=%s, Date=%s, Hours=%.2f, Author=%s",
					issueKey, started.Format("2006-01-02"), float64(wl.TimeSpentSeconds)/3600.0, wl.Author.DisplayName)
				filtered = append(filtered, mapJiraWorklogToWorkLog(&wl, issueKey, issueSummary, started))
			} else {
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

// mapJiraWorklogToWorkLog converts a jiraWorklogRecord to our WorkLog
func mapJiraWorklogToWorkLog(jiraWL *jiraWorklogRecord, issueKey, issueSummary string, started time.Time) models.WorkLog {
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
		Comment: jiraWL.Comment.String(),
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

// CreateWorklog creates a new worklog entry for the given issue
// CreateWorklog POSTs a work log to the given issue (comment is sent in ADF format).
func (jc *Client) CreateWorklog(issueKey string, req models.WorklogRequest) (*models.WorklogResponse, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", jc.baseURL, issueKey)
	log.Printf("[JIRA] POST %s", url)

	// Marshal request body - Jira API v3 expects comment in Atlassian Document Format (ADF)
	worklogBody := map[string]interface{}{
		"timeSpentSeconds": req.TimeSpentSeconds,
		"started":          req.Started,
	}
	
	// Format comment in Atlassian Document Format if provided
	if req.Comment != "" {
		worklogBody["comment"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": req.Comment,
						},
					},
				},
			},
		}
	}
	
	jsonData, err := json.Marshal(worklogBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal worklog request: %v", err)
	}
	
	log.Printf("[JIRA] Request body: %s", string(jsonData))

	// Create request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Authenticate request
	if err := jc.auth.AuthenticateRequest(httpReq); err != nil {
		return nil, fmt.Errorf("failed to authenticate request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := jc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[JIRA] Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated {
		log.Printf("[JIRA] Response body: %s", string(body))
		
		// Try to parse Jira error response
		var jiraError struct {
			ErrorMessages []string          `json:"errorMessages"`
			Errors        map[string]string `json:"errors"`
		}
		if err := json.Unmarshal(body, &jiraError); err == nil {
			if len(jiraError.ErrorMessages) > 0 {
				return nil, fmt.Errorf("%s", jiraError.ErrorMessages[0])
			}
			if len(jiraError.Errors) > 0 {
				for _, msg := range jiraError.Errors {
					return nil, fmt.Errorf("%s", msg)
				}
			}
		}
		
		// Fallback to generic error
		return nil, fmt.Errorf("failed to create worklog: HTTP %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var worklogResp models.WorklogResponse
	if err := json.Unmarshal(body, &worklogResp); err != nil {
		return nil, fmt.Errorf("failed to parse worklog response: %v", err)
	}

	log.Printf("[JIRA] Worklog created successfully: ID=%s", worklogResp.ID)
	return &worklogResp, nil
}
