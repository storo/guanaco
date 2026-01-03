// Package ui provides the GTK4/Libadwaita user interface.
package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const styleCSS = `
/* === GUANACO MODERN UI STYLES === */

/* Message Bubbles - Base */
.message-bubble {
  margin: 4px 0;
}

/* User messages: compact pill */
.message-user .card {
  background: alpha(@card_fg_color, 0.12);
  border-radius: 18px;
  padding: 10px 16px;
}

/* System messages: subtle centered pill */
.message-system .card {
  background: alpha(@accent_bg_color, 0.1);
  border-radius: 12px;
  padding: 8px 14px;
  font-style: italic;
}

/* Input Area */
.input-area {
  background: @card_bg_color;
  border-radius: 16px;
  padding: 8px 12px 8px 12px;
}

.input-textview {
  background: transparent;
}

.input-scrolled {
  background: transparent;
}

/* Sidebar */
.navigation-sidebar row {
  border-radius: 8px;
  margin: 2px 6px 2px 6px;
}

.navigation-sidebar row:hover {
  background: alpha(@accent_bg_color, 0.08);
}

.navigation-sidebar row:selected {
  background: alpha(@accent_bg_color, 0.15);
}

/* Attachment Pill */
.attachment-pill {
  padding: 4px 8px 4px 8px;
  border-radius: 16px;
  background: alpha(@accent_bg_color, 0.15);
}

.attachment-pill:hover {
  background: alpha(@accent_bg_color, 0.25);
}

/* Code Blocks */
.code-block {
  background: #282a36;
  border-radius: 8px;
  margin: 4px 0;
}

.code-block-header {
  border-bottom: 1px solid alpha(@borders, 0.3);
}

.code-lang {
  font-size: 12px;
  opacity: 0.7;
  color: #f8f8f2;
}

.code-content {
  font-family: monospace;
  font-size: 13px;
  color: #f8f8f2;
  background: transparent;
}

.code-content text {
  background: transparent;
}

/* Welcome Screen */
.welcome-logo {
  margin-bottom: 16px;
  opacity: 0.9;
}

.suggestion-pill {
  background-color: alpha(@card_bg_color, 0.5);
  border-radius: 20px;
  padding: 8px 16px;
}

.suggestion-pill:hover {
  background-color: alpha(@card_bg_color, 0.8);
}
`

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
	// Load custom CSS
	loadCSS()

	// Create main window if it doesn't exist
	if a.window == nil {
		a.window = NewMainWindow(a.Application)
	}

	a.window.Present()
}

// loadCSS loads the application stylesheet.
func loadCSS() {
	provider := gtk.NewCSSProvider()
	provider.LoadFromData(styleCSS)

	display := gdk.DisplayGetDefault()
	gtk.StyleContextAddProviderForDisplay(display, provider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

// Run starts the application.
func (a *Application) Run(args []string) int {
	return a.Application.Run(args)
}
