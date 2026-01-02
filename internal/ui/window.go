package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"

	"github.com/storo/guanaco/internal/config"
	"github.com/storo/guanaco/internal/ollama"
	"github.com/storo/guanaco/internal/store"
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
	sidebar      *Sidebar
	chatView     *ChatView

	// State
	ollamaClient  *ollama.Client
	ollamaHealthy bool
	db            *store.DB
}

// NewMainWindow creates a new main window.
func NewMainWindow(app *adw.Application) *MainWindow {
	win := &MainWindow{
		ollamaClient: ollama.NewClientDefault(),
	}

	win.ApplicationWindow = adw.NewApplicationWindow(&app.Application)
	win.SetDefaultSize(DefaultWindowWidth, DefaultWindowHeight)
	win.SetTitle("Guanaco")

	win.initDatabase()
	win.setupUI()
	win.checkOllamaHealth()

	return win
}

func (w *MainWindow) initDatabase() {
	dbPath := config.GetDatabasePath()
	db, err := store.NewDB(dbPath)
	if err != nil {
		// Log error but continue - app can work without persistence
		fmt.Printf("Warning: failed to open database: %v\n", err)
		return
	}
	w.db = db
}

func (w *MainWindow) setupUI() {
	// Create header bar
	w.headerBar = NewHeaderBar()
	w.headerBar.OnNewChat(w.onNewChat)

	// Create split view for sidebar and content
	w.splitView = adw.NewNavigationSplitView()
	w.splitView.SetMinSidebarWidth(200)
	w.splitView.SetMaxSidebarWidth(300)
	w.splitView.SetSidebarWidthFraction(0.25)

	// Sidebar with chat list
	w.sidebar = NewSidebar(w.db)
	w.sidebar.OnChatSelected(w.onChatSelected)

	sidebarPage := adw.NewNavigationPage(w.sidebar, "Chats")
	w.splitView.SetSidebar(sidebarPage)

	// Chat view
	w.chatView = NewChatView(w.ollamaClient, w.db)
	w.chatView.OnError(func(err error) {
		w.showToast("Error: " + err.Error())
	})

	contentPage := adw.NewNavigationPage(w.chatView, "Chat")
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
		w.sidebar.LoadChats()
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

	// Set current model in chat view
	if len(models) > 0 {
		w.chatView.SetModel(models[0].Name)
		w.showToast(fmt.Sprintf("Loaded %d models", len(models)))
	}
}

func (w *MainWindow) onNewChat() {
	w.chatView.NewChat()
	model := w.headerBar.CurrentModel()
	if model != "" {
		w.chatView.SetModel(model)
	}
}

func (w *MainWindow) onChatSelected(chat *store.Chat) {
	w.chatView.SetChat(chat)
}

func (w *MainWindow) showToast(message string) {
	toast := adw.NewToast(message)
	toast.SetTimeout(3)
	w.toastOverlay.AddToast(toast)
}
