package ui

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// showCalendarScreen displays the work log calendar in table format
func (a *App) showCalendarScreen() {
	log.Printf("[UI] Showing calendar screen with %d days", len(a.daySummary))

	// Create header buttons
	addWorklogBtn := widget.NewButton("Add Worklog", func() {
		// TODO: Implement add worklog functionality
		log.Println("[UI] Add Worklog clicked")
	})
	addWorklogBtn.Importance = widget.HighImportance

	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		a.showDateScreen()
	})
	closeBtn.Importance = widget.DangerImportance

	header := container.NewBorder(
		nil, nil, nil,
		container.NewHBox(addWorklogBtn, closeBtn),
	)

	// Build table data
	tableData := a.buildTableData()

	// Check if no data
	if len(tableData.tickets) == 0 {
		noDataLabel := widget.NewLabel("No work logs found for the selected date range.")
		noDataLabel.Alignment = fyne.TextAlignCenter

		a.calendarView = container.NewBorder(
			container.NewPadded(header),
			nil, nil, nil,
			container.NewCenter(noDataLabel),
		)
		a.window.SetContent(a.calendarView)
		return
	}

	// Create table (it has built-in scrolling)
	table := a.createWorklogTable(tableData)

	a.calendarView = container.NewBorder(
		container.NewPadded(header),
		nil, nil, nil,
		table, // Table widget has built-in scrolling
	)

	a.window.SetContent(a.calendarView)
}

// tableData holds the structured data for the table view
type tableData struct {
	dates        []time.Time
	tickets      []string
	ticketNames  map[string]string
	hours        map[string]map[string]float64 // ticket -> date -> hours
	dailyTotals  map[string]float64            // date -> total hours
	ticketTotals map[string]float64            // ticket -> total hours
	grandTotal   float64
}

