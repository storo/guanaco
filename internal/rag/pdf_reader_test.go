package rag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPdfReader_CanRead(t *testing.T) {
	reader := NewPdfReader()

	tests := []struct {
		filename string
		expected bool
	}{
		{"document.pdf", true},
		{"document.PDF", true},
		{"document.txt", false},
		{"document.doc", false},
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

func TestPdfReader_Read(t *testing.T) {
	reader := NewPdfReader()

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := reader.Read("testdata/nonexistent.pdf")
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})

	t.Run("read invalid file", func(t *testing.T) {
		// Create a temp file that's not a valid PDF
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.pdf")
		if err := os.WriteFile(tmpFile, []byte("not a pdf"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, err := reader.Read(tmpFile)
		if err == nil {
			t.Error("expected error for invalid PDF file")
		}
	})
}

// TestPdfReader_ReadRealPdf tests with a real PDF if available
func TestPdfReader_ReadRealPdf(t *testing.T) {
	// Skip if no test PDF available
	testPdf := "testdata/sample.pdf"
	if _, err := os.Stat(testPdf); os.IsNotExist(err) {
		t.Skip("skipping: no sample.pdf available for testing")
	}

	reader := NewPdfReader()
	content, err := reader.Read(testPdf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == "" {
		t.Error("expected non-empty content from PDF")
	}

	// Content should have been cleaned up (no excessive whitespace)
	if strings.Contains(content, "\n\n\n") {
		t.Error("content should not have excessive newlines")
	}
}
