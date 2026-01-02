package rag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTxtReader_Read(t *testing.T) {
	reader := NewTxtReader()

	t.Run("read sample file", func(t *testing.T) {
		content, err := reader.Read("testdata/sample.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(content, "sample text file") {
			t.Errorf("expected content to contain 'sample text file', got: %s", content)
		}

		if !strings.Contains(content, "áéíóú") {
			t.Errorf("expected content to contain UTF-8 characters, got: %s", content)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := reader.Read("testdata/nonexistent.txt")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("read empty file", func(t *testing.T) {
		// Create temp empty file
		tmpFile, err := os.CreateTemp("", "empty*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		content, err := reader.Read(tmpFile.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if content != "" {
			t.Errorf("expected empty content, got: %s", content)
		}
	})
}

func TestTxtReader_CanRead(t *testing.T) {
	reader := NewTxtReader()

	tests := []struct {
		filename string
		expected bool
	}{
		{"document.txt", true},
		{"document.TXT", true},
		{"document.text", true},
		{"document.md", true},
		{"document.markdown", true},
		{"document.pdf", false},
		{"document.doc", false},
		{"document.docx", false},
		{"document", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := reader.CanRead(tt.filename)
			if result != tt.expected {
				t.Errorf("CanRead(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func BenchmarkTxtReader_Read(b *testing.B) {
	// Create a larger test file
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.txt")

	content := strings.Repeat("Lorem ipsum dolor sit amet. ", 1000)
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to create test file: %v", err)
	}

	reader := NewTxtReader()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = reader.Read(tmpFile)
	}
}
