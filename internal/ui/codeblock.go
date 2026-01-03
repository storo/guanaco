package ui

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

// CodeBlock is a widget that displays code with a copy button.
type CodeBlock struct {
	*gtk.Box

	// UI components
	header    *gtk.Box
	langLabel *gtk.Label
	copyBtn   *gtk.Button
	codeLabel *gtk.Label

	// Data
	code     string
	language string
}

// NewCodeBlock creates a new code block widget.
func NewCodeBlock(code, language string) *CodeBlock {
	cb := &CodeBlock{
		code:     code,
		language: language,
	}

	cb.Box = gtk.NewBox(gtk.OrientationVertical, 0)
	cb.AddCSSClass("code-block")

	cb.setupUI()

	return cb
}

func (cb *CodeBlock) setupUI() {
	// Header with language and copy button
	cb.header = gtk.NewBox(gtk.OrientationHorizontal, 8)
	cb.header.AddCSSClass("code-block-header")
	cb.header.SetMarginStart(12)
	cb.header.SetMarginEnd(8)
	cb.header.SetMarginTop(6)
	cb.header.SetMarginBottom(4)

	// Language label
	if cb.language != "" {
		cb.langLabel = gtk.NewLabel(cb.language)
		cb.langLabel.AddCSSClass("code-lang")
		cb.langLabel.SetHExpand(true)
		cb.langLabel.SetXAlign(0)
		cb.header.Append(cb.langLabel)
	} else {
		// Spacer
		spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
		spacer.SetHExpand(true)
		cb.header.Append(spacer)
	}

	// Copy button
	cb.copyBtn = gtk.NewButton()
	cb.copyBtn.SetIconName("edit-copy-symbolic")
	cb.copyBtn.SetTooltipText("Copy code")
	cb.copyBtn.AddCSSClass("flat")
	cb.copyBtn.AddCSSClass("circular")
	cb.copyBtn.ConnectClicked(cb.copyToClipboard)
	cb.header.Append(cb.copyBtn)

	cb.Append(cb.header)

	// Code content
	cb.codeLabel = gtk.NewLabel(cb.code)
	cb.codeLabel.AddCSSClass("code-content")
	cb.codeLabel.SetWrap(true)
	cb.codeLabel.SetWrapMode(pango.WrapWordChar)
	cb.codeLabel.SetXAlign(0)
	cb.codeLabel.SetSelectable(true)
	cb.codeLabel.SetMarginStart(12)
	cb.codeLabel.SetMarginEnd(12)
	cb.codeLabel.SetMarginBottom(12)

	cb.Append(cb.codeLabel)
}

func (cb *CodeBlock) copyToClipboard() {
	display := gdk.DisplayGetDefault()
	clipboard := display.Clipboard()
	clipboard.SetText(cb.code)

	// Visual feedback - change icon temporarily
	cb.copyBtn.SetIconName("object-select-symbolic")
	cb.copyBtn.SetTooltipText("Copied!")

	// Reset after delay
	glib.TimeoutAdd(1500, func() bool {
		cb.copyBtn.SetIconName("edit-copy-symbolic")
		cb.copyBtn.SetTooltipText("Copy code")
		return false
	})
}

// SetCode updates the code content.
func (cb *CodeBlock) SetCode(code string) {
	cb.code = code
	cb.codeLabel.SetText(code)
}

// GetCode returns the code content.
func (cb *CodeBlock) GetCode() string {
	return cb.code
}
