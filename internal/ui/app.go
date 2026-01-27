package ui

import (
	"log"
	"time"

	"fyne.io/fyne/v2"

	"jira-calendar/internal/config"
	"jira-calendar/internal/jira"
	"jira-calendar/internal/models"
)

// App represents the main application
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	config     models.JiraConfig
	jiraClient *jira.Client

	// State
	workLogs   []models.WorkLog
	daySummary map[string]*models.DayWorkSummary
	startDate  time.Time
	endDate    time.Time
	loading    bool
	errorMsg   string

	// User info
	userDisplayName string
	userEmail       string
	userTimeZone    string
	userAvatarURL   string
	userLocale      string
	userAccountType string
	userAccountID   string
	userActive      bool
	userSelfURL     string

	// UI components
	configForm    *fyne.Container
	dateForm      *fyne.Container
	calendarView  *fyne.Container
	mainContainer *fyne.Container
}

// NewApp creates a new application instance
func NewApp(fyneApp fyne.App) *App {
	log.Printf("[UI] Creating new Fyne app instance")

	window := fyneApp.NewWindow("Jira Work Calendar")
	// Set initial window size (resizable)
	window.Resize(fyne.NewSize(1200, 700))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	app := &App{
		fyneApp:    fyneApp,
		window:     window,
		daySummary: make(map[string]*models.DayWorkSummary),
	}

	// Load existing config
	cfg := config.Load()
	if cfg.BaseURL == "" {
		log.Printf("[UI] No existing config found, will show config form")
		app.config = models.JiraConfig{}
	} else {
		app.config = cfg
		log.Printf("[UI] Loaded config for: %s", cfg.Username)

		// Try to create Jira client
		client, err := jira.NewClient(cfg)
		if err != nil {
			log.Printf("[UI] Failed to create Jira client: %v", err)
			app.showConfigScreen()
		} else {
			app.jiraClient = client
			app.userDisplayName = client.DisplayName
			app.userEmail = client.EmailAddress
			app.userTimeZone = client.TimeZone
			app.userAvatarURL = client.AvatarURL
			app.userLocale = client.Locale
			app.userAccountType = client.AccountType
			app.userAccountID = client.AccountID
			app.userActive = client.Active
			app.userSelfURL = client.SelfURL
			log.Printf("[UI] Jira client created successfully")
			app.showDateScreen()
		}
	}

	// Show config screen if no client
	if app.jiraClient == nil {
		app.showConfigScreen()
	}

	return app
}

// ShowAndRun displays the window and runs the app
func (a *App) ShowAndRun() {
	a.window.ShowAndRun()
}
