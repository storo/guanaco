package ui

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/i18n"
	"github.com/storo/guanaco/internal/ollama"
)

// InputArea is the chat input widget with expandable text entry.
type InputArea struct {
	*gtk.Box

	// Layout
	mainBox       *gtk.Box
	attachmentBox *gtk.Box
	inputBox      *gtk.Box

	// Input components
	textView     *gtk.TextView
	sendButton   *gtk.Button
	stopButton   *gtk.Button
	attachButton *gtk.Button
	scrolled     *gtk.ScrolledWindow

	// Model selector
	modelButton  *gtk.MenuButton
	modelLabel   *gtk.Label
	modelListBox *gtk.ListBox
	models       []ollama.Model
	currentModel string

	// State
	attachments    []*AttachmentPill
	loadingSpinner *gtk.Spinner

	// Callbacks
	onSend         func(text string)
	onAttach       func()
	onStop         func()
	onModelChanged func(string)
}

// NewInputArea creates a new input area.
func NewInputArea() *InputArea {
	ia := &InputArea{}

	ia.Box = gtk.NewBox(gtk.OrientationVertical, 4)
	ia.AddCSSClass("input-area")
	ia.SetMarginTop(8)
	ia.SetMarginBottom(16)
	ia.SetMarginStart(16)
	ia.SetMarginEnd(16)

	ia.setupUI()

	return ia
}

func (ia *InputArea) setupUI() {
	// Attachment pills box (hidden by default)
	ia.attachmentBox = gtk.NewBox(gtk.OrientationHorizontal, 4)
	ia.attachmentBox.SetMarginBottom(4)
	ia.attachmentBox.SetVisible(false)
	ia.Append(ia.attachmentBox)

	// Input row (horizontal box)
	ia.inputBox = gtk.NewBox(gtk.OrientationHorizontal, 8)
	ia.Append(ia.inputBox)

	// Attach button
	ia.attachButton = gtk.NewButton()
	ia.attachButton.SetIconName("mail-attachment-symbolic")
	ia.attachButton.SetTooltipText(i18n.T("Attach file"))
	ia.attachButton.AddCSSClass("flat")
	ia.attachButton.SetVAlign(gtk.AlignEnd)
	ia.attachButton.ConnectClicked(func() {
		if ia.onAttach != nil {
			ia.onAttach()
		}
	})
	ia.inputBox.Append(ia.attachButton)

	// Text view in scrolled window
	ia.textView = gtk.NewTextView()
	ia.textView.SetWrapMode(gtk.WrapWordChar)
	ia.textView.SetAcceptsTab(false)
	ia.textView.SetTopMargin(8)
	ia.textView.SetBottomMargin(8)
	ia.textView.SetLeftMargin(12)
	ia.textView.SetRightMargin(12)
	ia.textView.AddCSSClass("input-textview")

	// Handle key press for Ctrl+Enter to send
	keyController := gtk.NewEventControllerKey()
	keyController.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		if keyval == gdk.KEY_Return && state&gdk.ControlMask != 0 {
			ia.send()
			return true
		}
		return false
	})
	ia.textView.AddController(keyController)

	ia.scrolled = gtk.NewScrolledWindow()
	ia.scrolled.SetChild(ia.textView)
	ia.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	ia.scrolled.SetMinContentHeight(40)
	ia.scrolled.SetMaxContentHeight(150)
	ia.scrolled.SetHExpand(true)
	ia.scrolled.AddCSSClass("input-scrolled")
	ia.inputBox.Append(ia.scrolled)

	// Auto-resize based on content
	buffer := ia.textView.Buffer()
	buffer.ConnectChanged(func() {
		ia.updateHeight()
	})

	// Model selector dropdown
	ia.modelLabel = gtk.NewLabel("model")
	ia.modelLabel.AddCSSClass("dim-label")

	ia.modelButton = gtk.NewMenuButton()
	ia.modelButton.SetChild(ia.modelLabel)
	ia.modelButton.AddCSSClass("flat")
	ia.modelButton.SetVAlign(gtk.AlignEnd)
	ia.modelButton.SetTooltipText(i18n.T("Select model"))

	// Create popover with model list
	popover := gtk.NewPopover()
	popover.SetAutohide(true)

	ia.modelListBox = gtk.NewListBox()
	ia.modelListBox.SetSelectionMode(gtk.SelectionSingle)
	ia.modelListBox.AddCSSClass("boxed-list")
	ia.modelListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(ia.models) {
			ia.selectModel(ia.models[idx].Name)
			popover.Popdown()
		}
	})

	scrolledList := gtk.NewScrolledWindow()
	scrolledList.SetChild(ia.modelListBox)
	scrolledList.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolledList.SetMinContentHeight(100)
	scrolledList.SetMaxContentHeight(250)
	scrolledList.SetSizeRequest(200, -1)

	popover.SetChild(scrolledList)
	ia.modelButton.SetPopover(popover)
	ia.inputBox.Append(ia.modelButton)

	// Send button
	ia.sendButton = gtk.NewButton()
	ia.sendButton.SetIconName("go-up-symbolic")
	ia.sendButton.SetTooltipText(i18n.T("Send message (Ctrl+Enter)"))
	ia.sendButton.AddCSSClass("suggested-action")
	ia.sendButton.AddCSSClass("circular")
	ia.sendButton.SetVAlign(gtk.AlignEnd)
	ia.sendButton.ConnectClicked(ia.send)
	ia.inputBox.Append(ia.sendButton)

	// Stop button (hidden initially, shown during streaming)
	ia.stopButton = gtk.NewButton()
	ia.stopButton.SetIconName("media-playback-stop-symbolic")
	ia.stopButton.SetTooltipText(i18n.T("Stop generation"))
	ia.stopButton.AddCSSClass("destructive-action")
	ia.stopButton.AddCSSClass("circular")
	ia.stopButton.SetVAlign(gtk.AlignEnd)
	ia.stopButton.SetVisible(false)
	ia.stopButton.ConnectClicked(func() {
		if ia.onStop != nil {
			ia.onStop()
		}
	})
	ia.inputBox.Append(ia.stopButton)
}

