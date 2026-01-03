package rag

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

// ImageReader reads image files and converts them to base64.
type ImageReader struct{}

// NewImageReader creates a new image reader.
func NewImageReader() *ImageReader {
	return &ImageReader{}
}

// imageExtensions contains supported image file extensions.
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

// CanRead returns true if the file is a supported image format.
func (r *ImageReader) CanRead(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExtensions[ext]
}

// Read reads an image file and returns its base64-encoded content.
func (r *ImageReader) Read(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// IsImage checks if a filename is a supported image format.
func IsImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExtensions[ext]
}
