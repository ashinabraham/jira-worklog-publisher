package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jira-worklog-publisher/internal/models"
)

func TestNewBasicAuth(t *testing.T) {
	auth := NewBasicAuth("user", "token")
	if auth == nil {
		t.Fatal("NewBasicAuth returned nil")
	}
	if auth.(*BasicAuth).GetHTTPClient() == nil {
		t.Error("GetHTTPClient() returned nil")
	}
}

func TestBasicAuth_AuthenticateRequest(t *testing.T) {
	auth := NewBasicAuth("alice@example.com", "secret").(*BasicAuth)
	req, err := http.NewRequest("GET", "https://jira.example.com/rest/api/3/myself", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := auth.AuthenticateRequest(req); err != nil {
		t.Fatalf("AuthenticateRequest: %v", err)
	}
	user, pass, ok := req.BasicAuth()
	if !ok || user != "alice@example.com" || pass != "secret" {
		t.Errorf("BasicAuth not set: ok=%v user=%q pass=%q", ok, user, pass)
	}
}

func TestAggregateWorkLogs_Empty(t *testing.T) {
	got := AggregateWorkLogs(nil)
	if len(got) != 0 {
		t.Errorf("AggregateWorkLogs(nil) = %d entries, want 0", len(got))
	}
	got = AggregateWorkLogs([]models.WorkLog{})
	if len(got) != 0 {
		t.Errorf("AggregateWorkLogs([]) = %d entries, want 0", len(got))
	}
}

func TestAggregateWorkLogs_SingleDaySingleTicket(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	workLogs := []models.WorkLog{
		{
			ID:               "1",
			IssueKey:         "PROJ-1",
			IssueSummary:     "First",
			TimeSpentSeconds: 3600,
			Started:          base,
			Comment:          "did work",
		},
	}
	got := AggregateWorkLogs(workLogs)
	if len(got) != 1 {
		t.Fatalf("got %d days, want 1", len(got))
	}
	day := got["2024-01-15"]
	if day == nil {
		t.Fatal("missing key 2024-01-15")
	}
	if day.TotalHours != 1.0 {
		t.Errorf("TotalHours = %v, want 1", day.TotalHours)
	}
	if len(day.Tickets) != 1 {
		t.Fatalf("Tickets len = %d, want 1", len(day.Tickets))
	}
	if day.Tickets[0].IssueKey != "PROJ-1" || day.Tickets[0].Hours != 1.0 || len(day.Tickets[0].Comments) != 1 || day.Tickets[0].Comments[0] != "did work" {
		t.Errorf("Tickets[0] = %+v", day.Tickets[0])
	}
}

func TestAggregateWorkLogs_SameTicketMultipleLogs(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	workLogs := []models.WorkLog{
		{ID: "1", IssueKey: "PROJ-1", IssueSummary: "S", TimeSpentSeconds: 3600, Started: base, Comment: "a"},
		{ID: "2", IssueKey: "PROJ-1", IssueSummary: "S", TimeSpentSeconds: 1800, Started: base.Add(2 * time.Hour), Comment: "b"},
	}
	got := AggregateWorkLogs(workLogs)
	day := got["2024-01-15"]
	if day == nil {
		t.Fatal("missing day")
	}
	if day.TotalHours != 1.5 {
		t.Errorf("TotalHours = %v, want 1.5", day.TotalHours)
	}
	if len(day.Tickets) != 1 {
		t.Fatalf("Tickets len = %d, want 1", len(day.Tickets))
	}
	if day.Tickets[0].Hours != 1.5 {
		t.Errorf("Tickets[0].Hours = %v, want 1.5", day.Tickets[0].Hours)
	}
	if len(day.Tickets[0].Comments) != 2 {
		t.Errorf("Comments len = %d, want 2", len(day.Tickets[0].Comments))
	}
}

func TestAggregateWorkLogs_MultipleDaysAndTickets(t *testing.T) {
	d1 := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 1, 16, 9, 0, 0, 0, time.UTC)
	workLogs := []models.WorkLog{
		{ID: "1", IssueKey: "A-1", IssueSummary: "A", TimeSpentSeconds: 3600, Started: d1},
		{ID: "2", IssueKey: "B-1", IssueSummary: "B", TimeSpentSeconds: 7200, Started: d1},
		{ID: "3", IssueKey: "A-1", IssueSummary: "A", TimeSpentSeconds: 1800, Started: d2},
	}
	got := AggregateWorkLogs(workLogs)
	if len(got) != 2 {
		t.Fatalf("got %d days, want 2", len(got))
	}
	if got["2024-01-15"].TotalHours != 3.0 || len(got["2024-01-15"].Tickets) != 2 {
		t.Errorf("2024-01-15: %+v", got["2024-01-15"])
	}
	if got["2024-01-16"].TotalHours != 0.5 || len(got["2024-01-16"].Tickets) != 1 {
		t.Errorf("2024-01-16: %+v", got["2024-01-16"])
	}
}

