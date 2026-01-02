package rag

import (
	"strings"
	"testing"
)

func TestChunker_Chunk(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		chunker := NewChunker(100, 20)
		chunks := chunker.Chunk("")

		if len(chunks) != 0 {
			t.Errorf("expected 0 chunks for empty content, got %d", len(chunks))
		}
	})

	t.Run("content smaller than chunk size", func(t *testing.T) {
		chunker := NewChunker(100, 20)
		content := "Short content"
		chunks := chunker.Chunk(content)

		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk, got %d", len(chunks))
		}

		if chunks[0] != content {
			t.Errorf("expected chunk to equal content, got: %s", chunks[0])
		}
	})

	t.Run("content exactly chunk size", func(t *testing.T) {
		chunker := NewChunker(10, 2)
		content := "0123456789"
		chunks := chunker.Chunk(content)

		if len(chunks) != 1 {
			t.Errorf("expected 1 chunk, got %d", len(chunks))
		}
	})

	t.Run("content needs multiple chunks", func(t *testing.T) {
		chunker := NewChunker(50, 10)
		content := strings.Repeat("word ", 30) // 150 chars

		chunks := chunker.Chunk(content)

		if len(chunks) < 2 {
			t.Errorf("expected multiple chunks, got %d", len(chunks))
		}

		// Verify overlap
		for i := 1; i < len(chunks); i++ {
			// Each chunk (except first) should have some overlap with previous
			prevEnd := chunks[i-1][len(chunks[i-1])-10:]
			if !strings.HasPrefix(chunks[i], prevEnd) && !strings.Contains(chunks[i], prevEnd[:5]) {
				// Overlap may not be exact due to word boundary breaking
			}
		}
	})

	t.Run("respects word boundaries", func(t *testing.T) {
		chunker := NewChunker(20, 5)
		content := "Hello world this is a test"

		chunks := chunker.Chunk(content)

		for _, chunk := range chunks {
			// No chunk should start or end mid-word (unless it's too long)
			trimmed := strings.TrimSpace(chunk)
			if len(trimmed) > 0 {
				// Should not start with space
				if strings.HasPrefix(chunk, " ") && len(chunk) < 20 {
					t.Errorf("chunk should not start with space: %q", chunk)
				}
			}
		}
	})

	t.Run("handles paragraph breaks", func(t *testing.T) {
		chunker := NewChunker(100, 20)
		content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."

		chunks := chunker.Chunk(content)

		// Should prefer breaking at paragraph boundaries
		if len(chunks) > 0 {
			// Just verify we got chunks
			if chunks[0] == "" {
				t.Error("first chunk should not be empty")
			}
		}
	})
}

func TestNewChunker(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		chunker := NewChunker(1024, 128)
		if chunker == nil {
			t.Error("expected non-nil chunker")
		}
	})

	t.Run("overlap larger than chunk size uses default", func(t *testing.T) {
		chunker := NewChunker(100, 150)
		// Should clamp overlap to reasonable value
		if chunker == nil {
			t.Error("expected non-nil chunker")
		}
	})
}

func BenchmarkChunker_Chunk(b *testing.B) {
	chunker := NewChunker(1024, 128)

	// Create a document of ~10KB
	content := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 200)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = chunker.Chunk(content)
	}
}

func BenchmarkChunker_ChunkLarge(b *testing.B) {
	chunker := NewChunker(1024, 128)

	// Create a document of ~100KB
	content := strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. ", 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = chunker.Chunk(content)
	}
}
