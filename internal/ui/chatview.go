package ui

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/config"
	"github.com/storo/guanaco/internal/logger"
	"github.com/storo/guanaco/internal/ollama"
	"github.com/storo/guanaco/internal/rag"
	"github.com/storo/guanaco/internal/store"
)

// getGreeting returns a greeting based on the current time of day.
func getGreeting() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "Buenos días"
	case hour >= 12 && hour < 19:
		return "Buenas tardes"
	default:
		return "Buenas noches"
	}
}

// getUsername returns the current user's first name or username.
func getUsername() string {
	if u, err := user.Current(); err == nil {
		if u.Name != "" {
			return strings.Split(u.Name, " ")[0] // First name only
		}
		return u.Username
	}
	return ""
}

// ChatView displays the chat messages and handles interaction.
type ChatView struct {
	*gtk.Box

	// UI components
	scrolled    *gtk.ScrolledWindow
	messagesBox *gtk.Box
	welcomeView *gtk.Box
	loadingView *gtk.Box
	inputArea   *InputArea

	// State
	messages       []*MessageBubble
	currentBubble  *MessageBubble
	isStreaming    bool
	streamCancel   context.CancelFunc
	userAtBottom   bool // Track if user is at bottom for auto-scroll
	showingWelcome bool // Track if welcome view is showing

	// Dependencies
	ollamaClient  *ollama.Client
	streamHandler *ollama.StreamHandler
	db            *store.DB
	ragProcessor  *rag.Processor
	currentChat   *store.Chat
	currentModel  string
	appConfig     *config.AppConfig

	// Callbacks
	onError        func(error)
	onTitleChanged func(string)
	onChatCreated  func(*store.Chat)
}

// NewChatView creates a new chat view.
func NewChatView(client *ollama.Client, db *store.DB) *ChatView {
	cv := &ChatView{
		ollamaClient:   client,
		streamHandler:  ollama.NewStreamHandler(client),
		db:             db,
		ragProcessor:   rag.NewProcessor(),
		userAtBottom:   true, // Start at bottom
		showingWelcome: true, // Start showing welcome view
	}

	cv.Box = gtk.NewBox(gtk.OrientationVertical, 0)
	cv.SetVExpand(true)
	cv.SetHExpand(true)

	cv.setupUI()
	cv.setupDropTarget()
	cv.setupScrollTracking()

	return cv
}

