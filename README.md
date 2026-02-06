# JIRA Work Calendar

A desktop application for viewing Jira work logs in a calendar view and adding new work log entries. Built with [Wails v2](https://wails.io/) (Go backend + HTML/CSS/JS frontend).

## Features

- **Jira configuration** – Store Jira base URL, username, and API token (saved in `~/.jira-calendar-config.json`)
- **User info** – Display current user details (avatar, timezone, locale, etc.) from Jira
- **Date range calendar** – Select start/end dates with a calendar widget and load work logs
- **Work log calendar view** – Table of tickets with hours per date, weekend highlighting, and totals
- **Add work log** – Create work log entries (issue key, date, time, duration, comment) with inline error feedback
- **Refresh** – Reload calendar data from Jira; auto-refresh after adding a work log

## Requirements

- **Go** 1.24+
- **Wails CLI** – `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **macOS** – For building the `.app` bundle (Linux/Windows: use `wails build` for your platform)

## Quick Start

### Run in development

```bash
wails dev
```

### Build for production

```bash
# Current platform (e.g. macOS)
wails build

# macOS app bundle (output: build/bin/JIRA Work Calendar.app)
wails build -platform darwin

# Clean build
wails build -platform darwin -clean
```

### Install the app (macOS)

```bash
cp -R "build/bin/JIRA Work Calendar.app" /Applications/
```

## Project Structure

```
Jira_Calender/
├── main.go              # Wails entry point, window options, asset embedding
├── app.go               # Backend API exposed to frontend (GetConfig, GetWorkLogs, AddWorklog, etc.)
├── wails.json           # Wails config (name, icon, frontend dir)
├── appicon.icns         # macOS app icon
├── frontend/
│   └── dist/            # Frontend assets (served by Wails)
│       ├── index.html   # App shell and screens
│       ├── styles.css   # Dark theme and layout
│       └── main.js      # UI logic and Wails bindings
└── internal/
    ├── config/          # Load/save Jira config from home directory
    ├── jira/            # Jira REST API v3 client (auth, user, worklogs, create worklog)
    └── models/          # JiraConfig, WorkLog, WorklogRequest/Response, ADF comment handling
```

## Configuration

On first run, enter your Jira details:

- **Base URL** – e.g. `https://your-domain.atlassian.net`
- **Email** – Jira account email
- **API Token** – Create at [Atlassian API tokens](https://id.atlassian.com/manage-profile/security/api-tokens)

Config is stored at `~/.jira-calendar-config.json` (mode 0600).

## API Usage (Backend)

The app exposes these methods to the frontend via Wails bindings (see `app.go`):

| Method | Description |
|--------|-------------|
| `GetConfig()` | Load current Jira config |
| `SaveConfig(cfg)` | Save config and re-initialize Jira client |
| `GetUserInfo()` | Current user (displayName, avatar, timeZone, etc.) |
| `GetWorkLogs(startDate, endDate)` | Work logs in date range (YYYY-MM-DD) |
| `AddWorklog(issueKey, comment, started, timeSpentSeconds)` | Create a work log (comment ADF, started ISO 8601) |

## License

Private/Personal use.
