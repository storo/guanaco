package ui

import (
	"context"
	"os/exec"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/config"
	"github.com/storo/guanaco/internal/logger"
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
	appConfig     *config.AppConfig
	models        []ollama.Model
}

// NewMainWindow creates a new main window.
func NewMainWindow(app *adw.Application) *MainWindow {
	win := &MainWindow{
		ollamaClient: ollama.NewClientDefault(),
	}

	win.ApplicationWindow = adw.NewApplicationWindow(&app.Application)
	win.SetDefaultSize(DefaultWindowWidth, DefaultWindowHeight)
	win.SetTitle("Guanaco")

	win.loadConfig()
	win.initDatabase()
	win.setupUI()
	win.checkOllamaHealth()

	return win
}

func (w *MainWindow) loadConfig() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		cfg = config.DefaultConfig()
	}
	w.appConfig = cfg
	logger.Info("Config loaded", "defaultModel", cfg.DefaultModel, "language", cfg.ResponseLanguage)
}

func (w *MainWindow) initDatabase() {
	dbPath := config.GetDatabasePath()
	db, err := store.NewDB(dbPath)
	if err != nil {
		// Log error but continue - app can work without persistence
		logger.Error("Failed to open database", "path", dbPath, "error", err)
		return
	}
	logger.Info("Database opened", "path", dbPath)
	w.db = db
}

func (w *MainWindow) setupUI() {
	// Create header bar
	w.headerBar = NewHeaderBar()
	w.headerBar.OnDownloadModel(w.onDownloadModel)
	w.headerBar.OnChatSettings(w.onChatSettings)
	w.headerBar.OnToggleSidebar(w.onToggleSidebar)

	// Create split view for sidebar and content
	w.splitView = adw.NewNavigationSplitView()
	w.splitView.SetMinSidebarWidth(200)
	w.splitView.SetMaxSidebarWidth(300)
	w.splitView.SetSidebarWidthFraction(0.25)

	// Sidebar with chat list
	w.sidebar = NewSidebar(w.db)
	w.sidebar.OnChatSelected(w.onChatSelected)
	w.sidebar.OnNewChat(w.onNewChat)
	w.sidebar.OnChatDeleted(w.onChatDeleted)
	w.sidebar.OnSettings(w.onSettings)

	sidebarPage := adw.NewNavigationPage(w.sidebar, "Chats")
	w.splitView.SetSidebar(sidebarPage)

	// Apply sidebar visibility from config (collapsed=true means sidebar hidden)
	w.splitView.SetCollapsed(!w.appConfig.SidebarVisible)

	// Chat view
	w.chatView = NewChatView(w.ollamaClient, w.db)
	w.chatView.SetAppConfig(w.appConfig)
	w.chatView.OnError(func(err error) {
		w.showToast("Error: " + err.Error())
	})
	w.chatView.OnTitleChanged(func(title string) {
		w.sidebar.Refresh()
	})
	w.chatView.GetInputArea().OnModelChanged(w.onModelChanged)

	contentPage := adw.NewNavigationPage(w.chatView, "Chat")
	w.splitView.SetContent(contentPage)

	// Create status page for when Ollama is not running
	w.statusPage = adw.NewStatusPage()
	w.statusPage.SetIconName("dialog-warning-symbolic")
	w.statusPage.SetTitle("Ollama Not Detected")
	w.statusPage.SetDescription("Guanaco requires Ollama to be running.\nClick the button below to start Ollama.")

	// Button box for status page actions
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
	buttonBox.SetHAlign(gtk.AlignCenter)

	// Start Ollama button
	startButton := gtk.NewButton()
	startButton.SetLabel("Start Ollama")
	startButton.AddCSSClass("suggested-action")
	startButton.AddCSSClass("pill")
	startButton.ConnectClicked(w.onStartOllama)
	buttonBox.Append(startButton)

	// Retry button
	retryButton := gtk.NewButton()
	retryButton.SetLabel("Retry Connection")
	retryButton.AddCSSClass("pill")
	retryButton.ConnectClicked(func() {
		w.checkOllamaHealth()
	})
	buttonBox.Append(retryButton)

	w.statusPage.SetChild(buttonBox)

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
		logger.Error("Failed to load models", "error", err)
		w.showToast("Failed to load models: " + err.Error())
		return
	}

	w.models = models
	w.chatView.GetInputArea().SetModels(models)

	// Use default model from config, or first available
	defaultModel := ""
	if w.appConfig != nil && w.appConfig.DefaultModel != "" {
		// Check if default model exists
		for _, m := range models {
			if m.Name == w.appConfig.DefaultModel {
				defaultModel = w.appConfig.DefaultModel
				break
			}
		}
	}
	if defaultModel == "" && len(models) > 0 {
		defaultModel = models[0].Name
	}

	if defaultModel != "" {
		w.chatView.SetModel(defaultModel)
		w.chatView.GetInputArea().SetModel(defaultModel)
		logger.Info("Models loaded", "count", len(models), "defaultModel", defaultModel)
	} else {
		logger.Warn("No models found")
		w.showToast("No models found. Use the download button to pull a model.")
	}
}

