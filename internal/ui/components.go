package ui

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createUserInfoCard creates a compact card displaying user information with avatar
func (a *App) createUserInfoCard() fyne.CanvasObject {
	// Avatar image
	var avatarImg *canvas.Image
	if a.userAvatarURL != "" {
		avatarImg = a.loadAvatarImage(a.userAvatarURL)
	} else {
		// Use default icon if no avatar
		avatarImg = canvas.NewImageFromResource(theme.AccountIcon())
		avatarImg.FillMode = canvas.ImageFillContain
	}
	avatarImg.SetMinSize(fyne.NewSize(48, 48))

	// User details
	displayName := a.userDisplayName
	if displayName == "" {
		displayName = "User"
	}

	// Header: Avatar + Username
	nameLabel := widget.NewLabelWithStyle(displayName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	header := container.NewHBox(
		avatarImg,
		container.NewPadded(nameLabel),
	)

	// Prepare values with fallbacks
	email := a.userEmail
	if email == "" {
		email = "N/A"
	}

	timezone := a.userTimeZone
	if timezone == "" {
		timezone = "N/A"
	}

	accountID := a.userAccountID
	if accountID == "" {
		accountID = "N/A"
	}

	locale := a.userLocale
	if locale == "" {
		locale = "N/A"
	}

	accountType := a.userAccountType
	if accountType == "" {
		accountType = "N/A"
	}

	activeStatus := "Active"
	if !a.userActive {
		activeStatus = "Inactive"
	}

	// Create info rows with icon, label, and data
	infoRows := container.NewVBox(
		createDetailRow("📧", "Email", email),
		createDetailRow("🆔", "Account ID", accountID),
		createDetailRow("👤", "Account Type", accountType),
		createDetailRow("✅", "Status", activeStatus),
		createDetailRow("🌍", "Timezone", timezone),
		createDetailRow("🌐", "Locale", locale),
	)

	// Main content
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		infoRows,
	)

	return widget.NewCard("User Information", "", content)
}

// loadAvatarImage loads an avatar image from URL
func (a *App) loadAvatarImage(url string) *canvas.Image {
	log.Printf("[UI] Loading avatar from: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("[UI] Failed to load avatar: %v", err)
		img := canvas.NewImageFromResource(theme.AccountIcon())
		img.FillMode = canvas.ImageFillContain
		return img
	}
	defer resp.Body.Close()

	imgData, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Printf("[UI] Failed to decode avatar: %v", err)
		img := canvas.NewImageFromResource(theme.AccountIcon())
		img.FillMode = canvas.ImageFillContain
		return img
	}

	img := canvas.NewImageFromImage(imgData)
	img.FillMode = canvas.ImageFillContain
	return img
}

// createDetailRow creates an info row with icon, label, and data
// Format: <icon> <label>: <data>
func createDetailRow(icon, label, data string) fyne.CanvasObject {
	iconLabel := widget.NewLabel(icon)

	labelText := widget.NewLabel(label + ":")
	labelText.TextStyle = fyne.TextStyle{Bold: true}

	dataText := widget.NewLabel(data)

	return container.NewHBox(
		iconLabel,
		labelText,
		dataText,
	)
}

