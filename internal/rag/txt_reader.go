// Package rag provides document processing for RAG (Retrieval Augmented Generation).
package rag

import (
	"os"
	"path/filepath"
	"strings"
)

// Reader is an interface for document readers.
type Reader interface {
	// Read reads content from a file path.
	Read(path string) (string, error)
	// CanRead returns true if this reader can handle the file type.
	CanRead(filename string) bool
}

// TxtReader reads plain text files.
type TxtReader struct{}

// NewTxtReader creates a new text file reader.
func NewTxtReader() *TxtReader {
	return &TxtReader{}
}

// supportedExtensions lists extensions this reader supports.
var txtExtensions = map[string]bool{
	".txt":      true,
	".text":     true,
	".md":       true,
	".markdown": true,
}

// Read reads the content of a text file.
func (r *TxtReader) Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CanRead returns true if the file has a text extension.
func (r *TxtReader) CanRead(filename string) bool {
	if filename == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return txtExtensions[ext]
}
