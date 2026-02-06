// Package main is the entry point for the JIRA Worklog Publisher desktop application.
// It initializes the Wails runtime, embeds frontend assets, and runs the app window.
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// assets embeds the frontend/dist directory (HTML, CSS, JS) into the binary.
//
//go:embed all:frontend/dist
var assets embed.FS

// main starts the Wails application: creates the App backend, configures the window
// (fixed 1200x700), and binds the App to the frontend.
func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "JIRA Worklog Publisher",
		Width:     1200,
		Height:    700,
		MinWidth:  1200,
		MinHeight: 700,
		MaxWidth:  1200,
		MaxHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnDomReady:       app.OnDomReady,
		OnBeforeClose:    app.OnBeforeClose,
		OnShutdown:       app.OnShutdown,
		// Expose app methods to frontend
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