func (w *MainWindow) onNewChat() {
	w.chatView.NewChat()

	// Usar modelo por defecto de config, o el actual si no hay default
	model := ""
	if w.appConfig != nil && w.appConfig.DefaultModel != "" {
		model = w.appConfig.DefaultModel
	} else {
		model = w.chatView.GetInputArea().CurrentModel()
	}

	if model != "" {
		w.chatView.SetModel(model)
		w.chatView.GetInputArea().SetModel(model)
	}
}

func (w *MainWindow) onModelChanged(model string) {
	w.chatView.SetModel(model)
}

func (w *MainWindow) onChatSelected(chat *store.Chat) {
	w.chatView.SetChat(chat)
}

func (w *MainWindow) onChatDeleted(chatID int64) {
	// If the deleted chat is the current one, start a new chat
	if currentChat := w.chatView.GetCurrentChat(); currentChat != nil && currentChat.ID == chatID {
		w.chatView.NewChat()
	}
}

func (w *MainWindow) onDownloadModel() {
	dialog := NewModelDialog(&w.ApplicationWindow.Window, w.ollamaClient)
	dialog.OnModelDownloaded(func(model string) {
		w.loadModels()
		w.chatView.GetInputArea().SetModel(model)
		w.chatView.SetModel(model)
		w.showToast("Model " + model + " downloaded!")
	})
	dialog.Present()
}

func (w *MainWindow) onChatSettings() {
	// Ensure a chat exists before opening the dialog
	if w.chatView.GetCurrentChat() == nil {
		w.chatView.EnsureChat(w.chatView.GetInputArea().CurrentModel())
	}

	// Get current system prompt from chat
	currentPrompt := ""
	if chat := w.chatView.GetCurrentChat(); chat != nil {
		currentPrompt = chat.SystemPrompt
	}

	dialog := NewSystemPromptDialog(&w.ApplicationWindow.Window, currentPrompt)
	dialog.OnSave(func(prompt string) {
		if chat := w.chatView.GetCurrentChat(); chat != nil {
			chat.SystemPrompt = prompt
			if w.db != nil {
				w.db.UpdateChatSystemPrompt(chat.ID, prompt)
			}
			w.showToast("System prompt saved")
		}
	})
	dialog.Present()
}

func (w *MainWindow) showToast(message string) {
	toast := adw.NewToast(message)
	toast.SetTimeout(3)
	w.toastOverlay.AddToast(toast)
}

func (w *MainWindow) onStartOllama() {
	logger.Info("Attempting to start Ollama")
	w.showToast("Starting Ollama...")

	// Start ollama serve in background
	go func() {
		cmd := exec.Command("ollama", "serve")
		err := cmd.Start()

		if err != nil {
			logger.Error("Failed to start Ollama", "error", err)
			glib.IdleAdd(func() {
				w.showToast("Failed to start Ollama: " + err.Error())
			})
			return
		}

		// Wait a bit for Ollama to start
		time.Sleep(2 * time.Second)

		// Check health and update UI on main thread
		glib.IdleAdd(func() {
			w.checkOllamaHealth()
			if w.ollamaHealthy {
				logger.Info("Ollama started successfully")
				w.showToast("Ollama started successfully!")
				w.toastOverlay.SetChild(w.splitView)
			}
		})
	}()
}

func (w *MainWindow) onToggleSidebar() {
	w.appConfig.SidebarVisible = !w.appConfig.SidebarVisible
	w.splitView.SetCollapsed(!w.appConfig.SidebarVisible)
	w.appConfig.Save()
}

func (w *MainWindow) onSettings() {
	// Build model names list
	modelNames := make([]string, len(w.models))
	for i, m := range w.models {
		modelNames[i] = m.Name
	}

	dialog := NewSettingsDialog(&w.ApplicationWindow.Window, w.appConfig, modelNames)
	dialog.OnSave(func(cfg *config.AppConfig) {
		w.appConfig = cfg
		w.chatView.SetAppConfig(cfg)

		// Aplicar modelo por defecto inmediatamente si está configurado
		if cfg.DefaultModel != "" {
			w.chatView.GetInputArea().SetModel(cfg.DefaultModel)
			w.chatView.SetModel(cfg.DefaultModel)
		}

		w.showToast("Settings saved")
		logger.Info("Settings saved", "defaultModel", cfg.DefaultModel, "language", cfg.ResponseLanguage)
	})
	dialog.Present()
}
