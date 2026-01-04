package ui

import (
	"testing"
)

func TestTruncatePreview(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncate long string",
			input:  "hello world",
			maxLen: 5,
			want:   "hello…",
		},
		{
			name:   "newlines replaced",
			input:  "hello\nworld",
			maxLen: 20,
			want:   "hello world",
		},
		{
			name:   "newlines and truncate",
			input:  "hello\nworld\ntest",
			maxLen: 11,
			want:   "hello world…",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "whitespace trimmed",
			input:  "  hello  ",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "multiple newlines",
			input:  "line1\n\n\nline2",
			maxLen: 20,
			want:   "line1   line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePreview(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncatePreview(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