func (cv *ChatView) setupUI() {
	// Messages area
	cv.messagesBox = gtk.NewBox(gtk.OrientationVertical, 0)
	cv.messagesBox.SetVExpand(true)
	cv.messagesBox.SetMarginTop(8)
	cv.messagesBox.SetMarginBottom(16) // Extra space at bottom for comfortable reading

	// Welcome view for empty chats (professional layout)
	cv.welcomeView = gtk.NewBox(gtk.OrientationVertical, 8)
	cv.welcomeView.SetVExpand(true)
	cv.welcomeView.SetVAlign(gtk.AlignCenter)
	cv.welcomeView.SetHAlign(gtk.AlignCenter)
	cv.welcomeView.SetMarginStart(32)
	cv.welcomeView.SetMarginEnd(32)

	// Logo with fixed pixel size using GtkImage
	logoPath := "/home/storo/projects/guanaco/assets/icons/guanaco-logo.svg"
	logoImage := gtk.NewImageFromFile(logoPath)
	logoImage.SetPixelSize(160)
	logoImage.SetHAlign(gtk.AlignCenter)
	logoImage.AddCSSClass("welcome-logo")
	cv.welcomeView.Append(logoImage)

	// Dynamic greeting based on time of day
	greeting := getGreeting()
	username := getUsername()
	greetingText := greeting
	if username != "" {
		greetingText = fmt.Sprintf("%s, %s", greeting, username)
	}
	greetingLabel := gtk.NewLabel(greetingText)
	greetingLabel.AddCSSClass("title-1")
	greetingLabel.SetHAlign(gtk.AlignCenter)
	greetingLabel.SetMarginTop(8)
	cv.welcomeView.Append(greetingLabel)

	// Horizontal pills for suggestions
	pillsBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	pillsBox.SetHAlign(gtk.AlignCenter)
	pillsBox.SetMarginTop(24)

	// Helper function to create simple pills (icon + title)
	createPill := func(icon, title string) *gtk.Button {
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		btn.AddCSSClass("suggestion-pill")

		box := gtk.NewBox(gtk.OrientationHorizontal, 6)

		iconLabel := gtk.NewLabel(icon)
		box.Append(iconLabel)

		titleLabel := gtk.NewLabel(title)
		box.Append(titleLabel)

		btn.SetChild(box)
		return btn
	}

	pillsBox.Append(createPill("💡", "Explícame"))
	pillsBox.Append(createPill("💻", "Escribe"))
	pillsBox.Append(createPill("📝", "Resume"))
	pillsBox.Append(createPill("🌐", "Traduce"))

	cv.welcomeView.Append(pillsBox)

	// Loading view with spinner
	cv.loadingView = gtk.NewBox(gtk.OrientationVertical, 12)
	cv.loadingView.SetVExpand(true)
	cv.loadingView.SetVAlign(gtk.AlignCenter)
	cv.loadingView.SetHAlign(gtk.AlignCenter)

	spinner := gtk.NewSpinner()
	spinner.SetSizeRequest(32, 32)
	spinner.Start()
	cv.loadingView.Append(spinner)

	loadingLabel := gtk.NewLabel("Cargando...")
	loadingLabel.AddCSSClass("dim-label")
	cv.loadingView.Append(loadingLabel)

	// Scrolled window for messages (starts with welcome view)
	cv.scrolled = gtk.NewScrolledWindow()
	cv.scrolled.SetChild(cv.welcomeView)
	cv.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	cv.scrolled.SetVExpand(true)
	cv.Append(cv.scrolled)

	// Separator
	separator := gtk.NewSeparator(gtk.OrientationHorizontal)
	cv.Append(separator)

	// Input area
	cv.inputArea = NewInputArea()
	cv.inputArea.OnSend(cv.onSendMessage)
	cv.inputArea.OnAttach(cv.onAttachFile)
	cv.inputArea.OnStop(cv.StopStreaming)
	cv.Append(cv.inputArea)
}

func (cv *ChatView) setupDropTarget() {
	// Create drop target for files
	dropTarget := gtk.NewDropTarget(gio.GTypeFile, gdk.ActionCopy)

	dropTarget.ConnectDrop(func(value *glib.Value, x, y float64) bool {
		file := value.Object()
		if file == nil {
			return false
		}

		gfile, ok := file.Cast().(*gio.File)
		if !ok {
			return false
		}

		path := gfile.Path()
		if path == "" {
			return false
		}

		cv.processAndAttachFile(path)
		return true
	})

	cv.AddController(dropTarget)
}