func (ia *InputArea) send() {
	buffer := ia.textView.Buffer()
	start := buffer.StartIter()
	end := buffer.EndIter()
	text := buffer.Text(start, end, false)

	if text == "" {
		return
	}

	if ia.onSend != nil {
		ia.onSend(text)
	}

	// Clear the text
	buffer.SetText("")
}

// OnSend sets the callback for when a message is sent.
func (ia *InputArea) OnSend(callback func(text string)) {
	ia.onSend = callback
}

// OnAttach sets the callback for when the attach button is clicked.
func (ia *InputArea) OnAttach(callback func()) {
	ia.onAttach = callback
}

// SetSensitive enables or disables the input area.
func (ia *InputArea) SetInputSensitive(sensitive bool) {
	ia.textView.SetSensitive(sensitive)
	ia.sendButton.SetSensitive(sensitive)
	ia.attachButton.SetSensitive(sensitive)
}

// Focus sets focus to the text entry.
func (ia *InputArea) Focus() {
	ia.textView.GrabFocus()
}

// GetText returns the current text in the input.
func (ia *InputArea) GetText() string {
	buffer := ia.textView.Buffer()
	start := buffer.StartIter()
	end := buffer.EndIter()
	return buffer.Text(start, end, false)
}

// SetText sets the text in the input.
func (ia *InputArea) SetText(text string) {
	buffer := ia.textView.Buffer()
	buffer.SetText(text)
}

// AddAttachment adds an attachment pill to the input area.
func (ia *InputArea) AddAttachment(pill *AttachmentPill) {
	// Set up remove callback
	pill.OnRemove(func() {
		ia.RemoveAttachment(pill)
	})

	ia.attachments = append(ia.attachments, pill)
	ia.attachmentBox.Append(pill)
	ia.attachmentBox.SetVisible(true)
}

// RemoveAttachment removes an attachment pill from the input area.
func (ia *InputArea) RemoveAttachment(pill *AttachmentPill) {
	for i, p := range ia.attachments {
		if p == pill {
			ia.attachments = append(ia.attachments[:i], ia.attachments[i+1:]...)
			break
		}
	}
	ia.attachmentBox.Remove(pill)

	if len(ia.attachments) == 0 {
		ia.attachmentBox.SetVisible(false)
	}
}

