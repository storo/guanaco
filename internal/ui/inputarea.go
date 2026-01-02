package ui

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// InputArea is the chat input widget with expandable text entry.
type InputArea struct {
	*gtk.Box

	textView    *gtk.TextView
	sendButton  *gtk.Button
	attachButton *gtk.Button
	scrolled    *gtk.ScrolledWindow

	// Callbacks
	onSend   func(text string)
	onAttach func()
}

// NewInputArea creates a new input area.
func NewInputArea() *InputArea {
	ia := &InputArea{}

	ia.Box = gtk.NewBox(gtk.OrientationHorizontal, 8)
	ia.AddCSSClass("input-area")
	ia.SetMarginTop(8)
	ia.SetMarginBottom(8)
	ia.SetMarginStart(8)
	ia.SetMarginEnd(8)

	ia.setupUI()

	return ia
}

func (ia *InputArea) setupUI() {
	// Attach button
	ia.attachButton = gtk.NewButton()
	ia.attachButton.SetIconName("mail-attachment-symbolic")
	ia.attachButton.SetTooltipText("Attach file")
	ia.attachButton.AddCSSClass("flat")
	ia.attachButton.SetVAlign(gtk.AlignEnd)
	ia.attachButton.ConnectClicked(func() {
		if ia.onAttach != nil {
			ia.onAttach()
		}
	})
	ia.Append(ia.attachButton)

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
	ia.scrolled.SetMaxContentHeight(150)
	ia.scrolled.SetPropagateNaturalHeight(true)
	ia.scrolled.SetHExpand(true)
	ia.scrolled.AddCSSClass("input-scrolled")
	ia.Append(ia.scrolled)

	// Send button
	ia.sendButton = gtk.NewButton()
	ia.sendButton.SetIconName("go-up-symbolic")
	ia.sendButton.SetTooltipText("Send message (Ctrl+Enter)")
	ia.sendButton.AddCSSClass("suggested-action")
	ia.sendButton.AddCSSClass("circular")
	ia.sendButton.SetVAlign(gtk.AlignEnd)
	ia.sendButton.ConnectClicked(ia.send)
	ia.Append(ia.sendButton)
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
