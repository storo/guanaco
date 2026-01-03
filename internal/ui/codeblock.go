package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

// Shared syntax highlighter instance
var sharedHighlighter = NewSyntaxHighlighter()

// CodeBlock is a widget that displays code with syntax highlighting and a copy button.
type CodeBlock struct {
	*gtk.Box

	// UI components
	header     *gtk.Box
	langLabel  *gtk.Label
	copyBtn    *gtk.Button
	textView   *gtk.TextView
	textBuffer *gtk.TextBuffer
	scrolled   *gtk.ScrolledWindow

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
	cb.applyHighlighting()

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

	// Create text buffer and view for syntax highlighting
	cb.textBuffer = gtk.NewTextBuffer(nil)
	cb.textView = gtk.NewTextViewWithBuffer(cb.textBuffer)
	cb.textView.SetEditable(false)
	cb.textView.SetCursorVisible(false)
	cb.textView.SetMonospace(true)
	cb.textView.AddCSSClass("code-content")
	cb.textView.SetWrapMode(gtk.WrapWordChar)
	cb.textView.SetLeftMargin(12)
	cb.textView.SetRightMargin(12)
	cb.textView.SetTopMargin(4)
	cb.textView.SetBottomMargin(12)

	// Wrap in scrolled window for horizontal scrolling on long lines
	cb.scrolled = gtk.NewScrolledWindow()
	cb.scrolled.SetChild(cb.textView)
	cb.scrolled.SetPolicy(gtk.PolicyAutomatic, gtk.PolicyNever)
	cb.scrolled.SetMinContentHeight(20)
	cb.scrolled.SetMaxContentHeight(400)

	cb.Append(cb.scrolled)
}

func (cb *CodeBlock) applyHighlighting() {
	tokens := sharedHighlighter.Highlight(cb.code, cb.language)

	// Clear buffer
	cb.textBuffer.SetText("")

	// Get iterator at start
	iter := cb.textBuffer.StartIter()

	for _, tok := range tokens {
		if tok.Text == "" {
			continue
		}

		// Create or get tag for this style
		tag := cb.getOrCreateTag(tok.Color, tok.Bold, tok.Italic)

		if tag != nil {
			// Insert with tag
			startOffset := iter.Offset()
			cb.textBuffer.Insert(iter, tok.Text)
			startIter := cb.textBuffer.IterAtOffset(startOffset)
			endIter := cb.textBuffer.IterAtOffset(iter.Offset())
			cb.textBuffer.ApplyTag(tag, startIter, endIter)
		} else {
			// Insert without tag
			cb.textBuffer.Insert(iter, tok.Text)
		}
	}
}

func (cb *CodeBlock) getOrCreateTag(color string, bold, italic bool) *gtk.TextTag {
	if color == "" && !bold && !italic {
		return nil
	}

	tagName := fmt.Sprintf("syntax_%s_%v_%v", color, bold, italic)

	tagTable := cb.textBuffer.TagTable()
	tag := tagTable.Lookup(tagName)

	if tag == nil {
		tag = gtk.NewTextTag(tagName)

		if color != "" {
			tag.SetObjectProperty("foreground", color)
		}
		if bold {
			tag.SetObjectProperty("weight", pango.WeightBold)
		}
		if italic {
			tag.SetObjectProperty("style", pango.StyleItalic)
		}

		tagTable.Add(tag)
	}

	return tag
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

// SetCode updates the code content with new highlighting.
func (cb *CodeBlock) SetCode(code string) {
	cb.code = code
	cb.applyHighlighting()
}

// GetCode returns the code content.
func (cb *CodeBlock) GetCode() string {
	return cb.code
}