// GetAttachments returns all current attachments.
func (ia *InputArea) GetAttachments() []*AttachmentPill {
	return ia.attachments
}

// ClearAttachments removes all attachments.
func (ia *InputArea) ClearAttachments() {
	for _, pill := range ia.attachments {
		ia.attachmentBox.Remove(pill)
	}
	ia.attachments = nil
	ia.attachmentBox.SetVisible(false)
}

// HasAttachments returns true if there are any attachments.
func (ia *InputArea) HasAttachments() bool {
	return len(ia.attachments) > 0
}

// ShowLoadingIndicator shows a spinner while processing an attachment.
func (ia *InputArea) ShowLoadingIndicator() {
	if ia.loadingSpinner == nil {
		ia.loadingSpinner = gtk.NewSpinner()
		ia.loadingSpinner.SetSizeRequest(24, 24)
	}
	ia.loadingSpinner.Start()
	ia.attachmentBox.Prepend(ia.loadingSpinner)
	ia.attachmentBox.SetVisible(true)
}

// HideLoadingIndicator hides the processing spinner.
func (ia *InputArea) HideLoadingIndicator() {
	if ia.loadingSpinner != nil {
		ia.loadingSpinner.Stop()
		ia.attachmentBox.Remove(ia.loadingSpinner)
		if len(ia.attachments) == 0 {
			ia.attachmentBox.SetVisible(false)
		}
	}
}

// OnStop sets the callback for when the stop button is clicked.
func (ia *InputArea) OnStop(callback func()) {
	ia.onStop = callback
}

// SetStreamingMode toggles between send and stop buttons.
func (ia *InputArea) SetStreamingMode(streaming bool) {
	ia.sendButton.SetVisible(!streaming)
	ia.stopButton.SetVisible(streaming)
	ia.textView.SetSensitive(!streaming)
	ia.attachButton.SetSensitive(!streaming)
}

// selectModel updates the current model and triggers callback.
func (ia *InputArea) selectModel(model string) {
	ia.currentModel = model
	ia.modelLabel.SetText(model)
	if ia.onModelChanged != nil {
		ia.onModelChanged(model)
	}
}

// SetModels updates the list of available models.
func (ia *InputArea) SetModels(models []ollama.Model) {
	ia.models = models

	// Clear existing rows
	for {
		row := ia.modelListBox.RowAtIndex(0)
		if row == nil {
			break
		}
		ia.modelListBox.Remove(row)
	}

	// Add model rows
	for _, model := range models {
		label := gtk.NewLabel(model.Name)
		label.SetXAlign(0)
		label.SetMarginTop(8)
		label.SetMarginBottom(8)
		label.SetMarginStart(12)
		label.SetMarginEnd(12)

		row := gtk.NewListBoxRow()
		row.SetChild(label)
		ia.modelListBox.Append(row)
	}

	// Select first model if none selected
	if len(models) > 0 && ia.currentModel == "" {
		ia.selectModel(models[0].Name)
	}
}

// SetModel sets the current model.
func (ia *InputArea) SetModel(model string) {
	ia.currentModel = model
	ia.modelLabel.SetText(model)
}

// CurrentModel returns the currently selected model.
func (ia *InputArea) CurrentModel() string {
	return ia.currentModel
}

// OnModelChanged sets the callback for when the model changes.
func (ia *InputArea) OnModelChanged(callback func(string)) {
	ia.onModelChanged = callback
}

// updateHeight adjusts the input area height based on content.
func (ia *InputArea) updateHeight() {
	buffer := ia.textView.Buffer()
	text := buffer.Text(buffer.StartIter(), buffer.EndIter(), false)

	// Count lines (including line breaks)
	lines := strings.Count(text, "\n") + 1

	// Clamp between 1 and 6 lines
	if lines < 1 {
		lines = 1
	}
	if lines > 6 {
		lines = 6
	}

	// ~24px per line, min 40px
	height := lines * 24
	if height < 40 {
		height = 40
	}

	ia.scrolled.SetMinContentHeight(height)
}
