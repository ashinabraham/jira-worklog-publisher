// Package main provides the Wails backend: the App type and its methods
// are bound to the frontend and handle config, user info, work logs, and adding work logs.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"jira-calendar/internal/config"
	"jira-calendar/internal/jira"
	"jira-calendar/internal/models"
)

// App holds the Wails app context and the Jira client used by bound methods.
type App struct {
	ctx        context.Context
	jiraClient *jira.Client
	config     models.JiraConfig
}

// NewApp returns a new App instance (Jira client is nil until SaveConfig is called).
func NewApp() *App {
	return &App{}
}

// OnStartup is called by Wails when the app starts; the context is stored for later use.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	log.Println("[APP] Application started")
}

// OnDomReady is called by Wails when the frontend DOM is ready.
func (a *App) OnDomReady(ctx context.Context) {
	log.Println("[APP] DOM ready")
}

// OnBeforeClose is called by Wails before the window closes. Return true to prevent closing.
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	log.Println("[APP] Application closing")
	return false
}

// OnShutdown is called by Wails when the application is shutting down.
func (a *App) OnShutdown(ctx context.Context) {
	log.Println("[APP] Application shutdown")
}

// GetConfig returns the current Jira configuration from the config file.
func (a *App) GetConfig() models.JiraConfig {
	cfg := config.Load()
	a.config = cfg
	return cfg
}

// SaveConfig saves the Jira configuration to disk and initializes the Jira client for API calls.
func (a *App) SaveConfig(cfg models.JiraConfig) error {
	if err := config.Save(cfg); err != nil {
		return err
	}
	a.config = cfg
	
	// Create basic auth and Jira client
	auth := jira.NewBasicAuth(cfg.Username, cfg.APIToken)
	client, err := jira.NewClient(cfg.BaseURL, auth)
	if err != nil {
		return err
	}
	a.jiraClient = client
	
	return nil
}

// GetUserInfo returns the current Jira user (displayName, avatarURL, timeZone, locale, etc.).
// Requires prior successful SaveConfig.
func (a *App) GetUserInfo() (map[string]interface{}, error) {
	if a.jiraClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	return a.jiraClient.GetUserInfo()
}

// GetWorkLogs fetches work logs for the given date range. Dates must be in YYYY-MM-DD format.
func (a *App) GetWorkLogs(startDateStr, endDateStr string) ([]models.WorkLog, error) {
	if a.jiraClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %v", err)
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %v", err)
	}
	
	return a.jiraClient.GetWorkLogs(startDate, endDate)
}

// AddWorklog creates a new work log for the given issue. started should be ISO 8601 (e.g. 2024-01-15T10:00:00.000+0000).
func (a *App) AddWorklog(issueKey string, comment string, started string, timeSpentSeconds int) (*models.WorklogResponse, error) {
	if a.jiraClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	req := models.WorklogRequest{
		Comment:          comment,
		Started:          started,
		TimeSpentSeconds: timeSpentSeconds,
	}
	
	return a.jiraClient.CreateWorklog(issueKey, req)
}
