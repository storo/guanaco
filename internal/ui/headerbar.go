package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/i18n"
)

// HeaderBar is the application header bar.
type HeaderBar struct {
	*adw.HeaderBar

	// UI components
	toggleSidebarBtn *gtk.Button
	downloadButton   *gtk.Button
	settingsButton   *gtk.Button

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
	hb.toggleSidebarBtn.SetTooltipText(i18n.T("Toggle Sidebar"))
	hb.toggleSidebarBtn.ConnectClicked(func() {
		if hb.onToggleSidebar != nil {
			hb.onToggleSidebar()
		}
	})
	hb.PackStart(hb.toggleSidebarBtn)

	// Download model button
	hb.downloadButton = gtk.NewButton()
	hb.downloadButton.SetIconName("folder-download-symbolic")
	hb.downloadButton.SetTooltipText(i18n.T("Download Model"))
	hb.downloadButton.ConnectClicked(func() {
		if hb.onDownloadModel != nil {
			hb.onDownloadModel()
		}
	})
	hb.PackEnd(hb.downloadButton)

	// Chat settings button (system prompt)
	hb.settingsButton = gtk.NewButton()
	hb.settingsButton.SetIconName("emblem-system-symbolic")
	hb.settingsButton.SetTooltipText(i18n.T("Chat Settings"))
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
