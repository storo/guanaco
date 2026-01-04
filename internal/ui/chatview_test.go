package ui

import (
	"sync"
	"testing"
	"time"
)

func TestGetGreeting(t *testing.T) {
	greeting := getGreeting()
	if greeting == "" {
		t.Error("getGreeting() returned empty string")
	}

	// Should be one of the expected greetings
	validGreetings := []string{
		"Good morning",
		"Good afternoon",
		"Good evening",
		"Buenos días",
		"Buenas tardes",
		"Buenas noches",
	}

	found := false
	for _, valid := range validGreetings {
		if greeting == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("getGreeting() = %q, expected one of %v", greeting, validGreetings)
	}
}

func TestGetUsername(t *testing.T) {
	// Just verify it doesn't panic and returns some value
	username := getUsername()
	// Username can be empty if user info is unavailable, which is fine
	_ = username
}

func TestExtractUserText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "with attachment and double newline",
			input: "[📎 document.pdf]\n\nHello world",
			want:  "Hello world",
		},
		{
			name:  "with attachment no double newline",
			input: "[📎 image.png] Some text",
			want:  "Some text",
		},
		{
			name:  "only attachment",
			input: "[📎 file.txt]",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "text starting with bracket but not attachment",
			input: "[some text] more text",
			want:  "[some text] more text",
		},
		{
			name:  "multiple attachments",
			input: "[📎 file1.txt]\n\n[📎 file2.txt]\n\nActual text",
			want:  "[📎 file2.txt]\n\nActual text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUserText(tt.input)
			if got != tt.want {
				t.Errorf("extractUserText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenBuffer(t *testing.T) {
	t.Run("accumulates and flushes content", func(t *testing.T) {
		var flushed []string
		var mu sync.Mutex

		buffer := newTokenBuffer(20*time.Millisecond, func(content string) {
			mu.Lock()
			flushed = append(flushed, content)
			mu.Unlock()
		})

		// Write content
		buffer.Write("Hello")
		buffer.Write("Hello World")
		buffer.Write("Hello World!")

		// Wait for at least one flush
		time.Sleep(50 * time.Millisecond)

		buffer.Stop()

		mu.Lock()
		defer mu.Unlock()

		if len(flushed) == 0 {
			t.Error("expected at least one flush")
		}

		// Last flushed content should be the final accumulated content
		lastContent := flushed[len(flushed)-1]
		if lastContent != "Hello World!" {
			t.Errorf("last flushed content = %q, want %q", lastContent, "Hello World!")
		}
	})

	t.Run("stop triggers final flush", func(t *testing.T) {
		var lastFlushed string
		var mu sync.Mutex

		buffer := newTokenBuffer(1*time.Hour, func(content string) {
			mu.Lock()
			lastFlushed = content
			mu.Unlock()
		})

		buffer.Write("Final content")
		buffer.Stop()

		// Give goroutine time to execute final flush
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()

		if lastFlushed != "Final content" {
			t.Errorf("final flush content = %q, want %q", lastFlushed, "Final content")
		}
	})

	t.Run("empty content not flushed", func(t *testing.T) {
		flushCount := 0
		var mu sync.Mutex

		buffer := newTokenBuffer(10*time.Millisecond, func(content string) {
			mu.Lock()
			flushCount++
			mu.Unlock()
		})

		// Wait without writing anything
		time.Sleep(30 * time.Millisecond)
		buffer.Stop()

		mu.Lock()
		defer mu.Unlock()

		if flushCount != 0 {
			t.Errorf("flush count = %d, want 0 for empty buffer", flushCount)
		}
	})
}
