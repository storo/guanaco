package ui

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"github.com/storo/guanaco/internal/store"
)

// Shared markdown renderer for all message bubbles
var mdRenderer = NewMarkdownRenderer()

// MessageBubble is a widget that displays a single chat message.
type MessageBubble struct {
	*gtk.Box

	label   *gtk.Label
	role    store.Role
	content string
}

// NewMessageBubble creates a new message bubble.
func NewMessageBubble(role store.Role, content string) *MessageBubble {
	mb := &MessageBubble{
		role:    role,
		content: content,
	}

	mb.Box = gtk.NewBox(gtk.OrientationVertical, 4)
	mb.SetMarginTop(4)
	mb.SetMarginBottom(4)
	mb.SetMarginStart(12)
	mb.SetMarginEnd(12)

	mb.setupUI()

	return mb
}

func (mb *MessageBubble) setupUI() {
	// Create label for content with markdown rendering
	mb.label = gtk.NewLabel("")
	mb.label.SetWrap(true)
	mb.label.SetWrapMode(pango.WrapWordChar)
	mb.label.SetXAlign(0)
	mb.label.SetSelectable(true)
	mb.label.SetMarginTop(8)
	mb.label.SetMarginBottom(8)
	mb.label.SetMarginStart(12)
	mb.label.SetMarginEnd(12)
	mb.label.SetUseMarkup(true)

	// Render initial content
	if mb.content != "" {
		mb.label.SetMarkup(mdRenderer.ToPango(mb.content))
	}

	// Style based on role
	mb.AddCSSClass("message-bubble")

	switch mb.role {
	case store.RoleUser:
		mb.AddCSSClass("message-user")
		mb.SetHAlign(gtk.AlignEnd)
		mb.label.AddCSSClass("accent")
	case store.RoleAssistant:
		mb.AddCSSClass("message-assistant")
		mb.SetHAlign(gtk.AlignStart)
	case store.RoleSystem:
		mb.AddCSSClass("message-system")
		mb.SetHAlign(gtk.AlignCenter)
		mb.label.AddCSSClass("dim-label")
	}

	// Container for bubble styling
	container := gtk.NewBox(gtk.OrientationVertical, 0)
	container.AddCSSClass("card")
	container.Append(mb.label)

	mb.Append(container)
}

// SetContent updates the message content.
func (mb *MessageBubble) SetContent(content string) {
	mb.content = content
	mb.label.SetMarkup(mdRenderer.ToPango(content))
}

// AppendContent appends text to the current content.
func (mb *MessageBubble) AppendContent(text string) {
	mb.content += text
	mb.label.SetMarkup(mdRenderer.ToPango(mb.content))
}

// GetContent returns the current content.
func (mb *MessageBubble) GetContent() string {
	return mb.content
}

// GetRole returns the message role.
func (mb *MessageBubble) GetRole() store.Role {
	return mb.role
}