func (cv *ChatView) onAttachFile() {
	// Get parent window
	var parentWindow *gtk.Window
	if root := cv.Root(); root != nil {
		if nw, ok := root.CastType(gtk.GTypeWindow).(*gtk.Window); ok {
			parentWindow = nw
		}
	}

	// Create file chooser dialog
	dialog := gtk.NewFileChooserNative(
		"Select Document",
		parentWindow,
		gtk.FileChooserActionOpen,
		"Open",
		"Cancel",
	)

	// Add file filters
	allFilter := gtk.NewFileFilter()
	allFilter.SetName("All Supported Files")
	allFilter.AddPattern("*.txt")
	allFilter.AddPattern("*.md")
	allFilter.AddPattern("*.pdf")
	allFilter.AddPattern("*.jpg")
	allFilter.AddPattern("*.jpeg")
	allFilter.AddPattern("*.png")
	allFilter.AddPattern("*.webp")
	allFilter.AddPattern("*.gif")
	dialog.AddFilter(allFilter)

	imageFilter := gtk.NewFileFilter()
	imageFilter.SetName("Images")
	imageFilter.AddPattern("*.jpg")
	imageFilter.AddPattern("*.jpeg")
	imageFilter.AddPattern("*.png")
	imageFilter.AddPattern("*.webp")
	imageFilter.AddPattern("*.gif")
	dialog.AddFilter(imageFilter)

	textFilter := gtk.NewFileFilter()
	textFilter.SetName("Text Files")
	textFilter.AddPattern("*.txt")
	textFilter.AddPattern("*.md")
	dialog.AddFilter(textFilter)

	pdfFilter := gtk.NewFileFilter()
	pdfFilter.SetName("PDF Documents")
	pdfFilter.AddPattern("*.pdf")
	dialog.AddFilter(pdfFilter)

	dialog.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseAccept) {
			file := dialog.File()
			if file != nil {
				path := file.Path()
				if path != "" {
					cv.processAndAttachFile(path)
				}
			}
		}
		dialog.Destroy()
	})

	dialog.Show()
}

func (cv *ChatView) processAndAttachFile(path string) {
	filename := filepath.Base(path)
	logger.Info("Processing file attachment", "path", path)

	// Check if file type is supported
	if !cv.ragProcessor.CanProcess(filename) {
		cv.handleError(fmt.Errorf("unsupported file type: %s", filename))
		return
	}

	// Process in background
	go func() {
		result, err := cv.ragProcessor.Process(path)

		glib.IdleAdd(func() {
			if err != nil {
				cv.handleError(fmt.Errorf("failed to process %s: %w", filename, err))
				return
			}

			logger.Info("File processed successfully", "filename", result.Filename, "tokens", result.TokenEstimate)
			// Create and add attachment pill
			pill := NewAttachmentPill(result.Filename, result.Content)
			cv.inputArea.AddAttachment(pill)
		})
	}()
}

func (cv *ChatView) onSendMessage(text string) {
	if cv.isStreaming {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" && !cv.inputArea.HasAttachments() {
		return
	}

	// Validate model is selected
	if cv.currentModel == "" {
		cv.handleError(fmt.Errorf("please enter a model name (e.g., llama3.2)"))
		return
	}

	// Build full prompt with attachments
	data := cv.buildPromptWithAttachments(text)

	// Create chat if needed
	if cv.currentChat == nil {
		cv.createNewChat()
	}

	// Add user message (show original text in bubble, but send full prompt)
	displayText := text
	if cv.inputArea.HasAttachments() {
		attachmentNames := make([]string, 0)
		for _, pill := range cv.inputArea.GetAttachments() {
			attachmentNames = append(attachmentNames, pill.Filename())
		}
		if text != "" {
			displayText = fmt.Sprintf("[📎 %s]\n\n%s", strings.Join(attachmentNames, ", "), text)
		} else {
			displayText = fmt.Sprintf("[📎 %s]", strings.Join(attachmentNames, ", "))
		}
	}
	cv.addMessage(store.RoleUser, displayText)

	// Get attachments before clearing (need for DB save)
	attachments := cv.inputArea.GetAttachments()

	// Clear attachments after using them
	cv.inputArea.ClearAttachments()

	// Save to database with attachments
	if cv.db != nil && cv.currentChat != nil {
		msg, err := cv.db.AddMessage(cv.currentChat.ID, store.RoleUser, displayText)
		if err == nil && len(attachments) > 0 {
			for _, pill := range attachments {
				err := cv.db.AddAttachment(msg.ID, pill.Filename(), pill.Content())
				if err != nil {
					logger.Error("Failed to save attachment", "filename", pill.Filename(), "error", err)
				} else {
					logger.Info("Attachment saved", "messageID", msg.ID, "filename", pill.Filename(), "contentLen", len(pill.Content()))
				}
			}
		}
	}

	// Check if model exists, pull if needed, then stream
	cv.ensureModelAndStream(data)
}

