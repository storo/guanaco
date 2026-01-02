package ui

import (
	"context"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/ollama"
)

const (
	// DefaultWindowWidth is the default window width.
	DefaultWindowWidth = 900

	// DefaultWindowHeight is the default window height.
	DefaultWindowHeight = 600
)

// MainWindow is the main application window.
type MainWindow struct {
	*adw.ApplicationWindow

	// UI components
	headerBar    *HeaderBar
	splitView    *adw.NavigationSplitView
	toastOverlay *adw.ToastOverlay
	statusPage   *adw.StatusPage

	// State
	ollamaClient  *ollama.Client
	ollamaHealthy bool
}

// NewMainWindow creates a new main window.
func NewMainWindow(app *adw.Application) *MainWindow {
	win := &MainWindow{
		ollamaClient: ollama.NewClientDefault(),
	}

	win.ApplicationWindow = adw.NewApplicationWindow(&app.Application)
	win.SetDefaultSize(DefaultWindowWidth, DefaultWindowHeight)
	win.SetTitle("Guanaco")

	win.setupUI()
	win.checkOllamaHealth()

	return win
}

func (w *MainWindow) setupUI() {
	// Create header bar
	w.headerBar = NewHeaderBar()

	// Create split view for sidebar and content
	w.splitView = adw.NewNavigationSplitView()
	w.splitView.SetMinSidebarWidth(200)
	w.splitView.SetMaxSidebarWidth(300)
	w.splitView.SetSidebarWidthFraction(0.25)

	// Sidebar placeholder
	sidebarContent := gtk.NewBox(gtk.OrientationVertical, 0)
	sidebarLabel := gtk.NewLabel("Chats")
	sidebarLabel.AddCSSClass("title-2")
	sidebarLabel.SetMarginTop(12)
	sidebarLabel.SetMarginBottom(12)
	sidebarContent.Append(sidebarLabel)

	sidebarPage := adw.NewNavigationPage(sidebarContent, "Chats")
	w.splitView.SetSidebar(sidebarPage)

	// Content placeholder
	contentBox := gtk.NewBox(gtk.OrientationVertical, 0)
	contentLabel := gtk.NewLabel("Select a chat or start a new one")
	contentLabel.AddCSSClass("dim-label")
	contentBox.SetVExpand(true)
	contentBox.SetHExpand(true)
	contentBox.SetVAlign(gtk.AlignCenter)
	contentBox.SetHAlign(gtk.AlignCenter)
	contentBox.Append(contentLabel)

	contentPage := adw.NewNavigationPage(contentBox, "Chat")
	w.splitView.SetContent(contentPage)

	// Create status page for when Ollama is not running
	w.statusPage = adw.NewStatusPage()
	w.statusPage.SetIconName("dialog-warning-symbolic")
	w.statusPage.SetTitle("Ollama Not Detected")
	w.statusPage.SetDescription("Guanaco requires Ollama to be running.\nPlease start Ollama and restart the application.")

	// Toast overlay wraps content
	w.toastOverlay = adw.NewToastOverlay()
	w.toastOverlay.SetChild(w.splitView)

	// Main layout with toolbar view
	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(w.headerBar)
	toolbarView.SetContent(w.toastOverlay)

	w.SetContent(toolbarView)
}

func (w *MainWindow) checkOllamaHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w.ollamaHealthy = w.ollamaClient.IsHealthy(ctx)

	if !w.ollamaHealthy {
		w.showOllamaNotRunning()
	} else {
		w.loadModels()
	}
}

func (w *MainWindow) showOllamaNotRunning() {
	w.toastOverlay.SetChild(w.statusPage)
}

func (w *MainWindow) loadModels() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	models, err := w.ollamaClient.ListModels(ctx)
	if err != nil {
		w.showToast("Failed to load models")
		return
	}

	w.headerBar.SetModels(models)

	if len(models) > 0 {
		w.showToast("Loaded " + string(rune('0'+len(models))) + " models")
	}
}

func (w *MainWindow) showToast(message string) {
	toast := adw.NewToast(message)
	toast.SetTimeout(3)
	w.toastOverlay.AddToast(toast)
}
