package ui

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"jira-calendar/internal/jira"
)

// showDateScreen displays the date selection form
func (a *App) showDateScreen() {
	log.Printf("[UI] Showing date screen")

	// Default dates
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()

	// Create date entries with minimum width
	startDateEntry := widget.NewEntry()
	startDateEntry.SetText(startDate.Format("2006-01-02"))
	startDateEntry.SetPlaceHolder("YYYY-MM-DD")

	endDateEntry := widget.NewEntry()
	endDateEntry.SetText(endDate.Format("2006-01-02"))
	endDateEntry.SetPlaceHolder("YYYY-MM-DD")

	// Create date picker buttons
	startPickerBtn := widget.NewButton("📅 Pick Date", func() {
		a.showDatePicker(startDateEntry, startDate)
	})

	endPickerBtn := widget.NewButton("📅 Pick Date", func() {
		a.showDatePicker(endDateEntry, endDate)
	})

	// Create form with date pickers - entries will expand to fill available space
	startDateContainer := container.NewBorder(nil, nil, nil, startPickerBtn, startDateEntry)
	endDateContainer := container.NewBorder(nil, nil, nil, endPickerBtn, endDateEntry)

	// Use Fyne's built-in form for proper alignment
	dateForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Start Date", Widget: startDateContainer},
			{Text: "End Date", Widget: endDateContainer},
		},
	}

	// Load Calendar button (separate, on the left)
	loadBtn := widget.NewButton("Load Calendar", func() {
		a.handleLoadCalendar(startDateEntry.Text, endDateEntry.Text)
	})
	loadBtn.Importance = widget.HighImportance

	// User info card
	userCard := a.createUserInfoCard()

	// Create card with generous internal padding
	formCard := widget.NewCard("Select Date Range", "",
		container.NewPadded(
			container.NewPadded(dateForm),
		),
	)

	// Create form section
	formSection := container.NewVBox(
		formCard,
		container.NewPadded(loadBtn),
	)

	// Wrap both cards in a max container to ensure same width
	// Both will have the same max width and stay aligned
	cardContainer := container.NewVBox(
		container.NewPadded(userCard),
		container.NewPadded(widget.NewSeparator()),
		container.NewPadded(formSection),
	)

	// Center the card container as a whole - keeps cards aligned and same width
	centeredCards := container.NewCenter(
		container.NewPadded(cardContainer),
	)

	// Create settings button
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		a.showConfigScreen()
	})

	// Create title label with larger font using canvas.Text for size control
	titleText := canvas.NewText("Jira Work Calendar", theme.ForegroundColor())
	titleText.TextSize = 24 // Larger font size
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Alignment = fyne.TextAlignCenter

	// Layout - title centered, settings button on right
	header := container.NewBorder(
		nil, nil,
		nil,
		settingsBtn,
		container.NewCenter(titleText),
	)

	// Main content with centered card container
	mainContent := centeredCards

	a.dateForm = container.NewBorder(
		header,
		nil, nil, nil,
		container.NewPadded(
			container.NewPadded(
				container.NewPadded(mainContent),
			),
		),
	)

	a.window.SetContent(a.dateForm)
}

// handleLoadCalendar processes the calendar loading request
func (a *App) handleLoadCalendar(startDateStr, endDateStr string) {
	log.Printf("[UI] Load calendar requested: %s to %s", startDateStr, endDateStr)

	// Validate dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid start date format. Use YYYY-MM-DD"), a.window)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		dialog.ShowError(fmt.Errorf("invalid end date format. Use YYYY-MM-DD"), a.window)
		return
	}

	if endDate.Before(startDate) {
		dialog.ShowError(fmt.Errorf("end date must be after start date"), a.window)
		return
	}

	// Show loading dialog
	progress := dialog.NewProgressInfinite("Loading", "Fetching work logs from Jira...", a.window)
	progress.Show()

	// Fetch work logs in goroutine
	go func() {
		workLogs, err := a.jiraClient.GetWorkLogs(startDate, endDate)

		// Update UI on main thread
		progress.Hide()

		if err != nil {
			log.Printf("[UI] Failed to fetch work logs: %v", err)
			dialog.ShowError(fmt.Errorf("failed to fetch work logs: %v", err), a.window)
			return
		}

		log.Printf("[UI] Fetched %d work logs", len(workLogs))
		a.workLogs = workLogs

		// Aggregate by date
		a.daySummary = jira.AggregateWorkLogs(workLogs)

		if len(a.daySummary) == 0 {
			dialog.ShowInformation("No Data", "No work logs found for the selected date range.", a.window)
			return
		}

		a.showCalendarScreen()
	}()
}