// attachmentData holds parsed attachment information.
type attachmentData struct {
	textContent string
	images      []string
}

func (cv *ChatView) buildPromptWithAttachments(userText string) attachmentData {
	attachments := cv.inputArea.GetAttachments()
	if len(attachments) == 0 {
		return attachmentData{textContent: userText}
	}

	var builder strings.Builder
	var images []string

	// Separate images from documents
	for _, pill := range attachments {
		if pill.IsImage() {
			images = append(images, pill.Content())
		} else {
			builder.WriteString(fmt.Sprintf("[Document: %s]\n", pill.Filename()))
			builder.WriteString(pill.Content())
			builder.WriteString("\n\n")
		}
	}

	// Add user's question/message
	if userText != "" {
		if builder.Len() > 0 {
			builder.WriteString("User question: ")
		}
		builder.WriteString(userText)
	}

	return attachmentData{
		textContent: builder.String(),
		images:      images,
	}
}

func (cv *ChatView) ensureModelAndStream(data attachmentData) {
	ctx := context.Background()

	// Check if model exists locally
	if cv.ollamaClient.HasModel(ctx, cv.currentModel) {
		logger.Debug("Model available locally", "model", cv.currentModel)
		cv.startStreaming(data)
		return
	}

	logger.Info("Model not found, pulling", "model", cv.currentModel)

	// Model not found, need to pull it
	cv.isStreaming = true
	cv.inputArea.SetInputSensitive(false)

	// Create a status bubble to show download progress
	cv.currentBubble = cv.addMessage(store.RoleSystem, "Downloading model "+cv.currentModel+"...")

	go func() {
		err := cv.ollamaClient.PullModel(ctx, cv.currentModel, func(status string, completed, total int64) {
			var progressText string
			if total > 0 {
				percent := float64(completed) / float64(total) * 100
				progressText = fmt.Sprintf("Downloading %s: %s (%.1f%%)", cv.currentModel, status, percent)
			} else {
				progressText = fmt.Sprintf("Downloading %s: %s", cv.currentModel, status)
			}

			glib.IdleAdd(func() {
				cv.currentBubble.SetContent(progressText)
				cv.scrollToBottom()
			})
		})

		glib.IdleAdd(func() {
			if err != nil {
				cv.currentBubble.SetContent("Failed to download model: " + err.Error())
				cv.isStreaming = false
				cv.inputArea.SetInputSensitive(true)
				cv.inputArea.Focus()
				return
			}

			// Remove the download status bubble
			cv.messagesBox.Remove(cv.currentBubble)
			// Remove from messages slice
			for i, bubble := range cv.messages {
				if bubble == cv.currentBubble {
					cv.messages = append(cv.messages[:i], cv.messages[i+1:]...)
					break
				}
			}
			cv.currentBubble = nil
			cv.isStreaming = false

			// Now start the actual chat
			cv.startStreaming(data)
		})
	}()
}

func (cv *ChatView) createNewChat() {
	if cv.db == nil {
		cv.currentChat = &store.Chat{Model: cv.currentModel}
		return
	}

	model := cv.currentModel
	if model == "" {
		model = "llama3"
	}

	chat, err := cv.db.CreateChat(model)
	if err != nil {
		cv.handleError(err)
		return
	}
	cv.currentChat = chat

	// Notify that a new chat was created
	if cv.onChatCreated != nil {
		cv.onChatCreated(chat)
	}
}

func (cv *ChatView) addMessage(role store.Role, content string) *MessageBubble {
	// Switch from welcome view to messages on first message
	if cv.showingWelcome {
		cv.scrolled.SetChild(cv.messagesBox)
		cv.showingWelcome = false
	}

	bubble := NewMessageBubble(role, content)
	cv.messages = append(cv.messages, bubble)
	cv.messagesBox.Append(bubble)
	cv.scrollToBottom()
	return bubble
}

