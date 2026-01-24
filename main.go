package main

import (
	"log"

	"fyne.io/fyne/v2/app"

	"jira-calendar/internal/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[MAIN] Starting Jira Work Calendar")

	// Create Fyne application
	myApp := app.NewWithID("com.jiracalendar.app")

	// Create and run UI
	appInstance := ui.NewApp(myApp)
	appInstance.ShowAndRun()

	log.Println("[MAIN] Application closed")
}
