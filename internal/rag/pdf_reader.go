package rag

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PdfReader reads PDF files.
type PdfReader struct{}

// NewPdfReader creates a new PDF file reader.
func NewPdfReader() *PdfReader {
	return &PdfReader{}
}

// Read extracts text content from a PDF file.
func (r *PdfReader) Read(path string) (string, error) {
	f, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var builder strings.Builder
	totalPages := reader.NumPage()

	for i := 1; i <= totalPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Continue with other pages on error
			continue
		}

		builder.WriteString(text)
		if i < totalPages {
			builder.WriteString("\n\n")
		}
	}

	content := builder.String()
	content = cleanText(content)

	return content, nil
}

// CanRead returns true if the file is a PDF.
func (r *PdfReader) CanRead(filename string) bool {
	if filename == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".pdf"
}

// cleanText normalizes whitespace and removes excessive blank lines.
func cleanText(text string) string {
	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Remove excessive whitespace while preserving paragraphs
	multipleNewlines := regexp.MustCompile(`\n{3,}`)
	text = multipleNewlines.ReplaceAllString(text, "\n\n")

	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	return text
}
