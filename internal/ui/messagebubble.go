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

	contentBox *gtk.Box
	container  *gtk.Box
	role       store.Role
	content    string
}

// NewMessageBubble creates a new message bubble.
func NewMessageBubble(role store.Role, content string) *MessageBubble {
	mb := &MessageBubble{
		role:    role,
		content: content,
	}

	mb.Box = gtk.NewBox(gtk.OrientationHorizontal, 8)
	mb.SetMarginTop(4)
	mb.SetMarginBottom(4)
	mb.SetMarginStart(12)
	mb.SetMarginEnd(12)

	mb.setupUI()

	return mb
}

func (mb *MessageBubble) setupUI() {
	mb.AddCSSClass("message-bubble")
	mb.SetHExpand(true)

	// Content box holds text labels and code blocks
	mb.contentBox = gtk.NewBox(gtk.OrientationVertical, 8)
	mb.contentBox.SetMarginTop(8)
	mb.contentBox.SetMarginBottom(8)
	mb.contentBox.SetMarginStart(16)
	mb.contentBox.SetMarginEnd(16)

	switch mb.role {
	case store.RoleUser:
		// User: pill/card aligned right
		mb.AddCSSClass("message-user")
		mb.SetMarginEnd(16)

		mb.container = gtk.NewBox(gtk.OrientationVertical, 0)
		mb.container.AddCSSClass("card")
		mb.container.Append(mb.contentBox)

		// Spacer pushes bubble to the right
		spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
		spacer.SetHExpand(true)
		mb.Append(spacer)
		mb.Append(mb.container)

	case store.RoleAssistant:
		// Assistant: plain text, no card background
		mb.AddCSSClass("message-assistant")
		mb.SetMarginStart(16)
		mb.SetMarginEnd(48) // Leave space on the right

		// No container/card - just contentBox directly
		mb.Append(mb.contentBox)

	case store.RoleSystem:
		// System: centered, subtle card
		mb.AddCSSClass("message-system")

		mb.container = gtk.NewBox(gtk.OrientationVertical, 0)
		mb.container.AddCSSClass("card")
		mb.container.Append(mb.contentBox)

		spacerL := gtk.NewBox(gtk.OrientationHorizontal, 0)
		spacerL.SetHExpand(true)
		spacerR := gtk.NewBox(gtk.OrientationHorizontal, 0)
		spacerR.SetHExpand(true)
		mb.Append(spacerL)
		mb.Append(mb.container)
		mb.Append(spacerR)
	}

	// Render initial content
	if mb.content != "" {
		mb.renderContent()
	}
}

// renderContent parses the content and creates appropriate widgets.
func (mb *MessageBubble) renderContent() {
	// Clear existing content
	for {
		child := mb.contentBox.FirstChild()
		if child == nil {
			break
		}
		mb.contentBox.Remove(child)
	}

	// Parse content into parts
	parts := mdRenderer.Parse(mb.content)

	// If no parts, just add as text
	if len(parts) == 0 {
		label := mb.createTextLabel(mb.content)
		mb.contentBox.Append(label)
		return
	}

	// Add each part
	for _, part := range parts {
		switch part.Type {
		case "code":
			codeBlock := NewCodeBlock(part.Content, part.Language)
			mb.contentBox.Append(codeBlock)
		case "text":
			label := mb.createTextLabel(part.Content)
			mb.contentBox.Append(label)
		}
	}
}

// createTextLabel creates a styled label for text content.
func (mb *MessageBubble) createTextLabel(text string) *gtk.Label {
	label := gtk.NewLabel("")
	label.SetWrap(true)
	label.SetWrapMode(pango.WrapWordChar)
	label.SetXAlign(0)
	label.SetSelectable(true)
	label.SetUseMarkup(true)

	// Render as pango markup
	label.SetMarkup(mdRenderer.ToPango(text))

	// Style based on role
	if mb.role == store.RoleSystem {
		label.AddCSSClass("dim-label")
	}

	return label
}

// SetContent updates the message content.
func (mb *MessageBubble) SetContent(content string) {
	mb.content = content
	mb.renderContent()
}

// AppendContent appends text to the current content.
func (mb *MessageBubble) AppendContent(text string) {
	mb.content += text
	mb.renderContent()
}

// GetContent returns the current content.
func (mb *MessageBubble) GetContent() string {
	return mb.content
}

// GetRole returns the message role.
func (mb *MessageBubble) GetRole() store.Role {
	return mb.role
}
