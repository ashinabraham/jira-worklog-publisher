package ui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"jira-calendar/internal/models"
)

// showCalendarScreen displays the work log calendar
func (a *App) showCalendarScreen() {
	log.Printf("[UI] Showing calendar screen with %d days", len(a.daySummary))

	// Create back button
	backBtn := widget.NewButton("← Back to Date Selection", func() {
		a.showDateScreen()
	})

	// Create list of day summaries
	var summaryWidgets []fyne.CanvasObject

	// Sort dates
	var dates []string
	for date := range a.daySummary {
		dates = append(dates, date)
	}

	// Sort dates chronologically
	sortDates := func(dates []string) {
		for i := 0; i < len(dates); i++ {
			for j := i + 1; j < len(dates); j++ {
				if dates[i] > dates[j] {
					dates[i], dates[j] = dates[j], dates[i]
				}
			}
		}
	}
	sortDates(dates)

	// Create cards for each day
	for _, date := range dates {
		summary := a.daySummary[date]
		summaryWidgets = append(summaryWidgets, a.createDaySummaryCard(summary))
	}

	// Create scrollable content
	content := container.NewVBox(summaryWidgets...)
	scrollContainer := container.NewVScroll(content)

	a.calendarView = container.NewBorder(
		container.NewVBox(
			container.NewPadded(backBtn),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		scrollContainer,
	)

	a.window.SetContent(a.calendarView)
}

// createDaySummaryCard creates a card for a single day's work summary
func (a *App) createDaySummaryCard(summary *models.DayWorkSummary) fyne.CanvasObject {
	// Header with date and total hours
	dateLabel := widget.NewLabelWithStyle(
		summary.Date.Format("Monday, January 2, 2006"),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	hoursLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Total: %.2f hours", summary.TotalHours),
		fyne.TextAlignTrailing,
		fyne.TextStyle{Bold: true},
	)

	header := container.NewBorder(nil, nil, dateLabel, hoursLabel)

	// Tickets list
	var ticketWidgets []fyne.CanvasObject
	for _, ticket := range summary.Tickets {
		ticketLabel := widget.NewLabelWithStyle(
			fmt.Sprintf("%s: %s", ticket.IssueKey, ticket.Summary),
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		)

		hoursText := widget.NewLabel(fmt.Sprintf("  %.2f hours", ticket.Hours))

		// Comments
		var commentWidgets []fyne.CanvasObject
		for _, comment := range ticket.Comments {
			if comment != "" {
				commentLabel := widget.NewLabel(fmt.Sprintf("  • %s", comment))
				commentLabel.Wrapping = fyne.TextWrapWord
				commentWidgets = append(commentWidgets, commentLabel)
			}
		}

		ticketContainer := container.NewVBox(
			ticketLabel,
			hoursText,
		)

		if len(commentWidgets) > 0 {
			ticketContainer.Add(container.NewVBox(commentWidgets...))
		}

		ticketWidgets = append(ticketWidgets, ticketContainer)
		ticketWidgets = append(ticketWidgets, widget.NewSeparator())
	}

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		container.NewVBox(ticketWidgets...),
	)

	return widget.NewCard("", "", content)
}
