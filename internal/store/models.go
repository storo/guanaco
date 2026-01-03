// Package store provides data persistence using SQLite.
package store

import "time"

// Role represents the sender of a message in a chat.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Chat represents a conversation with the AI.
type Chat struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Model        string    `json:"model"`
	SystemPrompt string    `json:"system_prompt"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Message represents a single message in a chat.
type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Attachment represents a file attached to a message.
type Attachment struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"message_id"`
	Filename  string `json:"filename"`
	Content   string `json:"content"`
}

// NewChat creates a new Chat with default values.
func NewChat(model string) *Chat {
	now := time.Now()
	return &Chat{
		Title:     "New Chat",
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewMessage creates a new Message.
func NewMessage(chatID int64, role Role, content string) *Message {
	return &Message{
		ChatID:    chatID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// NewAttachment creates a new Attachment.
func NewAttachment(messageID int64, filename, content string) *Attachment {
	return &Attachment{
		MessageID: messageID,
		Filename:  filename,
		Content:   content,
	}
}
