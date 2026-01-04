package ui

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// ThinkingIndicator displays an animated "●●●" indicator while waiting for a response.
type ThinkingIndicator struct {
	*gtk.Box
	dots     [3]*gtk.Label
	tickerID glib.SourceHandle // Handle to stop the animation
	step     int               // Current active dot (0, 1, 2)
}

// NewThinkingIndicator creates a new animated thinking indicator.
func NewThinkingIndicator() *ThinkingIndicator {
	ti := &ThinkingIndicator{}
	ti.Box = gtk.NewBox(gtk.OrientationHorizontal, 6)
	ti.AddCSSClass("thinking-indicator")
	ti.SetMarginStart(16)
	ti.SetMarginTop(8)
	ti.SetMarginBottom(8)

	for i := 0; i < 3; i++ {
		dot := gtk.NewLabel("●")
		dot.AddCSSClass("thinking-dot")
		dot.SetOpacity(0.3)
		ti.dots[i] = dot
		ti.Append(dot)
	}

	// Start animation (every 200ms)
	ti.tickerID = glib.TimeoutAdd(200, ti.animate)

	return ti
}

// animate cycles through the dots, highlighting one at a time.
func (ti *ThinkingIndicator) animate() bool {
	// Reset opacity of all dots
	for _, dot := range ti.dots {
		dot.SetOpacity(0.3)
	}
	// Highlight the current dot
	ti.dots[ti.step].SetOpacity(1.0)
	ti.step = (ti.step + 1) % 3

	return true // Continue animation
}

// Stop stops the animation and cleans up resources.
func (ti *ThinkingIndicator) Stop() {
	if ti.tickerID > 0 {
		glib.SourceRemove(ti.tickerID)
		ti.tickerID = 0
	}
}
