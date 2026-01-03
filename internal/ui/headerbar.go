package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// HeaderBar is the application header bar.
type HeaderBar struct {
	*adw.HeaderBar

	// UI components
	toggleSidebarBtn *gtk.Button
	downloadButton   *gtk.Button
	settingsButton   *gtk.Button
	menuButton       *gtk.MenuButton

	// Callbacks
	onToggleSidebar func()
	onDownloadModel func()
	onChatSettings  func()
}

// NewHeaderBar creates a new header bar.
func NewHeaderBar() *HeaderBar {
	hb := &HeaderBar{}

	hb.HeaderBar = adw.NewHeaderBar()
	hb.SetShowStartTitleButtons(true)
	hb.SetShowEndTitleButtons(true)

	hb.setupUI()

	return hb
}

func (hb *HeaderBar) setupUI() {
	// Toggle sidebar button (start/left side)
	hb.toggleSidebarBtn = gtk.NewButton()
	hb.toggleSidebarBtn.SetIconName("sidebar-show-symbolic")
	hb.toggleSidebarBtn.SetTooltipText("Toggle Sidebar")
	hb.toggleSidebarBtn.ConnectClicked(func() {
		if hb.onToggleSidebar != nil {
			hb.onToggleSidebar()
		}
	})
	hb.PackStart(hb.toggleSidebarBtn)

	// Menu button (hamburger menu)
	hb.menuButton = gtk.NewMenuButton()
	hb.menuButton.SetIconName("open-menu-symbolic")
	hb.menuButton.SetTooltipText("Main Menu")
	hb.PackEnd(hb.menuButton)

	// Download model button
	hb.downloadButton = gtk.NewButton()
	hb.downloadButton.SetIconName("folder-download-symbolic")
	hb.downloadButton.SetTooltipText("Download Model")
	hb.downloadButton.ConnectClicked(func() {
		if hb.onDownloadModel != nil {
			hb.onDownloadModel()
		}
	})
	hb.PackEnd(hb.downloadButton)

	// Chat settings button (system prompt)
	hb.settingsButton = gtk.NewButton()
	hb.settingsButton.SetIconName("emblem-system-symbolic")
	hb.settingsButton.SetTooltipText("Chat Settings")
	hb.settingsButton.ConnectClicked(func() {
		if hb.onChatSettings != nil {
			hb.onChatSettings()
		}
	})
	hb.PackEnd(hb.settingsButton)
}

// OnDownloadModel sets the callback for when the download button is clicked.
func (hb *HeaderBar) OnDownloadModel(callback func()) {
	hb.onDownloadModel = callback
}

// OnChatSettings sets the callback for when the settings button is clicked.
func (hb *HeaderBar) OnChatSettings(callback func()) {
	hb.onChatSettings = callback
}

// OnToggleSidebar sets the callback for when the toggle sidebar button is clicked.
func (hb *HeaderBar) OnToggleSidebar(callback func()) {
	hb.onToggleSidebar = callback
}
