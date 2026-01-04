package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/i18n"
)

// SystemPromptDialog is a dialog for editing the system prompt.
type SystemPromptDialog struct {
	*adw.Window

	// UI components
	textView  *gtk.TextView
	saveBtn   *gtk.Button
	cancelBtn *gtk.Button

	// State
	initialPrompt string

	// Callbacks
	onSave func(string)
}

// NewSystemPromptDialog creates a new system prompt dialog.
func NewSystemPromptDialog(parent *gtk.Window, currentPrompt string) *SystemPromptDialog {
	d := &SystemPromptDialog{
		initialPrompt: currentPrompt,
	}

	d.Window = adw.NewWindow()
	d.SetTitle(i18n.T("System Prompt"))
	d.SetModal(true)
	d.SetDefaultSize(450, 380)
	d.SetResizable(true)
	if parent != nil {
		d.SetTransientFor(parent)
	}

	d.setupUI()

	return d
}

func (d *SystemPromptDialog) setupUI() {
	// Header bar with close button
	headerBar := adw.NewHeaderBar()
	headerBar.SetShowEndTitleButtons(true)
	headerBar.SetShowStartTitleButtons(true)
	headerBar.SetTitleWidget(gtk.NewLabel(i18n.T("System Prompt")))

	// Main content box
	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginTop(16)
	content.SetMarginBottom(24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)

	// Description
	desc := gtk.NewLabel(i18n.T("Set instructions that define how the AI should behave in this chat."))
	desc.AddCSSClass("dim-label")
	desc.SetWrap(true)
	desc.SetXAlign(0)
	content.Append(desc)

	// Text view in scrolled window
	d.textView = gtk.NewTextView()
	d.textView.SetWrapMode(gtk.WrapWordChar)
	d.textView.SetTopMargin(8)
	d.textView.SetBottomMargin(8)
	d.textView.SetLeftMargin(8)
	d.textView.SetRightMargin(8)

	// Set initial text
	if d.initialPrompt != "" {
		d.textView.Buffer().SetText(d.initialPrompt)
	}

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(d.textView)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetMinContentHeight(120)
	scrolled.SetVExpand(true)
	scrolled.AddCSSClass("card")
	content.Append(scrolled)

	// Button box
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	buttonBox.SetHAlign(gtk.AlignEnd)
	buttonBox.SetMarginTop(16)

	// Cancel button
	d.cancelBtn = gtk.NewButton()
	d.cancelBtn.SetLabel(i18n.T("Cancel"))
	d.cancelBtn.ConnectClicked(func() {
		d.Close()
	})
	buttonBox.Append(d.cancelBtn)

	// Save button
	d.saveBtn = gtk.NewButton()
	d.saveBtn.SetLabel(i18n.T("Save"))
	d.saveBtn.AddCSSClass("suggested-action")
	d.saveBtn.ConnectClicked(func() {
		buffer := d.textView.Buffer()
		start := buffer.StartIter()
		end := buffer.EndIter()
		text := buffer.Text(start, end, false)

		if d.onSave != nil {
			d.onSave(text)
		}
		d.Close()
	})
	buttonBox.Append(d.saveBtn)

	content.Append(buttonBox)

	// Use ToolbarView to add header bar
	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(content)

	d.SetContent(toolbarView)
}

// OnSave sets the callback for when the prompt is saved.
func (d *SystemPromptDialog) OnSave(callback func(string)) {
	d.onSave = callback
}
