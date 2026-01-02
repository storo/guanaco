package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewStreamHandler(t *testing.T) {
	client := NewClientDefault()
	handler := NewStreamHandler(client)

	if handler == nil {
		t.Fatal("NewStreamHandler() returned nil")
	}

	if handler.client != client {
		t.Error("NewStreamHandler() did not set client")
	}
}

func TestStreamHandler_Chat_ReceivesTokens(t *testing.T) {
	// Mock server that streams tokens
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			return
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("server does not support flushing")
		}

		// Send streaming response
		tokens := []string{"Hello", " ", "world", "!"}
		for i, token := range tokens {
			resp := map[string]interface{}{
				"message": map[string]string{
					"role":    "assistant",
					"content": token,
				},
				"done": i == len(tokens)-1,
			}
			data, _ := json.Marshal(resp)
			w.Write(data)
			w.Write([]byte("\n"))
			flusher.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	handler := NewStreamHandler(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var received []string
	var mu sync.Mutex

	err := handler.Chat(ctx, &ChatRequest{
		Model: "test",
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
	}, func(token string) {
		mu.Lock()
		received = append(received, token)
		mu.Unlock()
	})

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	result := strings.Join(received, "")
	if result != "Hello world!" {
		t.Errorf("Chat() received = %q, want %q", result, "Hello world!")
	}
}

func TestStreamHandler_Chat_Cancellation(t *testing.T) {
	// Mock server that streams slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, _ := w.(http.Flusher)

		for i := 0; i < 100; i++ {
			select {
			case <-r.Context().Done():
				return
			default:
				resp := map[string]interface{}{
					"message": map[string]string{
						"role":    "assistant",
						"content": "token",
					},
					"done": false,
				}
				data, _ := json.Marshal(resp)
				w.Write(data)
				w.Write([]byte("\n"))
				flusher.Flush()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	handler := NewStreamHandler(client)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	tokenCount := 0
	err := handler.Chat(ctx, &ChatRequest{
		Model:    "test",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	}, func(token string) {
		tokenCount++
	})

	// Should get context deadline exceeded
	if err == nil {
		t.Error("Chat() should return error on cancellation")
	}

	// Should have received some but not all tokens
	if tokenCount == 0 {
		t.Error("Should have received at least some tokens before cancellation")
	}
	if tokenCount >= 100 {
		t.Error("Should not have received all tokens due to cancellation")
	}
}

func TestStreamHandler_Chat_Error(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	handler := NewStreamHandler(client)

	ctx := context.Background()
	err := handler.Chat(ctx, &ChatRequest{
		Model:    "nonexistent",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	}, func(token string) {})

	if err == nil {
		t.Error("Chat() should return error for 500 response")
	}
}

func TestChatRequest_Validation(t *testing.T) {
	req := &ChatRequest{
		Model: "llama3",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
	}

	if req.Model != "llama3" {
		t.Errorf("Model = %q, want %q", req.Model, "llama3")
	}

	if len(req.Messages) != 2 {
		t.Errorf("Messages length = %d, want 2", len(req.Messages))
	}
}