func TestAggregateWorkLogs_EmptyCommentNotAppended(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	workLogs := []models.WorkLog{
		{ID: "1", IssueKey: "P-1", IssueSummary: "S", TimeSpentSeconds: 3600, Started: base, Comment: ""},
	}
	got := AggregateWorkLogs(workLogs)
	day := got["2024-01-15"]
	if day == nil || len(day.Tickets) != 1 {
		t.Fatal("expected one ticket")
	}
	if len(day.Tickets[0].Comments) != 0 {
		t.Errorf("expected no comments, got %v", day.Tickets[0].Comments)
	}
}

func TestNewClient_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/myself" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if user, _, ok := r.BasicAuth(); !ok || user == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"acc-123","displayName":"Test User","emailAddress":"test@example.com","timeZone":"UTC"}`))
	}))
	defer server.Close()

	auth := NewBasicAuth("user", "token")
	client, err := NewClient(server.URL, auth)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.AccountID != "acc-123" {
		t.Errorf("AccountID = %q, want acc-123", client.AccountID)
	}
}

func TestNewClient_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	auth := NewBasicAuth("user", "bad")
	_, err := NewClient(server.URL, auth)
	if err == nil {
		t.Fatal("NewClient expected error on 401")
	}
}

func TestNewClient_NormalizesBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"x","displayName":"","emailAddress":"","timeZone":""}`))
	}))
	defer server.Close()

	auth := NewBasicAuth("u", "t")
	client, err := NewClient(server.URL+"/", auth)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	// baseURL should not have trailing slash so API paths are correct
	if client.baseURL != server.URL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, server.URL)
	}
}

func TestClient_GetUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"id","displayName":"Dev","emailAddress":"dev@ex.com","avatarUrls":{"48x48":"https://avatar/48"},"timeZone":"UTC","locale":"en","accountType":"atlassian","active":true}`))
	}))
	defer server.Close()

	auth := NewBasicAuth("u", "t")
	client, err := NewClient(server.URL, auth)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	info, err := client.GetUserInfo()
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if info["displayName"] != "Dev" || info["accountID"] != "id" || info["avatarURL"] != "https://avatar/48" {
		t.Errorf("GetUserInfo = %+v", info)
	}
}

func TestClient_CreateWorklog_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/rest/api/3/myself":
			w.Write([]byte(`{"accountId":"x","displayName":"","emailAddress":"","timeZone":""}`))
			return
		case "/rest/api/3/issue/PROJ-1/worklog":
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"wl-1","self":"https://jira/rest/api/3/issue/PROJ-1/worklog/wl-1","started":"2024-01-15T10:00:00.000+0000","timeSpentSeconds":3600,"timeSpent":"1h","comment":"did work"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	auth := NewBasicAuth("u", "t")
	client, err := NewClient(server.URL, auth)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	req := models.WorklogRequest{
		Comment:          "did work",
		Started:          "2024-01-15T10:00:00.000+0000",
		TimeSpentSeconds: 3600,
	}
	resp, err := client.CreateWorklog("PROJ-1", req)
	if err != nil {
		t.Fatalf("CreateWorklog: %v", err)
	}
	if resp.ID != "wl-1" || resp.TimeSpentSeconds != 3600 {
		t.Errorf("CreateWorklog response = %+v", resp)
	}
}
