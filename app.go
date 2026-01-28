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

// App struct
type App struct {
	ctx        context.Context
	jiraClient *jira.Client
	config     models.JiraConfig
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	log.Println("[APP] Application started")
}

// OnDomReady is called when the frontend is ready
func (a *App) OnDomReady(ctx context.Context) {
	log.Println("[APP] DOM ready")
}

// OnBeforeClose is called before the app closes
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	log.Println("[APP] Application closing")
	return false
}

// OnShutdown is called when the app is shutting down
func (a *App) OnShutdown(ctx context.Context) {
	log.Println("[APP] Application shutdown")
}

// GetConfig returns the current Jira configuration
func (a *App) GetConfig() models.JiraConfig {
	cfg := config.Load()
	a.config = cfg
	return cfg
}

// SaveConfig saves the Jira configuration
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

// GetUserInfo returns the current user information
func (a *App) GetUserInfo() (map[string]interface{}, error) {
	if a.jiraClient == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	
	return a.jiraClient.GetUserInfo()
}

// GetWorkLogs fetches work logs for the given date range
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

// AddWorklog creates a new worklog entry for the given issue
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
