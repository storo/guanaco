package ui

import (
	"context"
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/ollama"
	"github.com/storo/guanaco/internal/store"
)

// ChatView displays the chat messages and handles interaction.
type ChatView struct {
	*gtk.Box

	// UI components
	scrolled     *gtk.ScrolledWindow
	messagesBox  *gtk.Box
	inputArea    *InputArea

	// State
	messages      []*MessageBubble
	currentBubble *MessageBubble
	isStreaming   bool

	// Dependencies
	ollamaClient *ollama.Client
	streamHandler *ollama.StreamHandler
	db           *store.DB
	currentChat  *store.Chat
	currentModel string

	// Callbacks
	onError func(error)
}

// NewChatView creates a new chat view.
func NewChatView(client *ollama.Client, db *store.DB) *ChatView {
	cv := &ChatView{
		ollamaClient:  client,
		streamHandler: ollama.NewStreamHandler(client),
		db:            db,
	}

	cv.Box = gtk.NewBox(gtk.OrientationVertical, 0)
	cv.SetVExpand(true)
	cv.SetHExpand(true)

	cv.setupUI()

	return cv
}

func (cv *ChatView) setupUI() {
	// Messages area
	cv.messagesBox = gtk.NewBox(gtk.OrientationVertical, 0)
	cv.messagesBox.SetVExpand(true)
	cv.messagesBox.SetMarginTop(8)
	cv.messagesBox.SetMarginBottom(8)

	// Scrolled window for messages
	cv.scrolled = gtk.NewScrolledWindow()
	cv.scrolled.SetChild(cv.messagesBox)
	cv.scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	cv.scrolled.SetVExpand(true)
	cv.Append(cv.scrolled)

	// Separator
	separator := gtk.NewSeparator(gtk.OrientationHorizontal)
	cv.Append(separator)

	// Input area
	cv.inputArea = NewInputArea()
	cv.inputArea.OnSend(cv.onSendMessage)
	cv.Append(cv.inputArea)
}

func (cv *ChatView) onSendMessage(text string) {
	if cv.isStreaming {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return
	}

	// Create chat if needed
	if cv.currentChat == nil {
		cv.createNewChat()
	}

	// Add user message
	cv.addMessage(store.RoleUser, text)

	// Save to database
	if cv.db != nil && cv.currentChat != nil {
		cv.db.AddMessage(cv.currentChat.ID, store.RoleUser, text)
	}

	// Start streaming response
	cv.startStreaming(text)
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
}

func (cv *ChatView) addMessage(role store.Role, content string) *MessageBubble {
	bubble := NewMessageBubble(role, content)
	cv.messages = append(cv.messages, bubble)
	cv.messagesBox.Append(bubble)
	cv.scrollToBottom()
	return bubble
}

func (cv *ChatView) startStreaming(userMessage string) {
	cv.isStreaming = true
	cv.inputArea.SetInputSensitive(false)

	// Create placeholder for response
	cv.currentBubble = cv.addMessage(store.RoleAssistant, "")

	// Build message history
	messages := cv.buildMessageHistory()
	messages = append(messages, ollama.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Start streaming in goroutine
	go func() {
		var response strings.Builder

		ctx := context.Background()
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
			cv.isStreaming = false
			cv.inputArea.SetInputSensitive(true)
			cv.inputArea.Focus()

			if err != nil {
				cv.handleError(err)
				return
			}

			// Save assistant response to database
			finalContent := response.String()
			if cv.db != nil && cv.currentChat != nil && finalContent != "" {
				cv.db.AddMessage(cv.currentChat.ID, store.RoleAssistant, finalContent)
			}
		})
	}()
}

func (cv *ChatView) buildMessageHistory() []ollama.Message {
	var messages []ollama.Message

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

func (cv *ChatView) scrollToBottom() {
	adj := cv.scrolled.VAdjustment()
	adj.SetValue(adj.Upper() - adj.PageSize())
}

func (cv *ChatView) handleError(err error) {
	if cv.onError != nil {
		cv.onError(err)
	}
}

// SetModel sets the current model for chat.
func (cv *ChatView) SetModel(model string) {
	cv.currentModel = model
}

// SetChat loads an existing chat.
func (cv *ChatView) SetChat(chat *store.Chat) {
	cv.currentChat = chat
	cv.currentModel = chat.Model
	cv.clearMessages()

	if cv.db == nil {
		return
	}

	// Load messages from database
	messages, err := cv.db.GetMessages(chat.ID)
	if err != nil {
		cv.handleError(err)
		return
	}

	for _, msg := range messages {
		cv.addMessage(msg.Role, msg.Content)
	}
}

// NewChat starts a new chat.
func (cv *ChatView) NewChat() {
	cv.currentChat = nil
	cv.clearMessages()
}

func (cv *ChatView) clearMessages() {
	for _, bubble := range cv.messages {
		cv.messagesBox.Remove(bubble)
	}
	cv.messages = nil
	cv.currentBubble = nil
}

// OnError sets the error callback.
func (cv *ChatView) OnError(callback func(error)) {
	cv.onError = callback
}

// IsStreaming returns whether a response is currently streaming.
func (cv *ChatView) IsStreaming() bool {
	return cv.isStreaming
}
