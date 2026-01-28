# Wails Implementation Setup

This branch (`wails-implementation`) contains the Wails-based rewrite of the Jira Calendar application.

## Changes Made

### ✅ Removed
- **Fyne UI code** - All `internal/ui/*` files removed
- **Fyne dependency** - Removed from `go.mod`
- **go-jira dependency** - Removed from `go.mod` (unmaintained)

### ✅ Created
- **Custom Jira SDK** - `internal/jira/client.go` rewritten without go-jira
  - Direct HTTP calls to Jira REST API v3
  - Basic authentication
  - User info fetching
  - JQL search
  - Worklog retrieval
- **Wails App Structure**
  - `main.go` - Wails application entry point
  - `app.go` - Backend methods exposed to frontend
- **Frontend Structure**
  - `frontend/dist/index.html` - Main HTML
  - `frontend/dist/styles.css` - Dark theme styles
  - `frontend/dist/main.js` - Frontend JavaScript
- **Wails Config** - `wails.json`

## Project Structure

```
jira-calendar/
├── app.go                 # Wails backend (exposed methods)
├── main.go                # Wails entry point
├── wails.json             # Wails configuration
├── frontend/
│   └── dist/              # Frontend assets
│       ├── index.html
│       ├── styles.css
│       └── main.js
└── internal/
    ├── config/            # Config management (unchanged)
    ├── jira/              # Custom Jira SDK (rewritten)
    └── models/            # Data models (unchanged)
```

## Custom Jira SDK Features

The new `internal/jira/client.go` implements:
- ✅ Basic authentication (username + API token)
- ✅ User info retrieval (`/rest/api/3/myself`)
- ✅ JQL search (`/rest/api/3/search/jql`)
- ✅ Worklog fetching (`/rest/api/3/issue/{id}/worklog`)
- ✅ Pagination support
- ✅ Date range filtering
- ✅ Work log aggregation

## Wails Backend Methods

Exposed to frontend via `app.go`:
- `GetConfig()` - Load Jira configuration
- `SaveConfig(config)` - Save and authenticate
- `GetUserInfo()` - Get current user details
- `GetWorkLogs(startDate, endDate)` - Fetch work logs

## Next Steps

1. **Build the app**: `wails build` or `go build`
2. **Run in dev mode**: `wails dev` (requires frontend build setup)
3. **Implement calendar table** in `frontend/dist/main.js`
4. **Add date picker** UI components
5. **Style the calendar** with proper table layout

## Dependencies

- **Wails v2** - Desktop app framework
- **Standard library only** - No external Jira dependencies

## Building

```bash
# Development
wails dev

# Production build
wails build
```
