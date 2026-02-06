# Wails Implementation – Developer Guide

This document describes the Wails-based JIRA Work Calendar app for developers.

## Architecture

- **Backend (Go)**: `main.go` (entry + embed), `app.go` (bound API), `internal/config`, `internal/jira`, `internal/models`.
- **Frontend**: Static assets in `frontend/dist/` (HTML, CSS, JS). No build step; files are embedded and served by Wails.
- **Bridge**: Wails generates bindings so the frontend can call `window.go.main.App` methods (GetConfig, SaveConfig, GetUserInfo, GetWorkLogs, AddWorklog).

## Project layout

```
jira-calendar/
├── main.go                 # Wails Run(), window options, embed frontend/dist
├── app.go                  # App struct + methods bound to frontend
├── wails.json              # App name, icon, frontend dir
├── appicon.icns            # macOS app icon
├── frontend/
│   └── dist/
│       ├── index.html      # Screens: config, date (home), calendar, add-worklog modal
│       ├── styles.css      # Dark theme, layout, widgets
│       └── main.js         # initApp, screens, forms, calendar table, worklog form
└── internal/
    ├── config/             # Load/Save Jira config (~/.jira-calendar-config.json)
    ├── jira/               # REST API v3 client, BasicAuth, GetUserInfo, GetWorkLogs, CreateWorklog
    └── models/             # JiraConfig, WorkLog, WorklogRequest/Response, ADF comment
```

## Backend (Go)

### Entry and app lifecycle

- `main.go`: embeds `frontend/dist`, creates `App`, runs Wails with fixed 1200x700 window.
- `app.go`: `OnStartup`, `OnDomReady`, `OnBeforeClose`, `OnShutdown`; plus bound methods below.

### Bound methods (see app.go)

| Method | Purpose |
|--------|--------|
| `GetConfig()` | Load config from disk. |
| `SaveConfig(cfg)` | Save config and create Jira client (BasicAuth). |
| `GetUserInfo()` | Current user from `/rest/api/3/myself`. |
| `GetWorkLogs(start, end)` | Work logs for date range (JQL + worklog API). |
| `AddWorklog(issueKey, comment, started, timeSpentSeconds)` | Create work log (comment sent as ADF). |

### Jira client (internal/jira)

- **Authenticator**: interface (e.g. BasicAuth with username + API token).
- **Client**: holds baseURL, auth, AccountID. NewClient() validates auth via `/rest/api/3/myself`.
- **GetWorkLogs**: JQL for issues with worklogs in range, then fetches worklogs per issue, filters by AccountID.
- **CreateWorklog**: POST to `/rest/api/3/issue/{key}/worklog`; comment body is ADF. Parses Jira error response for user-facing messages.

### Config and models

- **config**: Load/Save `models.JiraConfig` to `~/.jira-calendar-config.json` (mode 0600).
- **models**: JiraConfig, WorkLog, WorkLogAuthor, WorklogRequest, WorklogResponse, ADF comment handling (GetCommentText).

## Frontend (frontend/dist)

- **index.html**: Three main screens (config, date, calendar) + add-worklog modal; structure for calendar widget and date picker.
- **main.js**: 
  - Init: `DOMContentLoaded` → wait for `window.go.main.App` → `initApp()` → checkConfig, setup forms and buttons.
  - Screens: showConfigScreen, showDateScreen, showCalendarScreen; calendar state and current date range for refresh.
  - Date form: calendar widget (start/end), Load Calendar → GetWorkLogs → showCalendarScreen.
  - Calendar view: renderCalendar (table), refresh button (refreshCalendar), Add Worklog → showAddWorklogModal.
  - Add worklog: form with issue key, date picker, time, duration, comment; submit → AddWorklog, then hide modal and refreshCalendar; errors shown in form.
- **styles.css**: Layout, dark theme, calendar widget, date picker, modal, error message, refresh button (including refreshing state).

## Building and running

```bash
# Development (live reload if configured)
wails dev

# Production build (current platform)
wails build

# macOS app bundle
wails build -platform darwin -clean
# Output: build/bin/JIRA Work Calendar.app
```

## Dependencies

- **Wails v2** – desktop framework.
- **Standard library + Wails** – no other Jira libraries; Jira REST API v3 used directly.

## Extending

- **Auth**: Implement `Authenticator` and use it in `NewClient`; wire from config in `SaveConfig`.
- **UI**: Edit `frontend/dist/*`; no separate frontend build step.
- **New backend methods**: Add to `App` in `app.go`; they become available on `window.go.main.App` in JS.
