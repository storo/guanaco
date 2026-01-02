package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/ollama"
)

const (
	// DefaultModel is the default model to use.
	DefaultModel = "llama3.2"
)

// HeaderBar is the application header bar with model selector.
type HeaderBar struct {
	*adw.HeaderBar

	// UI components
	modelEntry    *gtk.Entry
	newChatButton *gtk.Button
	menuButton    *gtk.MenuButton

	// State
	models       []ollama.Model
	currentModel string

	// Callbacks
	onModelChanged func(string)
}

// NewHeaderBar creates a new header bar.
func NewHeaderBar() *HeaderBar {
	hb := &HeaderBar{
		currentModel: DefaultModel,
	}

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

	// Model entry - editable text field for model name
	hb.modelEntry = gtk.NewEntry()
	hb.modelEntry.SetText(DefaultModel)
	hb.modelEntry.SetPlaceholderText("Model name (e.g., llama3.2)")
	hb.modelEntry.SetWidthChars(20)
	hb.modelEntry.SetMaxWidthChars(30)
	hb.modelEntry.AddCSSClass("model-entry")

	// Update model when text changes
	hb.modelEntry.ConnectChanged(func() {
		hb.currentModel = hb.modelEntry.Text()
		if hb.onModelChanged != nil {
			hb.onModelChanged(hb.currentModel)
		}
	})

	hb.SetTitleWidget(hb.modelEntry)

	// New chat button
	hb.newChatButton = gtk.NewButton()
	hb.newChatButton.SetIconName("list-add-symbolic")
	hb.newChatButton.SetTooltipText("New Chat")
	hb.newChatButton.AddCSSClass("suggested-action")
	hb.PackEnd(hb.newChatButton)
}

// SetModels updates suggestions based on available models.
func (hb *HeaderBar) SetModels(models []ollama.Model) {
	hb.models = models

	// If we have models and current is default, use first available
	if len(models) > 0 && hb.currentModel == DefaultModel {
		hb.currentModel = models[0].Name
		hb.modelEntry.SetText(hb.currentModel)
	}
}

// CurrentModel returns the currently entered model name.
func (hb *HeaderBar) CurrentModel() string {
	return hb.modelEntry.Text()
}

// SetModel sets the current model.
func (hb *HeaderBar) SetModel(model string) {
	hb.currentModel = model
	hb.modelEntry.SetText(model)
}

// OnModelChanged sets the callback for when the model changes.
func (hb *HeaderBar) OnModelChanged(callback func(string)) {
	hb.onModelChanged = callback
}

// OnNewChat sets the callback for when the new chat button is clicked.
func (hb *HeaderBar) OnNewChat(callback func()) {
	hb.newChatButton.ConnectClicked(callback)
}