func (cv *ChatView) startStreaming(data attachmentData) {
	ctx, cancel := context.WithCancel(context.Background())
	cv.streamCancel = cancel

	cv.isStreaming = true
	cv.inputArea.SetStreamingMode(true)

	// Create placeholder for response
	cv.currentBubble = cv.addMessage(store.RoleAssistant, "")

	// Build message history
	messages := cv.buildMessageHistory()

	// Log what we're sending
	logger.Info("Sending to model", "historyCount", len(messages), "newContentLen", len(data.textContent))
	for i, m := range messages {
		logger.Info("History message", "index", i, "role", m.Role, "contentLen", len(m.Content))
	}

	// Add user message with optional images
	userMsg := ollama.Message{
		Role:    "user",
		Content: data.textContent,
	}
	if len(data.images) > 0 {
		userMsg.Images = data.images
	}
	messages = append(messages, userMsg)

	// Start streaming in goroutine
	go func() {
		var response strings.Builder

		err := cv.streamHandler.Chat(ctx, &ollama.ChatRequest{
			Model:    cv.currentModel,
			Messages: messages,
		}, func(token string) {
			response.WriteString(token)
			content := response.String()

			// Update UI on main thread
			glib.IdleAdd(func() {
				cv.currentBubble.SetContent(content)
				cv.scrollToBottom()
			})
		})

		// Finalize on main thread
		glib.IdleAdd(func() {
			cv.streamCancel = nil
			cv.isStreaming = false
			cv.inputArea.SetStreamingMode(false)
			cv.inputArea.Focus()

			// Don't show error if user cancelled
			if err != nil && err != context.Canceled {
				cv.handleError(err)
				return
			}

			// Save assistant response to database (even if cancelled, save partial)
			finalContent := response.String()
			if cv.db != nil && cv.currentChat != nil && finalContent != "" {
				cv.db.AddMessage(cv.currentChat.ID, store.RoleAssistant, finalContent)

				// Generate title for new chats
				if cv.currentChat.Title == "New Chat" {
					go cv.generateTitle()
				}
			}
		})
	}()
}

// StopStreaming cancels the current streaming response.
func (cv *ChatView) StopStreaming() {
	if cv.streamCancel != nil {
		cv.streamCancel()
	}
}

