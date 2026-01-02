package rag

import (
	"strings"
	"unicode"
)

// Chunker splits text into overlapping chunks for context windows.
type Chunker struct {
	chunkSize int
	overlap   int
}

// NewChunker creates a new chunker with specified chunk size and overlap.
// Overlap is clamped to at most 50% of chunk size.
func NewChunker(chunkSize, overlap int) *Chunker {
	if chunkSize < 1 {
		chunkSize = 1024
	}
	maxOverlap := chunkSize / 2
	if overlap > maxOverlap {
		overlap = maxOverlap
	}
	if overlap < 0 {
		overlap = 0
	}
	return &Chunker{
		chunkSize: chunkSize,
		overlap:   overlap,
	}
}

// Chunk splits text into overlapping chunks, preferring natural break points.
func (c *Chunker) Chunk(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if len(text) <= c.chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + c.chunkSize
		if end >= len(text) {
			// Last chunk
			chunks = append(chunks, strings.TrimSpace(text[start:]))
			break
		}

		// Find best break point
		breakPoint := c.findBreakPoint(text, start, end)
		chunk := strings.TrimSpace(text[start:breakPoint])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move start, accounting for overlap
		start = breakPoint - c.overlap
		if start < 0 {
			start = 0
		}
		// Ensure we make progress
		if start >= breakPoint {
			start = breakPoint
		}
	}

	return chunks
}

// findBreakPoint finds the best position to break text between start and end.
// Prefers paragraph breaks, then sentence ends, then word boundaries.
func (c *Chunker) findBreakPoint(text string, start, end int) int {
	if end >= len(text) {
		return len(text)
	}

	segment := text[start:end]

	// Try to find paragraph break (double newline)
	if idx := strings.LastIndex(segment, "\n\n"); idx > len(segment)/2 {
		return start + idx + 2
	}

	// Try to find sentence end
	for _, sep := range []string{". ", "! ", "? ", ".\n", "!\n", "?\n"} {
		if idx := strings.LastIndex(segment, sep); idx > len(segment)/2 {
			return start + idx + len(sep)
		}
	}

	// Try to find newline
	if idx := strings.LastIndex(segment, "\n"); idx > len(segment)/2 {
		return start + idx + 1
	}

	// Try to find word boundary (space)
	if idx := strings.LastIndex(segment, " "); idx > len(segment)/4 {
		return start + idx + 1
	}

	// No good break point found, break at end
	return end
}

// ChunkWithMetadata returns chunks with position information.
type ChunkInfo struct {
	Text  string
	Start int
	End   int
	Index int
}

// ChunkWithInfo splits text into chunks with position metadata.
func (c *Chunker) ChunkWithInfo(text string) []ChunkInfo {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	if len(text) <= c.chunkSize {
		return []ChunkInfo{{
			Text:  text,
			Start: 0,
			End:   len(text),
			Index: 0,
		}}
	}

	var chunks []ChunkInfo
	start := 0
	index := 0

	for start < len(text) {
		end := start + c.chunkSize
		if end >= len(text) {
			chunks = append(chunks, ChunkInfo{
				Text:  strings.TrimSpace(text[start:]),
				Start: start,
				End:   len(text),
				Index: index,
			})
			break
		}

		breakPoint := c.findBreakPoint(text, start, end)
		chunk := strings.TrimSpace(text[start:breakPoint])
		if chunk != "" {
			chunks = append(chunks, ChunkInfo{
				Text:  chunk,
				Start: start,
				End:   breakPoint,
				Index: index,
			})
			index++
		}

		start = breakPoint - c.overlap
		if start < 0 {
			start = 0
		}
		if start >= breakPoint {
			start = breakPoint
		}
	}

	return chunks
}

// EstimateTokens provides a rough token count estimate (chars / 4).
func EstimateTokens(text string) int {
	// Rough approximation: ~4 characters per token for English
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return (count + 3) / 4
}
