package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/config"
)

// Language represents a selectable language option.
type Language struct {
	Code string
	Name string
}

var availableLanguages = []Language{
	{"auto", "Auto (System)"},
	{"en", "English"},
	{"es", "Español"},
	{"pt", "Português"},
	{"fr", "Français"},
	{"de", "Deutsch"},
}

// SettingsDialog is a dialog for configuring application settings.
type SettingsDialog struct {
	*adw.Window

	// UI components
	modelDropdown    *gtk.DropDown
	languageDropdown *gtk.DropDown
	systemPromptView *gtk.TextView

	// Data
	config *config.AppConfig
	models []string

	// Callbacks
	onSave func(*config.AppConfig)
}

// NewSettingsDialog creates a new settings dialog.
func NewSettingsDialog(parent *gtk.Window, cfg *config.AppConfig, models []string) *SettingsDialog {
	d := &SettingsDialog{
		config: cfg,
		models: models,
	}

	d.Window = adw.NewWindow()
	d.SetTitle("Settings")
	d.SetModal(true)
	d.SetDefaultSize(450, 500)
	if parent != nil {
		d.SetTransientFor(parent)
	}

	d.setupUI()

	return d
}

func (d *SettingsDialog) setupUI() {
	// Header bar
	headerBar := adw.NewHeaderBar()
	headerBar.SetShowEndTitleButtons(true)
	headerBar.SetShowStartTitleButtons(true)
	headerBar.SetTitleWidget(gtk.NewLabel("Settings"))

	// Main content
	content := gtk.NewBox(gtk.OrientationVertical, 16)
	content.SetMarginTop(16)
	content.SetMarginBottom(24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)

	// === Default Model ===
	modelLabel := gtk.NewLabel("Default Model:")
	modelLabel.SetXAlign(0)
	modelLabel.AddCSSClass("heading")
	content.Append(modelLabel)

	d.modelDropdown = d.createModelDropdown()
	content.Append(d.modelDropdown)

	// === Response Language ===
	langLabel := gtk.NewLabel("Response Language:")
	langLabel.SetXAlign(0)
	langLabel.SetMarginTop(8)
	langLabel.AddCSSClass("heading")
	content.Append(langLabel)

	d.languageDropdown = d.createLanguageDropdown()
	content.Append(d.languageDropdown)

	// === Global System Prompt ===
	promptLabel := gtk.NewLabel("Global System Prompt:")
	promptLabel.SetXAlign(0)
	promptLabel.SetMarginTop(8)
	promptLabel.AddCSSClass("heading")
	content.Append(promptLabel)

	promptHint := gtk.NewLabel("Applied to all new chats (chat-specific prompts take priority)")
	promptHint.SetXAlign(0)
	promptHint.AddCSSClass("dim-label")
	promptHint.AddCSSClass("caption")
	content.Append(promptHint)

	d.systemPromptView = gtk.NewTextView()
	d.systemPromptView.SetWrapMode(gtk.WrapWord)
	d.systemPromptView.Buffer().SetText(d.config.GlobalSystemPrompt)

	promptScrolled := gtk.NewScrolledWindow()
	promptScrolled.SetChild(d.systemPromptView)
	promptScrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	promptScrolled.SetMinContentHeight(120)
	promptScrolled.SetVExpand(true)
	promptScrolled.AddCSSClass("card")
	content.Append(promptScrolled)

	// === Buttons ===
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	buttonBox.SetHAlign(gtk.AlignEnd)
	buttonBox.SetMarginTop(16)

	cancelBtn := gtk.NewButton()
	cancelBtn.SetLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		d.Close()
	})
	buttonBox.Append(cancelBtn)

	saveBtn := gtk.NewButton()
	saveBtn.SetLabel("Save")
	saveBtn.AddCSSClass("suggested-action")
	saveBtn.ConnectClicked(d.onSaveClicked)
	buttonBox.Append(saveBtn)

	content.Append(buttonBox)

	// Layout
	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(content)

	d.SetContent(toolbarView)
}

func (d *SettingsDialog) createModelDropdown() *gtk.DropDown {
	// Create string list for models
	modelList := gtk.NewStringList(nil)

	// Add "None" option first
	modelList.Append("(None - use first available)")

	selectedIdx := uint(0)
	for i, model := range d.models {
		modelList.Append(model)
		if model == d.config.DefaultModel {
			selectedIdx = uint(i + 1) // +1 because of "None" option
		}
	}

	dropdown := gtk.NewDropDown(modelList, nil)
	dropdown.SetSelected(selectedIdx)

	return dropdown
}

func (d *SettingsDialog) createLanguageDropdown() *gtk.DropDown {
	langList := gtk.NewStringList(nil)

	selectedIdx := uint(0)
	for i, lang := range availableLanguages {
		langList.Append(lang.Name)
		if lang.Code == d.config.ResponseLanguage {
			selectedIdx = uint(i)
		}
	}

	dropdown := gtk.NewDropDown(langList, nil)
	dropdown.SetSelected(selectedIdx)

	return dropdown
}

func (d *SettingsDialog) onSaveClicked() {
	// Get selected model
	modelIdx := d.modelDropdown.Selected()
	if modelIdx == 0 {
		d.config.DefaultModel = ""
	} else if int(modelIdx-1) < len(d.models) {
		d.config.DefaultModel = d.models[modelIdx-1]
	}

	// Get selected language
	langIdx := d.languageDropdown.Selected()
	if int(langIdx) < len(availableLanguages) {
		d.config.ResponseLanguage = availableLanguages[langIdx].Code
	}

	// Get system prompt
	buffer := d.systemPromptView.Buffer()
	start, end := buffer.Bounds()
	d.config.GlobalSystemPrompt = buffer.Text(start, end, false)

	// Save and notify
	d.config.Save()

	if d.onSave != nil {
		d.onSave(d.config)
	}

	d.Close()
}

// OnSave sets the callback for when settings are saved.
func (d *SettingsDialog) OnSave(callback func(*config.AppConfig)) {
	d.onSave = callback
}
