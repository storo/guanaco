package ui

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightToken represents a single token with styling information.
type HighlightToken struct {
	Text   string
	Color  string // Hex color like "#FF0000"
	Bold   bool
	Italic bool
}

// SyntaxHighlighter provides syntax highlighting using Chroma.
type SyntaxHighlighter struct {
	style *chroma.Style
}

// NewSyntaxHighlighter creates a new syntax highlighter.
func NewSyntaxHighlighter() *SyntaxHighlighter {
	// Use a dark theme that works well with Adwaita dark
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	return &SyntaxHighlighter{
		style: style,
	}
}

// Highlight tokenizes the code and returns styled tokens.
func (sh *SyntaxHighlighter) Highlight(code, language string) []HighlightToken {
	// Get lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		// Try to detect from content
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Coalesce runs of identical token types
	lexer = chroma.Coalesce(lexer)

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Return plain text on error
		return []HighlightToken{{Text: code}}
	}

	var tokens []HighlightToken
	for _, token := range iterator.Tokens() {
		entry := sh.style.Get(token.Type)

		// Get color, fallback to white if not set
		color := ""
		if entry.Colour.IsSet() {
			color = entry.Colour.String()
		}

		tokens = append(tokens, HighlightToken{
			Text:   token.Value,
			Color:  color,
			Bold:   entry.Bold == chroma.Yes,
			Italic: entry.Italic == chroma.Yes,
		})
	}

	return tokens
}

// GetBackgroundColor returns the style's background color.
func (sh *SyntaxHighlighter) GetBackgroundColor() string {
	bg := sh.style.Get(chroma.Background)
	if bg.Background.IsSet() {
		return bg.Background.String()
	}
	return "#282a36" // Dracula default background
}
