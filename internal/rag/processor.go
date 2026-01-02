package rag

import (
	"fmt"
	"path/filepath"
)

// DefaultChunkSize is the default chunk size in characters.
const DefaultChunkSize = 2048

// DefaultOverlap is the default overlap between chunks.
const DefaultOverlap = 256

// DocumentResult contains processed document information.
type DocumentResult struct {
	// Filename is the base name of the processed file.
	Filename string

	// Content is the full extracted text content.
	Content string

	// Chunks are the text split into overlapping segments.
	Chunks []string

	// TokenEstimate is an approximate token count.
	TokenEstimate int
}

// Processor handles document processing for RAG.
type Processor struct {
	readers []Reader
	chunker *Chunker
}

// NewProcessor creates a new document processor with default readers.
func NewProcessor() *Processor {
	return &Processor{
		readers: []Reader{
			NewTxtReader(),
			NewPdfReader(),
		},
		chunker: NewChunker(DefaultChunkSize, DefaultOverlap),
	}
}

// SetChunkSize configures the chunker with new size and overlap.
func (p *Processor) SetChunkSize(size, overlap int) {
	p.chunker = NewChunker(size, overlap)
}

// AddReader adds a custom reader to the processor.
func (p *Processor) AddReader(reader Reader) {
	p.readers = append(p.readers, reader)
}

// CanProcess returns true if any reader can handle the file.
func (p *Processor) CanProcess(filename string) bool {
	for _, reader := range p.readers {
		if reader.CanRead(filename) {
			return true
		}
	}
	return false
}

// Process reads and chunks a document file.
func (p *Processor) Process(path string) (*DocumentResult, error) {
	filename := filepath.Base(path)

	// Find appropriate reader
	var content string
	var err error
	var found bool

	for _, reader := range p.readers {
		if reader.CanRead(filename) {
			content, err = reader.Read(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", filename, err)
			}
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("unsupported file type: %s", filename)
	}

	// Chunk the content
	chunks := p.chunker.Chunk(content)

	return &DocumentResult{
		Filename:      filename,
		Content:       content,
		Chunks:        chunks,
		TokenEstimate: EstimateTokens(content),
	}, nil
}

// ProcessForContext processes a document and formats it for LLM context.
func (p *Processor) ProcessForContext(path string) (string, error) {
	result, err := p.Process(path)
	if err != nil {
		return "", err
	}

	// Format as context block
	return fmt.Sprintf("[Document: %s]\n%s", result.Filename, result.Content), nil
}

// SupportedExtensions returns a list of supported file extensions.
func (p *Processor) SupportedExtensions() []string {
	return []string{".txt", ".text", ".md", ".markdown", ".pdf"}
}
