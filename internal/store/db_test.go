package store

import (
	"testing"
)

func TestNewDB(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("NewDB() returned nil")
	}
}

func TestDB_SchemaCreated(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	// Verify chats table exists
	var tableName string
	err = db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='chats'").Scan(&tableName)
	if err != nil {
		t.Errorf("chats table not found: %v", err)
	}

	// Verify messages table exists
	err = db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&tableName)
	if err != nil {
		t.Errorf("messages table not found: %v", err)
	}

	// Verify attachments table exists
	err = db.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='attachments'").Scan(&tableName)
	if err != nil {
		t.Errorf("attachments table not found: %v", err)
	}
}

func TestDB_CreateChat(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, err := db.CreateChat("llama3")
	if err != nil {
		t.Fatalf("CreateChat() error = %v", err)
	}

	if chat.ID == 0 {
		t.Error("CreateChat() did not set ID")
	}

	if chat.Model != "llama3" {
		t.Errorf("CreateChat() model = %q, want %q", chat.Model, "llama3")
	}

	if chat.Title != "New Chat" {
		t.Errorf("CreateChat() title = %q, want %q", chat.Title, "New Chat")
	}
}

func TestDB_GetChat(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	created, _ := db.CreateChat("llama3")

	chat, err := db.GetChat(created.ID)
	if err != nil {
		t.Fatalf("GetChat() error = %v", err)
	}

	if chat.ID != created.ID {
		t.Errorf("GetChat() ID = %d, want %d", chat.ID, created.ID)
	}
}

func TestDB_ListChats(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	// Create some chats
	db.CreateChat("llama3")
	db.CreateChat("mistral")

	chats, err := db.ListChats()
	if err != nil {
		t.Fatalf("ListChats() error = %v", err)
	}

	if len(chats) != 2 {
		t.Errorf("ListChats() returned %d chats, want 2", len(chats))
	}
}

func TestDB_UpdateChatTitle(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, _ := db.CreateChat("llama3")

	err = db.UpdateChatTitle(chat.ID, "My Custom Title")
	if err != nil {
		t.Fatalf("UpdateChatTitle() error = %v", err)
	}

	updated, _ := db.GetChat(chat.ID)
	if updated.Title != "My Custom Title" {
		t.Errorf("UpdateChatTitle() title = %q, want %q", updated.Title, "My Custom Title")
	}
}

func TestDB_DeleteChat(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, _ := db.CreateChat("llama3")

	err = db.DeleteChat(chat.ID)
	if err != nil {
		t.Fatalf("DeleteChat() error = %v", err)
	}

	_, err = db.GetChat(chat.ID)
	if err == nil {
		t.Error("DeleteChat() did not delete the chat")
	}
}

func TestDB_AddMessage(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, _ := db.CreateChat("llama3")

	msg, err := db.AddMessage(chat.ID, RoleUser, "Hello, world!")
	if err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	if msg.ID == 0 {
		t.Error("AddMessage() did not set ID")
	}

	if msg.Content != "Hello, world!" {
		t.Errorf("AddMessage() content = %q, want %q", msg.Content, "Hello, world!")
	}
}

func TestDB_GetMessages(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, _ := db.CreateChat("llama3")
	db.AddMessage(chat.ID, RoleUser, "Hello")
	db.AddMessage(chat.ID, RoleAssistant, "Hi there!")

	messages, err := db.GetMessages(chat.ID)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("GetMessages() returned %d messages, want 2", len(messages))
	}

	// Should be in order
	if messages[0].Role != RoleUser {
		t.Errorf("First message role = %q, want %q", messages[0].Role, RoleUser)
	}
}

func TestDB_CascadeDelete(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	chat, _ := db.CreateChat("llama3")
	db.AddMessage(chat.ID, RoleUser, "Hello")

	// Delete chat should cascade to messages
	db.DeleteChat(chat.ID)

	messages, _ := db.GetMessages(chat.ID)
	if len(messages) != 0 {
		t.Errorf("Messages should be deleted with chat, got %d", len(messages))
	}
}
