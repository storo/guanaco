package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/i18n"
)

// AttachmentPill is a visual widget showing an attached document.
type AttachmentPill struct {
	*gtk.Box

	// UI components
	label     *gtk.Label
	removeBtn *gtk.Button

	// Data
	filename string
	content  string
	isImage  bool

	// Callbacks
	onRemove func()
}

// NewAttachmentPill creates a new attachment pill widget.
func NewAttachmentPill(filename, content string) *AttachmentPill {
	pill := &AttachmentPill{
		filename: filename,
		content:  content,
		isImage:  isImageFile(filename),
	}

	pill.Box = gtk.NewBox(gtk.OrientationHorizontal, 4)
	pill.AddCSSClass("attachment-pill")
	pill.AddCSSClass("card")

	pill.setupUI()

	return pill
}

// isImageFile checks if a filename is an image.
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	}
	return false
}

func (p *AttachmentPill) setupUI() {
	// Icon based on file type
	var iconName string
	if p.isImage {
		iconName = "image-x-generic-symbolic"
	} else {
		iconName = "text-x-generic-symbolic"
	}
	icon := gtk.NewImageFromIconName(iconName)
	icon.SetMarginStart(8)
	p.Append(icon)

	// Filename label
	displayName := p.filename
	if len(displayName) > 20 {
		ext := filepath.Ext(displayName)
		base := displayName[:len(displayName)-len(ext)]
		if len(base) > 17 {
			base = base[:17]
		}
		displayName = base + "…" + ext
	}

	p.label = gtk.NewLabel(displayName)
	p.label.SetTooltipText(fmt.Sprintf(i18n.T("%s (%d chars)"), p.filename, len(p.content)))
	p.label.SetMarginStart(4)
	p.label.SetMarginEnd(4)
	p.Append(p.label)

	// Remove button
	p.removeBtn = gtk.NewButton()
	p.removeBtn.SetIconName("window-close-symbolic")
	p.removeBtn.AddCSSClass("flat")
	p.removeBtn.AddCSSClass("circular")
	p.removeBtn.SetTooltipText(i18n.T("Remove attachment"))
	p.removeBtn.ConnectClicked(func() {
		if p.onRemove != nil {
			p.onRemove()
		}
	})
	p.Append(p.removeBtn)
}

// Filename returns the attachment filename.
func (p *AttachmentPill) Filename() string {
	return p.filename
}

// Content returns the extracted document content.
func (p *AttachmentPill) Content() string {
	return p.content
}

// OnRemove sets the callback for when the remove button is clicked.
func (p *AttachmentPill) OnRemove(callback func()) {
	p.onRemove = callback
}

// IsImage returns true if this attachment is an image.
func (p *AttachmentPill) IsImage() bool {
	return p.isImage
}
