package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/ollama"
)

// HeaderBar is the application header bar with model selector.
type HeaderBar struct {
	*adw.HeaderBar

	// UI components
	modelDropdown *gtk.DropDown
	newChatButton *gtk.Button
	menuButton    *gtk.MenuButton

	// State
	models       []ollama.Model
	currentModel string
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
	// Menu button (hamburger menu)
	hb.menuButton = gtk.NewMenuButton()
	hb.menuButton.SetIconName("open-menu-symbolic")
	hb.menuButton.SetTooltipText("Main Menu")
	hb.PackStart(hb.menuButton)

	// Model dropdown (will be populated later)
	hb.modelDropdown = gtk.NewDropDownFromStrings([]string{"No models"})
	hb.modelDropdown.SetSensitive(false)
	hb.SetTitleWidget(hb.modelDropdown)

	// New chat button
	hb.newChatButton = gtk.NewButton()
	hb.newChatButton.SetIconName("list-add-symbolic")
	hb.newChatButton.SetTooltipText("New Chat")
	hb.newChatButton.AddCSSClass("suggested-action")
	hb.PackEnd(hb.newChatButton)
}

// SetModels updates the model dropdown with available models.
func (hb *HeaderBar) SetModels(models []ollama.Model) {
	hb.models = models

	if len(models) == 0 {
		hb.modelDropdown.SetSensitive(false)
		return
	}

	// Create string list for dropdown
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.Name
	}

	// Replace dropdown with new one containing models
	newDropdown := gtk.NewDropDownFromStrings(names)
	newDropdown.SetSensitive(true)

	// Connect to selection changes using NotifyProperty
	newDropdown.NotifyProperty("selected", func() {
		hb.onModelSelected()
	})

	hb.SetTitleWidget(newDropdown)
	hb.modelDropdown = newDropdown

	// Set first model as current
	if len(models) > 0 {
		hb.currentModel = models[0].Name
	}
}

func (hb *HeaderBar) onModelSelected() {
	selected := hb.modelDropdown.Selected()
	if int(selected) < len(hb.models) {
		hb.currentModel = hb.models[selected].Name
	}
}

// CurrentModel returns the currently selected model name.
func (hb *HeaderBar) CurrentModel() string {
	return hb.currentModel
}

// OnNewChat sets the callback for when the new chat button is clicked.
func (hb *HeaderBar) OnNewChat(callback func()) {
	hb.newChatButton.ConnectClicked(callback)
}
