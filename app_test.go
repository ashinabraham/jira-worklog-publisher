package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jira-worklog-publisher/internal/config"
	"jira-worklog-publisher/internal/models"
)

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp() returned nil")
	}
	if a.jiraClient != nil {
		t.Error("new App should have nil jiraClient")
	}
}

func TestGetConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := `{"baseURL":"https://app-test.atlassian.net","username":"app@test.com","apiToken":"t"}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	defer config.SetConfigPathForTest(func() string { return path })()

	app := NewApp()
	cfg := app.GetConfig()
	if cfg.BaseURL != "https://app-test.atlassian.net" || cfg.Username != "app@test.com" {
		t.Errorf("GetConfig() = %+v", cfg)
	}
}

func TestGetUserInfo_NotAuthenticated(t *testing.T) {
	app := NewApp()
	_, err := app.GetUserInfo()
	if err == nil || err.Error() != "not authenticated" {
		t.Errorf("GetUserInfo() = %v, want not authenticated error", err)
	}
}

func TestGetWorkLogs_NotAuthenticated(t *testing.T) {
	app := NewApp()
	_, err := app.GetWorkLogs("2024-01-01", "2024-01-31")
	if err == nil || err.Error() != "not authenticated" {
		t.Errorf("GetWorkLogs() = %v, want not authenticated error", err)
	}
}

func TestGetWorkLogs_InvalidStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"x","displayName":"","emailAddress":"","timeZone":""}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	defer config.SetConfigPathForTest(func() string { return filepath.Join(dir, "config.json") })()

	app := NewApp()
	if err := app.SaveConfig(models.JiraConfig{BaseURL: server.URL, Username: "u", APIToken: "t"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	_, err := app.GetWorkLogs("not-a-date", "2024-01-31")
	if err == nil {
		t.Fatal("GetWorkLogs() expected error for invalid start date")
	}
	if !strings.HasPrefix(err.Error(), "invalid start date") {
		t.Errorf("GetWorkLogs() = %v", err)
	}
}

func TestGetWorkLogs_InvalidEndDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"x","displayName":"","emailAddress":"","timeZone":""}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	defer config.SetConfigPathForTest(func() string { return filepath.Join(dir, "config.json") })()

	app := NewApp()
	if err := app.SaveConfig(models.JiraConfig{BaseURL: server.URL, Username: "u", APIToken: "t"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	_, err := app.GetWorkLogs("2024-01-01", "invalid")
	if err == nil {
		t.Fatal("GetWorkLogs() expected error for invalid end date")
	}
	if !strings.HasPrefix(err.Error(), "invalid end date") {
		t.Errorf("GetWorkLogs() = %v", err)
	}
}

func TestAddWorklog_NotAuthenticated(t *testing.T) {
	app := NewApp()
	_, err := app.AddWorklog("PROJ-1", "comment", "2024-01-15T10:00:00.000+0000", 3600)
	if err == nil || err.Error() != "not authenticated" {
		t.Errorf("AddWorklog() = %v, want not authenticated error", err)
	}
}

func TestSaveConfig_ThenGetUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"accountId":"acc-1","displayName":"Test User","emailAddress":"test@example.com","timeZone":"UTC"}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	defer config.SetConfigPathForTest(func() string { return filepath.Join(dir, "config.json") })()

	app := NewApp()
	cfg := models.JiraConfig{
		BaseURL:  server.URL,
		Username: "user",
		APIToken: "token",
	}
	if err := app.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	info, err := app.GetUserInfo()
	if err != nil {
		t.Fatalf("GetUserInfo: %v", err)
	}
	if info["displayName"] != "Test User" || info["accountID"] != "acc-1" {
		t.Errorf("GetUserInfo = %+v", info)
	}
}
