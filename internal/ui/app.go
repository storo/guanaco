// Package ui provides the GTK4/Libadwaita user interface.
package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
)

const (
	// AppID is the application identifier.
	AppID = "com.github.storo.Guanaco"
)

// Application wraps the Adwaita application.
type Application struct {
	*adw.Application
	window *MainWindow
}

// NewApplication creates a new Guanaco application.
func NewApplication() *Application {
	app := &Application{}

	app.Application = adw.NewApplication(AppID, gio.ApplicationFlagsNone)
	app.ConnectActivate(app.onActivate)

	return app
}

// onActivate is called when the application is activated.
func (a *Application) onActivate() {
	// Create main window if it doesn't exist
	if a.window == nil {
		a.window = NewMainWindow(a.Application)
	}

	a.window.Present()
}

// Run starts the application.
func (a *Application) Run(args []string) int {
	return a.Application.Run(args)
}
