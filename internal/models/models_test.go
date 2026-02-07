package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWorklogResponse_GetCommentText_EmptyNil(t *testing.T) {
	var wr WorklogResponse
	if got := wr.GetCommentText(); got != "" {
		t.Errorf("GetCommentText() = %q, want \"\"", got)
	}
	wr.Comment = nil
	if got := wr.GetCommentText(); got != "" {
		t.Errorf("GetCommentText() with nil = %q, want \"\"", got)
	}
}

func TestWorklogResponse_GetCommentText_String(t *testing.T) {
	wr := WorklogResponse{Comment: "hello world"}
	if got := wr.GetCommentText(); got != "hello world" {
		t.Errorf("GetCommentText() = %q, want \"hello world\"", got)
	}
}

func TestWorklogResponse_GetCommentText_ADF(t *testing.T) {
	adf := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "first"},
					map[string]interface{}{"type": "text", "text": "second"},
				},
			},
		},
	}
	wr := WorklogResponse{Comment: adf}
	if got := wr.GetCommentText(); got != "first second" {
		t.Errorf("GetCommentText() ADF = %q, want \"first second\"", got)
	}
}

func TestWorklogResponse_GetCommentText_ADFMultipleParagraphs(t *testing.T) {
	adf := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "para1"},
				},
			},
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": "para2"},
				},
			},
		},
	}
	wr := WorklogResponse{Comment: adf}
	if got := wr.GetCommentText(); got != "para1 para2" {
		t.Errorf("GetCommentText() ADF = %q, want \"para1 para2\"", got)
	}
}

func TestJiraConfig_JSONRoundTrip(t *testing.T) {
	cfg := JiraConfig{
		BaseURL:  "https://example.atlassian.net",
		Username: "user@example.com",
		APIToken: "secret",
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded JiraConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.BaseURL != cfg.BaseURL || decoded.Username != cfg.Username || decoded.APIToken != cfg.APIToken {
		t.Errorf("round trip: got %+v", decoded)
	}
}

func TestWorklogRequest_JSONRoundTrip(t *testing.T) {
	req := WorklogRequest{
		Comment:          "did stuff",
		Started:          "2024-01-15T10:00:00.000+0000",
		TimeSpentSeconds: 3600,
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorklogRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Comment != req.Comment || decoded.Started != req.Started || decoded.TimeSpentSeconds != req.TimeSpentSeconds {
		t.Errorf("round trip: got %+v", decoded)
	}
}

func TestWorkLog_JSONRoundTrip(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	wl := WorkLog{
		ID:               "123",
		IssueKey:         "PROJ-1",
		IssueSummary:     "Summary",
		TimeSpent:        "1h",
		TimeSpentSeconds: 3600,
		Started:          now,
		Author: WorkLogAuthor{
			DisplayName:  "Dev",
			AccountID:    "abc",
			EmailAddress: "dev@example.com",
		},
		Comment: "comment",
	}
	data, err := json.Marshal(wl)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded WorkLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.ID != wl.ID || decoded.IssueKey != wl.IssueKey || decoded.Comment != wl.Comment {
		t.Errorf("round trip: got %+v", decoded)
	}
}
