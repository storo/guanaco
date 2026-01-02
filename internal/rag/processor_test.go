package rag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessor_Process(t *testing.T) {
	processor := NewProcessor()

	t.Run("process text file", func(t *testing.T) {
		result, err := processor.Process("testdata/sample.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Filename != "sample.txt" {
			t.Errorf("expected filename 'sample.txt', got %s", result.Filename)
		}

		if result.Content == "" {
			t.Error("expected non-empty content")
		}

		if result.TokenEstimate == 0 {
			t.Error("expected non-zero token estimate")
		}

		if len(result.Chunks) == 0 {
			t.Error("expected at least one chunk")
		}
	})

	t.Run("process non-existent file", func(t *testing.T) {
		_, err := processor.Process("testdata/nonexistent.xyz")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("process unsupported file type", func(t *testing.T) {
		// Create a temp file with unsupported extension
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.xyz")
		if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, err := processor.Process(tmpFile)
		if err == nil {
			t.Error("expected error for unsupported file type")
		}
		if !strings.Contains(err.Error(), "unsupported") {
			t.Errorf("expected 'unsupported' in error, got: %v", err)
		}
	})
}

func TestProcessor_CanProcess(t *testing.T) {
	processor := NewProcessor()

	tests := []struct {
		filename string
		expected bool
	}{
		{"document.txt", true},
		{"document.md", true},
		{"document.pdf", true},
		{"document.doc", false},
		{"document.docx", false},
		{"document.xlsx", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := processor.CanProcess(tt.filename)
			if result != tt.expected {
				t.Errorf("CanProcess(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestProcessor_WithChunkSize(t *testing.T) {
	processor := NewProcessor()
	processor.SetChunkSize(50, 10)

	// Create a file with enough content to chunk
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := strings.Repeat("This is test content. ", 20)
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := processor.Process(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Chunks) < 2 {
		t.Errorf("expected multiple chunks with small chunk size, got %d", len(result.Chunks))
	}
}

func TestDocumentResult(t *testing.T) {
	result := &DocumentResult{
		Filename:      "test.txt",
		Content:       "Hello world",
		Chunks:        []string{"Hello world"},
		TokenEstimate: 3,
	}

	if result.Filename != "test.txt" {
		t.Errorf("expected filename 'test.txt', got %s", result.Filename)
	}
}

func BenchmarkProcessor_Process(b *testing.B) {
	// Create a larger test file
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench.txt")
	content := strings.Repeat("Lorem ipsum dolor sit amet. ", 500)
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		b.Fatalf("failed to create test file: %v", err)
	}

	processor := NewProcessor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = processor.Process(tmpFile)
	}
}