// buildTableData processes work logs into table structure
func (a *App) buildTableData() *tableData {
	data := &tableData{
		ticketNames:  make(map[string]string),
		hours:        make(map[string]map[string]float64),
		dailyTotals:  make(map[string]float64),
		ticketTotals: make(map[string]float64),
	}

	// Generate all dates in the range
	currentDate := a.startDate
	for !currentDate.After(a.endDate) {
		data.dates = append(data.dates, currentDate)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Collect all unique tickets
	ticketSet := make(map[string]bool)
	for _, summary := range a.daySummary {
		for _, ticket := range summary.Tickets {
			if !ticketSet[ticket.IssueKey] {
				ticketSet[ticket.IssueKey] = true
				data.tickets = append(data.tickets, ticket.IssueKey)
				data.ticketNames[ticket.IssueKey] = ticket.Summary
			}
		}
	}

	// Build hours matrix and track earliest date per ticket
	ticketEarliestDate := make(map[string]time.Time)
	
	for dateStr, summary := range a.daySummary {
		for _, ticket := range summary.Tickets {
			if data.hours[ticket.IssueKey] == nil {
				data.hours[ticket.IssueKey] = make(map[string]float64)
			}
			data.hours[ticket.IssueKey][dateStr] = ticket.Hours
			data.dailyTotals[dateStr] += ticket.Hours
			data.ticketTotals[ticket.IssueKey] += ticket.Hours
			data.grandTotal += ticket.Hours
			
			// Track earliest date for this ticket
			if earliest, exists := ticketEarliestDate[ticket.IssueKey]; !exists || summary.Date.Before(earliest) {
				ticketEarliestDate[ticket.IssueKey] = summary.Date
			}
		}
	}

	// Sort tickets by earliest worklog date (oldest first)
	for i := 0; i < len(data.tickets); i++ {
		for j := i + 1; j < len(data.tickets); j++ {
			ticket1 := data.tickets[i]
			ticket2 := data.tickets[j]
			
			date1, exists1 := ticketEarliestDate[ticket1]
			date2, exists2 := ticketEarliestDate[ticket2]
			
			// If both have dates, compare them
			if exists1 && exists2 && date1.After(date2) {
				data.tickets[i], data.tickets[j] = data.tickets[j], data.tickets[i]
			}
		}
	}

	return data
}

// createWorklogTable creates a simple, clean scrollable table
func (a *App) createWorklogTable(data *tableData) fyne.CanvasObject {
	numCols := len(data.dates) + 2   // Ticket + dates + total
	numRows := len(data.tickets) + 2 // Header + tickets + total row

	log.Printf("[UI] Creating table: %d rows x %d cols", numRows, numCols)

	// Single unified table - simpler and more reliable
	table := widget.NewTable(
		// Size
		func() (int, int) {
			return numRows, numCols
		},
		// Create template with background for weekend highlighting
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.RGBA{R: 220, G: 230, B: 240, A: 255})
			label := widget.NewLabel("")
			return container.NewStack(bg, container.NewPadded(label))
		},
		// Update cell
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			stack := cell.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			labelContainer := stack.Objects[1].(*fyne.Container)
			label := labelContainer.Objects[0].(*widget.Label)

			row, col := id.Row, id.Col

			// Reset
			label.TextStyle = fyne.TextStyle{}
			label.Alignment = fyne.TextAlignCenter
			label.Truncation = fyne.TextTruncateEllipsis
			bg.FillColor = color.Transparent

			// Check if this is a weekend column
			isWeekend := false
			if col > 0 && col < numCols-1 {
				idx := col - 1
				if idx < len(data.dates) {
					weekday := data.dates[idx].Weekday()
					isWeekend = (weekday == time.Saturday || weekday == time.Sunday)
					if isWeekend {
						// Light blue/gray for weekends
						bg.FillColor = color.RGBA{R: 146, G: 38, B: 38, A: 255}
					}
				}
			}
			bg.Refresh()

			// Header row
			if row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				if col == 0 {
					label.Alignment = fyne.TextAlignLeading
					label.SetText("Ticket")
				} else if col == numCols-1 {
					label.SetText("Total")
				} else {
					idx := col - 1
					if idx < len(data.dates) {
						date := data.dates[idx]
						label.SetText(fmt.Sprintf("%02d\n%s", date.Day(), date.Format("Mon")))
					}
				}
				return
			}

			// Total row
			if row == numRows-1 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				if col == 0 {
					label.Alignment = fyne.TextAlignLeading
					label.SetText("Total Hours")
				} else if col == numCols-1 {
					label.SetText(fmt.Sprintf("%.0fh", data.grandTotal))
				} else {
					idx := col - 1
					if idx < len(data.dates) {
						dateStr := data.dates[idx].Format("2006-01-02")
						if total := data.dailyTotals[dateStr]; total > 0 {
							label.SetText(fmt.Sprintf("%.0fh", total))
						} else {
							label.SetText("")
						}
					}
				}
				return
			}

			// Data rows
			ticketIdx := row - 1
			if ticketIdx < 0 || ticketIdx >= len(data.tickets) {
				return
			}

			ticket := data.tickets[ticketIdx]

			if col == 0 {
				// Ticket column
				label.Alignment = fyne.TextAlignLeading
				ticketName := data.ticketNames[ticket]
				label.SetText(fmt.Sprintf("%s - %s", ticket, ticketName))
			} else if col == numCols-1 {
				// Total column
				label.TextStyle = fyne.TextStyle{Bold: true}
				if hours := data.ticketTotals[ticket]; hours > 0 {
					label.SetText(fmt.Sprintf("%.0fh", hours))
				} else {
					label.SetText("")
				}
			} else {
				// Date columns
				idx := col - 1
				if idx < len(data.dates) {
					dateStr := data.dates[idx].Format("2006-01-02")
					if hours := data.hours[ticket][dateStr]; hours > 0 {
						label.SetText(fmt.Sprintf("%.0fh", hours))
					} else {
						label.SetText("")
					}
				}
			}
		},
	)

	// Set column widths
	table.SetColumnWidth(0, 400) // Ticket
	for i := 1; i < numCols-1; i++ {
		table.SetColumnWidth(i, 100) // Dates
	}
	table.SetColumnWidth(numCols-1, 100) // Total

	// Set row heights for better alignment
	for i := 0; i < numRows; i++ {
		if i == 0 {
			table.SetRowHeight(i, 70) // Header row - taller for date labels
		} else {
			table.SetRowHeight(i, 60) // Data rows - consistent height
		}
	}

	// Make ticket column sticky (stays visible when scrolling horizontally)
	table.StickyColumnCount = 1
	table.StickyRowCount = 1

	return table
}