func (cv *ChatView) buildMessageHistory() []ollama.Message {
	var messages []ollama.Message

	// Build effective system prompt (chat-specific > global, + language instruction)
	chatPrompt := ""
	if cv.currentChat != nil {
		chatPrompt = cv.currentChat.SystemPrompt
	}

	var systemPrompt string
	if cv.appConfig != nil {
		systemPrompt = cv.appConfig.GetEffectiveSystemPrompt(chatPrompt)
	} else if chatPrompt != "" {
		systemPrompt = chatPrompt
	}

	if systemPrompt != "" {
		messages = append(messages, ollama.Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// If we have DB, load messages with attachments for full context
	if cv.db != nil && cv.currentChat != nil {
		dbMessages, err := cv.db.GetMessages(cv.currentChat.ID)
		if err == nil {
			logger.Info("Building message history from DB", "chatID", cv.currentChat.ID, "messageCount", len(dbMessages))
			for _, msg := range dbMessages {
				content := msg.Content

				// For user messages, check if there are attachments
				if msg.Role == store.RoleUser {
					attachments, _ := cv.db.GetMessageAttachments(msg.ID)
					logger.Info("Checking attachments for message", "messageID", msg.ID, "attachmentCount", len(attachments))
					if len(attachments) > 0 {
						content = cv.rebuildContentWithAttachments(msg.Content, attachments)
						logger.Info("Rebuilt content with attachments", "originalLen", len(msg.Content), "newLen", len(content))
					}
				}

				messages = append(messages, ollama.Message{
					Role:    string(msg.Role),
					Content: content,
				})
			}
			return messages
		}
	}

	// Fallback to bubbles in memory (no DB or error)
	for _, bubble := range cv.messages {
		if bubble == cv.currentBubble {
			continue // Skip the current streaming bubble
		}

		role := "user"
		if bubble.GetRole() == store.RoleAssistant {
			role = "assistant"
		} else if bubble.GetRole() == store.RoleSystem {
			role = "system"
		}

		messages = append(messages, ollama.Message{
			Role:    role,
			Content: bubble.GetContent(),
		})
	}

	return messages
}

// rebuildContentWithAttachments reconstructs the full prompt from display text and attachments.
func (cv *ChatView) rebuildContentWithAttachments(displayText string, attachments []store.Attachment) string {
	var builder strings.Builder

	// Add document contents
	for _, att := range attachments {
		builder.WriteString(fmt.Sprintf("[Document: %s]\n", att.Filename))
		builder.WriteString(att.Content)
		builder.WriteString("\n\n")
	}

	// Extract user's actual text (remove the [📎 ...] prefix)
	userText := extractUserText(displayText)
	if userText != "" {
		if builder.Len() > 0 {
			builder.WriteString("User question: ")
		}
		builder.WriteString(userText)
	}

	return builder.String()
}

// extractUserText removes the attachment indicator prefix from display text.
func extractUserText(displayText string) string {
	// Remove "[📎 filename]\n\n" or "[📎 filename]" prefix
	if strings.HasPrefix(displayText, "[📎") {
		if idx := strings.Index(displayText, "]\n\n"); idx != -1 {
			return displayText[idx+3:]
		}
		if idx := strings.Index(displayText, "]"); idx != -1 {
			return strings.TrimSpace(displayText[idx+1:])
		}
	}
	return displayText
}

func (cv *ChatView) scrollToBottom() {
	// Don't auto-scroll if user scrolled up during streaming
	if cv.isStreaming && !cv.userAtBottom {
		return
	}
	adj := cv.scrolled.VAdjustment()
	adj.SetValue(adj.Upper() - adj.PageSize())
}

// setupScrollTracking tracks user scroll position for auto-scroll lock.
func (cv *ChatView) setupScrollTracking() {
	adj := cv.scrolled.VAdjustment()
	adj.ConnectValueChanged(func() {
		// User is at bottom if within 50px of the end
		cv.userAtBottom = adj.Value() >= adj.Upper()-adj.PageSize()-50
	})
}

func (cv *ChatView) handleError(err error) {
	logger.Error("ChatView error", "error", err)
	if cv.onError != nil {
		cv.onError(err)
	}
}

// SetModel sets the current model for chat.
func (cv *ChatView) SetModel(model string) {
	cv.currentModel = model
}

// SetAppConfig sets the application configuration.
func (cv *ChatView) SetAppConfig(cfg *config.AppConfig) {
	cv.appConfig = cfg
}

// SetChat loads an existing chat.
func (cv *ChatView) SetChat(chat *store.Chat) {
	// Skip if already viewing this chat (prevents reload during streaming)
	if cv.currentChat != nil && cv.currentChat.ID == chat.ID {
		return
	}

	cv.currentChat = chat
	cv.currentModel = chat.Model
	cv.inputArea.SetModel(chat.Model)
	cv.clearMessages()

	if cv.db == nil {
		return
	}

	// Show loading spinner
	cv.scrolled.SetChild(cv.loadingView)
	cv.showingWelcome = false // Loading view, not welcome

	// Capture chat ID for the goroutine
	chatID := chat.ID

	// Load messages asynchronously
	go func() {
		messages, err := cv.db.GetMessages(chatID)

		// Update UI on main thread
		glib.IdleAdd(func() {
			// Check if we're still on the same chat
			if cv.currentChat == nil || cv.currentChat.ID != chatID {
				return
			}

			if err != nil {
				cv.handleError(err)
				cv.scrolled.SetChild(cv.welcomeView)
				cv.showingWelcome = true
				return
			}

			// Switch to messages view
			cv.scrolled.SetChild(cv.messagesBox)
			cv.showingWelcome = false

			for _, msg := range messages {
				cv.addMessage(msg.Role, msg.Content)
			}

			// If no messages, show welcome view
			if len(messages) == 0 {
				cv.scrolled.SetChild(cv.welcomeView)
				cv.showingWelcome = true
			}
		})
	}()
}

// NewChat starts a new chat.
func (cv *ChatView) NewChat() {
	cv.currentChat = nil
	cv.clearMessages()
}

// EnsureChat creates a new chat if none exists.
func (cv *ChatView) EnsureChat(model string) {
	if cv.currentChat == nil {
		cv.currentModel = model
		cv.createNewChat()
	}
}

func (cv *ChatView) clearMessages() {
	for _, bubble := range cv.messages {
		cv.messagesBox.Remove(bubble)
	}
	cv.messages = nil
	cv.currentBubble = nil

	// Show welcome view again
	cv.scrolled.SetChild(cv.welcomeView)
	cv.showingWelcome = true
}

// OnError sets the error callback.
func (cv *ChatView) OnError(callback func(error)) {
	cv.onError = callback
}

// IsStreaming returns whether a response is currently streaming.
func (cv *ChatView) IsStreaming() bool {
	return cv.isStreaming
}

// GetCurrentChat returns the current chat.
func (cv *ChatView) GetCurrentChat() *store.Chat {
	return cv.currentChat
}

// GetInputArea returns the input area for external access.
func (cv *ChatView) GetInputArea() *InputArea {
	return cv.inputArea
}

// OnTitleChanged sets the callback for when the chat title changes.
func (cv *ChatView) OnTitleChanged(callback func(string)) {
	cv.onTitleChanged = callback
}

// OnChatCreated sets the callback for when a new chat is created.
func (cv *ChatView) OnChatCreated(callback func(*store.Chat)) {
	cv.onChatCreated = callback
}

// generateTitle asks the model to generate a short title for the conversation.
func (cv *ChatView) generateTitle() {
	if cv.db == nil || cv.currentChat == nil || len(cv.messages) < 2 {
		return
	}

	// Get first user message
	var userMsg string
	for _, bubble := range cv.messages {
		if bubble.GetRole() == store.RoleUser {
			userMsg = bubble.GetContent()
			break
		}
	}

	if userMsg == "" {
		return
	}

	// Truncate if too long
	if len(userMsg) > 200 {
		userMsg = userMsg[:200]
	}

	logger.Info("Generating title for chat", "chatID", cv.currentChat.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := fmt.Sprintf("Generate a very short title (3-5 words max) for a conversation that starts with: %q\nRespond with ONLY the title, nothing else.", userMsg)

	var title strings.Builder
	err := cv.streamHandler.Chat(ctx, &ollama.ChatRequest{
		Model:    cv.currentModel,
		Messages: []ollama.Message{{Role: "user", Content: prompt}},
	}, func(token string) {
		title.WriteString(token)
	})

	if err != nil {
		logger.Error("Failed to generate title", "error", err)
		return
	}

	newTitle := strings.TrimSpace(title.String())
	// Remove quotes if present
	newTitle = strings.Trim(newTitle, "\"'")

	if newTitle == "" || len(newTitle) > 60 {
		return
	}

	// Update in database
	if err := cv.db.UpdateChatTitle(cv.currentChat.ID, newTitle); err != nil {
		logger.Error("Failed to update chat title", "error", err)
		return
	}

	cv.currentChat.Title = newTitle
	logger.Info("Chat title updated", "chatID", cv.currentChat.ID, "title", newTitle)

	// Notify UI on main thread
	glib.IdleAdd(func() {
		if cv.onTitleChanged != nil {
			cv.onTitleChanged(newTitle)
		}
	})
}
