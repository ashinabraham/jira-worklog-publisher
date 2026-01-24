package ui

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"jira-calendar/internal/config"
	"jira-calendar/internal/jira"
	"jira-calendar/internal/models"
)

// showConfigScreen displays the configuration form
func (a *App) showConfigScreen() {
	log.Printf("[UI] Showing config screen")

	// Create form entries
	baseURLEntry := widget.NewEntry()
	baseURLEntry.SetPlaceHolder("https://your-domain.atlassian.net")
	baseURLEntry.SetText(a.config.BaseURL)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("your-email@example.com")
	usernameEntry.SetText(a.config.Username)

	apiTokenEntry := widget.NewPasswordEntry()
	apiTokenEntry.SetPlaceHolder("Your Jira API Token")
	apiTokenEntry.SetText(a.config.APIToken)

	// Create invisible spacer to force minimum width
	widthSpacer := canvas.NewRectangle(color.Transparent)
	widthSpacer.SetMinSize(fyne.NewSize(600, 1)) // 600px wide, 1px tall

	// Create form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Jira Base URL", Widget: baseURLEntry},
			{Text: "Username/Email", Widget: usernameEntry},
			{Text: "API Token", Widget: apiTokenEntry},
		},
		OnSubmit: func() {
			a.handleConfigSave(baseURLEntry.Text, usernameEntry.Text, apiTokenEntry.Text)
		},
	}

	// Create close button for top right (only show if jiraClient exists)
	var header *fyne.Container
	if a.jiraClient != nil {
		closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
			// Go back to date screen
			a.showDateScreen()
		})
		closeBtn.Importance = widget.DangerImportance

		header = container.NewBorder(
			nil, nil,
			nil, closeBtn,
		)
	} else {
		// No close button for first-time setup
		header = container.NewBorder(
			nil, nil,
			nil, nil,
		)
	}

	// Create instruction text
	instructions := widget.NewLabel("Enter your Jira credentials to get started.\n" +
		"You can generate an API token at:\n" +
		"https://id.atlassian.com/manage-profile/security/api-tokens")
	instructions.Wrapping = fyne.TextWrapWord
	
	// Create instruction card with width spacer
	instructionCard := widget.NewCard("Jira Configuration", "",
		container.NewVBox(
			widthSpacer, // Force minimum width
			container.NewPadded(instructions),
		),
	)
	
	// Wrap form in a card
	formCard := widget.NewCard("", "",
		container.NewPadded(form),
	)

	// Group both cards
	cardContent := container.NewVBox(
		instructionCard,
		widget.NewSeparator(),
		formCard,
	)

	// Center the content
	centeredContent := container.NewCenter(
		container.NewPadded(cardContent),
	)

	// Simple border layout with header at top
	a.configForm = container.NewBorder(
		container.NewPadded(header), // header with close button at top
		nil, nil, nil,
		centeredContent, // center content that adjusts to window
	)

	a.window.SetContent(a.configForm)
}

// handleConfigSave processes the configuration form submission
func (a *App) handleConfigSave(baseURL, username, apiToken string) {
	log.Printf("[UI] Saving config")

	if baseURL == "" || username == "" || apiToken == "" {
		dialog.ShowError(fmt.Errorf("all fields are required"), a.window)
		return
	}

	a.config = models.JiraConfig{
		BaseURL:  baseURL,
		Username: username,
		APIToken: apiToken,
	}

	// Save config
	if err := config.Save(a.config); err != nil {
		dialog.ShowError(fmt.Errorf("failed to save config: %v", err), a.window)
		return
	}

	// Create Jira client
	client, err := jira.NewClient(a.config)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to connect to Jira: %v", err), a.window)
		return
	}

	a.jiraClient = client
	a.userDisplayName = client.DisplayName
	a.userEmail = client.EmailAddress
	a.userTimeZone = client.TimeZone
	a.userAvatarURL = client.AvatarURL
	a.userLocale = client.Locale
	a.userAccountType = client.AccountType
	a.userAccountID = client.AccountID
	a.userActive = client.Active
	a.userSelfURL = client.SelfURL

	log.Printf("[UI] Config saved and client created successfully")
	dialog.ShowInformation("Success", "Connected to Jira successfully!", a.window)

	a.showDateScreen()
}
