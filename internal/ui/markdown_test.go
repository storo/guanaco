package ui

import (
	"testing"
)

func TestMarkdownToPango(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "plain text",
			markdown: "Hello world",
			expected: "Hello world",
		},
		{
			name:     "bold text",
			markdown: "Hello **world**",
			expected: "Hello <b>world</b>",
		},
		{
			name:     "italic text",
			markdown: "Hello *world*",
			expected: "Hello <i>world</i>",
		},
		{
			name:     "bold and italic",
			markdown: "Hello ***world***",
			expected: "Hello <i><b>world</b></i>",
		},
		{
			name:     "inline code",
			markdown: "Use `fmt.Println`",
			expected: "Use <tt>fmt.Println</tt>",
		},
		{
			name:     "code block",
			markdown: "```go\nfmt.Println(\"hello\")\n```",
			expected: "<tt>fmt.Println(&#34;hello&#34;)</tt>",
		},
		{
			name:     "code block without language",
			markdown: "```\nsome code\n```",
			expected: "<tt>some code</tt>",
		},
		{
			name:     "heading 1",
			markdown: "# Title",
			expected: "<span size=\"x-large\" weight=\"bold\">Title</span>",
		},
		{
			name:     "heading 2",
			markdown: "## Subtitle",
			expected: "<span size=\"large\" weight=\"bold\">Subtitle</span>",
		},
		{
			name:     "heading 3",
			markdown: "### Section",
			expected: "<span size=\"medium\" weight=\"bold\">Section</span>",
		},
		{
			name:     "link",
			markdown: "[Google](https://google.com)",
			expected: "<a href=\"https://google.com\">Google</a>",
		},
		{
			name:     "strikethrough",
			markdown: "Hello ~~world~~",
			expected: "Hello <s>world</s>",
		},
		{
			name:     "unordered list",
			markdown: "- item 1\n- item 2",
			expected: "• item 1\n  • item 2",
		},
		{
			name:     "ordered list",
			markdown: "1. first\n2. second",
			expected: "1. first\n  2. second",
		},
		{
			name:     "blockquote",
			markdown: "> This is a quote",
			expected: "<i>▎ This is a quote</i>",
		},
		{
			name:     "horizontal rule",
			markdown: "---",
			expected: "────────",
		},
		{
			name:     "complex markdown",
			markdown: "# Hello\n\nThis is **bold** and *italic* with `code`.\n\n- Item 1\n- Item 2",
			expected: "<span size=\"x-large\" weight=\"bold\">Hello</span>\n\nThis is <b>bold</b> and <i>italic</i> with <tt>code</tt>.\n\n  • Item 1\n  • Item 2",
		},
		{
			name:     "escape ampersand",
			markdown: "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
	}

	renderer := NewMarkdownRenderer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.ToPango(tt.markdown)
			if result != tt.expected {
				t.Errorf("ToPango(%q)\ngot:  %q\nwant: %q", tt.markdown, result, tt.expected)
			}
		})
	}
}

func BenchmarkMarkdownToPango(b *testing.B) {
	renderer := NewMarkdownRenderer()
	markdown := `# Hello World

This is a **bold** statement with *italic* text and some ` + "`code`" + `.

## Features

- Item 1
- Item 2
- Item 3

> This is a blockquote

` + "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = renderer.ToPango(markdown)
	}
}
