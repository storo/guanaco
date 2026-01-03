package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS chats (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    title         TEXT NOT NULL DEFAULT 'New Chat',
    model         TEXT NOT NULL,
    system_prompt TEXT NOT NULL DEFAULT '',
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id     INTEGER NOT NULL,
    role        TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'system')),
    content     TEXT NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS attachments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id  INTEGER NOT NULL,
    filename    TEXT NOT NULL,
    content     TEXT NOT NULL,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_attachments_message_id ON attachments(message_id);
`

// migration adds new columns to existing databases
const migration = `
-- Add system_prompt column if it doesn't exist
ALTER TABLE chats ADD COLUMN system_prompt TEXT NOT NULL DEFAULT '';
`

// DB wraps the SQLite database connection.
type DB struct {
	db *sql.DB

	// Prepared statements for performance
	stmtCreateChat            *sql.Stmt
	stmtGetChat               *sql.Stmt
	stmtListChats             *sql.Stmt
	stmtUpdateChatTitle       *sql.Stmt
	stmtUpdateChatSystemPrompt *sql.Stmt
	stmtDeleteChat            *sql.Stmt
	stmtAddMessage            *sql.Stmt
	stmtGetMessages           *sql.Stmt
}

// NewDB creates a new database connection and initializes the schema.
func NewDB(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite with modernc.org requires single connection for writes
	sqlDB.SetMaxOpenConns(1)

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create schema
	if _, err := sqlDB.Exec(schema); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Run migrations (ignore errors for columns that already exist)
	sqlDB.Exec(migration)

	db := &DB{db: sqlDB}

	// Prepare statements
	if err := db.prepareStatements(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return db, nil
}

func (d *DB) prepareStatements() error {
	var err error

	d.stmtCreateChat, err = d.db.Prepare(`
		INSERT INTO chats (title, model, system_prompt, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare CreateChat: %w", err)
	}

	d.stmtGetChat, err = d.db.Prepare(`
		SELECT id, title, model, system_prompt, created_at, updated_at
		FROM chats WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare GetChat: %w", err)
	}

	d.stmtListChats, err = d.db.Prepare(`
		SELECT id, title, model, system_prompt, created_at, updated_at
		FROM chats ORDER BY updated_at DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare ListChats: %w", err)
	}

	d.stmtUpdateChatTitle, err = d.db.Prepare(`
		UPDATE chats SET title = ?, updated_at = ? WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare UpdateChatTitle: %w", err)
	}

	d.stmtUpdateChatSystemPrompt, err = d.db.Prepare(`
		UPDATE chats SET system_prompt = ?, updated_at = ? WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare UpdateChatSystemPrompt: %w", err)
	}

	d.stmtDeleteChat, err = d.db.Prepare(`DELETE FROM chats WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("failed to prepare DeleteChat: %w", err)
	}

	d.stmtAddMessage, err = d.db.Prepare(`
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare AddMessage: %w", err)
	}

	d.stmtGetMessages, err = d.db.Prepare(`
		SELECT id, chat_id, role, content, created_at
		FROM messages WHERE chat_id = ? ORDER BY created_at ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare GetMessages: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	// Close prepared statements
	if d.stmtCreateChat != nil {
		d.stmtCreateChat.Close()
	}
	if d.stmtGetChat != nil {
		d.stmtGetChat.Close()
	}
	if d.stmtListChats != nil {
		d.stmtListChats.Close()
	}
	if d.stmtUpdateChatTitle != nil {
		d.stmtUpdateChatTitle.Close()
	}
	if d.stmtUpdateChatSystemPrompt != nil {
		d.stmtUpdateChatSystemPrompt.Close()
	}
	if d.stmtDeleteChat != nil {
		d.stmtDeleteChat.Close()
	}
	if d.stmtAddMessage != nil {
		d.stmtAddMessage.Close()
	}
	if d.stmtGetMessages != nil {
		d.stmtGetMessages.Close()
	}

	return d.db.Close()
}

// CreateChat creates a new chat with the given model.
func (d *DB) CreateChat(model string) (*Chat, error) {
	now := time.Now()
	chat := NewChat(model)
	chat.CreatedAt = now
	chat.UpdatedAt = now

	result, err := d.stmtCreateChat.Exec(chat.Title, chat.Model, chat.SystemPrompt, chat.CreatedAt, chat.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	chat.ID = id
	return chat, nil
}

// GetChat retrieves a chat by ID.
func (d *DB) GetChat(id int64) (*Chat, error) {
	chat := &Chat{}
	err := d.stmtGetChat.QueryRow(id).Scan(
		&chat.ID,
		&chat.Title,
		&chat.Model,
		&chat.SystemPrompt,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	return chat, nil
}

// ListChats returns all chats ordered by update time (most recent first).
func (d *DB) ListChats() ([]*Chat, error) {
	rows, err := d.stmtListChats.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		chat := &Chat{}
		err := rows.Scan(
			&chat.ID,
			&chat.Title,
			&chat.Model,
			&chat.SystemPrompt,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, chat)
	}

	return chats, rows.Err()
}

// UpdateChatTitle updates the title of a chat.
func (d *DB) UpdateChatTitle(id int64, title string) error {
	_, err := d.stmtUpdateChatTitle.Exec(title, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update chat title: %w", err)
	}
	return nil
}

// UpdateChatSystemPrompt updates the system prompt of a chat.
func (d *DB) UpdateChatSystemPrompt(id int64, systemPrompt string) error {
	_, err := d.stmtUpdateChatSystemPrompt.Exec(systemPrompt, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update chat system prompt: %w", err)
	}
	return nil
}

// DeleteChat deletes a chat and its messages (cascade).
func (d *DB) DeleteChat(id int64) error {
	_, err := d.stmtDeleteChat.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}
	return nil
}

// AddMessage adds a message to a chat.
func (d *DB) AddMessage(chatID int64, role Role, content string) (*Message, error) {
	now := time.Now()
	msg := &Message{
		ChatID:    chatID,
		Role:      role,
		Content:   content,
		CreatedAt: now,
	}

	result, err := d.stmtAddMessage.Exec(msg.ChatID, msg.Role, msg.Content, msg.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	msg.ID = id
	return msg, nil
}

// GetMessages retrieves all messages for a chat in chronological order.
func (d *DB) GetMessages(chatID int64) ([]*Message, error) {
	rows, err := d.stmtGetMessages.Query(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		msg := &Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.Role,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

// AddAttachment saves an attachment for a message.
func (d *DB) AddAttachment(messageID int64, filename, content string) error {
	_, err := d.db.Exec(
		"INSERT INTO attachments (message_id, filename, content) VALUES (?, ?, ?)",
		messageID, filename, content,
	)
	if err != nil {
		return fmt.Errorf("failed to add attachment: %w", err)
	}
	return nil
}

// GetMessageAttachments returns attachments for a message.
func (d *DB) GetMessageAttachments(messageID int64) ([]Attachment, error) {
	rows, err := d.db.Query(
		"SELECT id, message_id, filename, content FROM attachments WHERE message_id = ?",
		messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}
	defer rows.Close()

	var attachments []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.MessageID, &a.Filename, &a.Content); err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, a)
	}
	return attachments, rows.Err()
}