// showDatePicker shows a calendar date selector with day grid
func (a *App) showDatePicker(entry *widget.Entry, defaultDate time.Time) {
	log.Printf("[UI] Showing date picker")

	// State variables
	var displayMonth time.Time = time.Date(defaultDate.Year(), defaultDate.Month(), 1, 0, 0, 0, 0, time.Local)

	// Dialog reference (will be set later)
	var d dialog.Dialog

	// Month and year selectors
	months := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	years := []string{}
	currentYear := time.Now().Year()
	for y := currentYear - 5; y <= currentYear+1; y++ {
		years = append(years, fmt.Sprintf("%d", y))
	}

	monthSelect := widget.NewSelect(months, nil)
	yearSelect := widget.NewSelect(years, nil)

	// Calendar grid container
	var calendarGrid *fyne.Container

	// Function to build calendar grid
	buildCalendarGrid := func() *fyne.Container {
		// Get first day of month and last day
		firstDay := time.Date(displayMonth.Year(), displayMonth.Month(), 1, 0, 0, 0, 0, time.Local)
		lastDay := firstDay.AddDate(0, 1, -1)

		// Day of week for first day (0 = Sunday)
		startWeekday := int(firstDay.Weekday())

		// Day headers
		dayHeaders := []fyne.CanvasObject{
			widget.NewLabelWithStyle("Sun", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Mon", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Tue", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Wed", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Thu", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Fri", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Sat", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		}

		// Create grid (6 rows x 7 days)
		var dayButtons []fyne.CanvasObject
		dayButtons = append(dayButtons, dayHeaders...)

		dayNum := 1
		for row := 0; row < 6; row++ {
			for col := 0; col < 7; col++ {
				cellIndex := row*7 + col

				// Empty cells before first day
				if cellIndex < startWeekday {
					dayButtons = append(dayButtons, widget.NewLabel(""))
					continue
				}

				// Empty cells after last day
				if dayNum > lastDay.Day() {
					dayButtons = append(dayButtons, widget.NewLabel(""))
					continue
				}

				// Create button for this day
				day := dayNum
				dayBtn := widget.NewButton(fmt.Sprintf("%d", day), func() {
					selected := time.Date(displayMonth.Year(), displayMonth.Month(), day, 0, 0, 0, 0, time.Local)
					entry.SetText(selected.Format("2006-01-02"))
					log.Printf("[UI] Date selected and confirmed: %s", selected.Format("2006-01-02"))
					// Close the dialog immediately
					if d != nil {
						d.Hide()
					}
				})

				// Highlight today
				today := time.Now()
				if displayMonth.Year() == today.Year() &&
					displayMonth.Month() == today.Month() &&
					day == today.Day() {
					dayBtn.Importance = widget.HighImportance
				}

				dayButtons = append(dayButtons, dayBtn)
				dayNum++
			}
		}

		return container.NewGridWithColumns(7, dayButtons...)
	}

	// Update display function
	updateDisplay := func() {
		monthIdx := 0
		for i, m := range months {
			if m == monthSelect.Selected {
				monthIdx = i + 1
				break
			}
		}
		year := currentYear
		fmt.Sscanf(yearSelect.Selected, "%d", &year)

		displayMonth = time.Date(year, time.Month(monthIdx), 1, 0, 0, 0, 0, time.Local)

		// Rebuild calendar grid
		newGrid := buildCalendarGrid()
		calendarGrid.Objects = newGrid.Objects
		calendarGrid.Refresh()

		log.Printf("[UI] Display updated to: %s %d", months[monthIdx-1], year)
	}

	// Set initial values
	monthSelect.SetSelected(months[displayMonth.Month()-1])
	yearSelect.SetSelected(fmt.Sprintf("%d", displayMonth.Year()))

	// Connect change handlers
	monthSelect.OnChanged = func(string) { updateDisplay() }
	yearSelect.OnChanged = func(string) { updateDisplay() }

	// Build initial calendar
	calendarGrid = buildCalendarGrid()

	// Quick selection buttons
	todayBtn := widget.NewButton("Today", func() {
		today := time.Now()
		entry.SetText(today.Format("2006-01-02"))
		log.Printf("[UI] Today selected: %s", today.Format("2006-01-02"))
		if d != nil {
			d.Hide()
		}
	})

	weekAgoBtn := widget.NewButton("1 Week Ago", func() {
		date := time.Now().AddDate(0, 0, -7)
		entry.SetText(date.Format("2006-01-02"))
		log.Printf("[UI] 1 Week Ago selected: %s", date.Format("2006-01-02"))
		if d != nil {
			d.Hide()
		}
	})

	// Content layout
	content := container.NewVBox(
		container.NewHBox(monthSelect, yearSelect),
		widget.NewSeparator(),
		calendarGrid,
		widget.NewSeparator(),
		container.NewHBox(todayBtn, weekAgoBtn),
	)

	// Create dialog with dismiss button
	d = dialog.NewCustom("Select Date", "Close", content, a.window)

	d.Resize(fyne.NewSize(450, 500))
	d.Show()
}
