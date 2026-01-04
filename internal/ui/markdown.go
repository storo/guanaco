package ui

import (
	"bytes"
	"html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// ContentPart represents a parsed content part (text or code).
type ContentPart struct {
	Type     string // "text" or "code"
	Content  string
	Language string // Only for code blocks
}

// MarkdownRenderer converts Markdown to Pango markup for GTK labels.
type MarkdownRenderer struct {
	md goldmark.Markdown
}

// normalizeMarkdown converts common model output patterns to proper Markdown.
// This helps handle cases where models use Unicode bullets or forget heading syntax.
func normalizeMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false

	for i, line := range lines {
		// Track code blocks to avoid modifying code
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		if inCodeBlock {
			result = append(result, line)
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Convert Unicode bullets to Markdown list syntax
		if strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "▪") || strings.HasPrefix(trimmed, "▸") {
			// Preserve original indentation
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			// Remove the Unicode bullet and get the rest
			rest := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "•"), "▪"), "▸")
			rest = strings.TrimLeft(rest, " ")
			line = indent + "- " + rest
			result = append(result, line)
			continue
		}

		// Detect potential headers: short line without punctuation, after blank, with content following
		// Skip if line starts with a number followed by . (ordered list item)
		isOrderedList := len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(trimmed[:3], ".")
		if len(trimmed) > 0 && len(trimmed) < 60 &&
			!isOrderedList &&
			!strings.HasPrefix(trimmed, "#") &&
			!strings.HasPrefix(trimmed, "-") &&
			!strings.HasPrefix(trimmed, "*") &&
			!strings.HasPrefix(trimmed, ">") &&
			!strings.HasPrefix(trimmed, "|") &&
			!strings.HasPrefix(trimmed, "[") &&
			!strings.Contains(trimmed, "```") &&
			!strings.HasSuffix(trimmed, ".") &&
			!strings.HasSuffix(trimmed, ",") &&
			!strings.HasSuffix(trimmed, ";") &&
			!strings.HasSuffix(trimmed, "?") &&
			!strings.HasSuffix(trimmed, "!") &&
			!strings.HasSuffix(trimmed, ":") {

			// Check context: after blank line (or start) and has content after
			isAfterBlank := i == 0 || strings.TrimSpace(lines[i-1]) == ""
			hasContentAfter := false
			if i < len(lines)-1 {
				nextLine := strings.TrimSpace(lines[i+1])
				// Content after should not be empty and not be another potential header
				hasContentAfter = nextLine != "" && !strings.HasPrefix(nextLine, "#")
			}

			if isAfterBlank && hasContentAfter {
				line = "## " + trimmed
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// NewMarkdownRenderer creates a new markdown renderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		md: goldmark.New(
			goldmark.WithExtensions(
				extension.Strikethrough,
				extension.Table,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
	}
}

// ToPango converts markdown text to Pango markup.
func (r *MarkdownRenderer) ToPango(markdown string) string {
	// First decode any HTML entities in the input
	markdown = html.UnescapeString(markdown)
	// Normalize common model output patterns
	markdown = normalizeMarkdown(markdown)

	source := []byte(markdown)
	reader := text.NewReader(source)
	doc := r.md.Parser().Parse(reader)

	var buf bytes.Buffer
	r.renderNode(&buf, doc, source, 0)

	result := buf.String()
	// Clean up extra newlines
	result = strings.TrimSpace(result)
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")

	return result
}

func (r *MarkdownRenderer) renderNode(buf *bytes.Buffer, node ast.Node, source []byte, depth int) {
	switch n := node.(type) {
	case *ast.Document:
		r.renderChildren(buf, n, source, depth)

	case *ast.Heading:
		var size string
		switch n.Level {
		case 1:
			size = "x-large"
		case 2:
			size = "large"
		default:
			size = "medium"
		}
		buf.WriteString("<span size=\"")
		buf.WriteString(size)
		buf.WriteString("\" weight=\"bold\">")
		r.renderChildren(buf, n, source, depth)
		buf.WriteString("</span>")
		buf.WriteString("\n\n")

	case *ast.Paragraph:
		r.renderChildren(buf, n, source, depth)
		if n.NextSibling() != nil {
			buf.WriteString("\n\n")
		}

	case *ast.Text:
		content := string(n.Segment.Value(source))
		buf.WriteString(html.EscapeString(content))
		if n.HardLineBreak() || n.SoftLineBreak() {
			buf.WriteString("\n")
		}

	case *ast.TextBlock:
		r.renderChildren(buf, n, source, depth)

	case *ast.Emphasis:
		if n.Level == 2 {
			buf.WriteString("<b>")
			r.renderChildren(buf, n, source, depth)
			buf.WriteString("</b>")
		} else {
			buf.WriteString("<i>")
			r.renderChildren(buf, n, source, depth)
			buf.WriteString("</i>")
		}

	case *ast.CodeSpan:
		buf.WriteString("<tt>")
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			if t, ok := child.(*ast.Text); ok {
				segment := t.Segment
				buf.WriteString(html.EscapeString(string(segment.Value(source))))
			}
		}
		buf.WriteString("</tt>")

	case *ast.FencedCodeBlock:
		buf.WriteString("<tt>")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			content := string(line.Value(source))
			if i == lines.Len()-1 {
				content = strings.TrimRight(content, "\n")
			}
			buf.WriteString(html.EscapeString(content))
		}
		buf.WriteString("</tt>")

	case *ast.CodeBlock:
		buf.WriteString("<tt>")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			content := string(line.Value(source))
			if i == lines.Len()-1 {
				content = strings.TrimRight(content, "\n")
			}
			buf.WriteString(html.EscapeString(content))
		}
		buf.WriteString("</tt>")

	case *ast.Link:
		buf.WriteString("<a href=\"")
		buf.WriteString(html.EscapeString(string(n.Destination)))
		buf.WriteString("\">")
		r.renderChildren(buf, n, source, depth)
		buf.WriteString("</a>")

	case *ast.AutoLink:
		url := string(n.URL(source))
		buf.WriteString("<a href=\"")
		buf.WriteString(html.EscapeString(url))
		buf.WriteString("\">")
		buf.WriteString(html.EscapeString(url))
		buf.WriteString("</a>")

	case *ast.List:
		for i, child := 0, n.FirstChild(); child != nil; i, child = i+1, child.NextSibling() {
			if listItem, ok := child.(*ast.ListItem); ok {
				if n.IsOrdered() {
					buf.WriteString("  ")
					buf.WriteString(string(rune('1' + i)))
					buf.WriteString(". ")
				} else {
					buf.WriteString("  • ")
				}
				r.renderListItemContent(buf, listItem, source, depth)
				if child.NextSibling() != nil {
					buf.WriteString("\n")
				}
			}
		}
		if n.NextSibling() != nil {
			buf.WriteString("\n\n")
		}

	case *ast.ListItem:
		// Handled by List

	case *ast.Blockquote:
		buf.WriteString("<i>▎ ")
		r.renderBlockquoteContent(buf, n, source, depth)
		buf.WriteString("</i>")
		if n.NextSibling() != nil {
			buf.WriteString("\n\n")
		}

	case *ast.ThematicBreak:
		buf.WriteString("\n────────\n")

	case *east.Table:
		// Render table rows with pipe separators
		for row := n.FirstChild(); row != nil; row = row.NextSibling() {
			if tableRow, ok := row.(*east.TableRow); ok {
				buf.WriteString("  ")
				isFirst := true
				for cell := tableRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
					if !isFirst {
						buf.WriteString(" │ ")
					}
					r.renderChildren(buf, cell, source, depth)
					isFirst = false
				}
				buf.WriteString("\n")
			} else if tableHeader, ok := row.(*east.TableHeader); ok {
				buf.WriteString("  ")
				isFirst := true
				for cell := tableHeader.FirstChild(); cell != nil; cell = cell.NextSibling() {
					if !isFirst {
						buf.WriteString(" │ ")
					}
					buf.WriteString("<b>")
					r.renderChildren(buf, cell, source, depth)
					buf.WriteString("</b>")
					isFirst = false
				}
				buf.WriteString("\n")
			}
		}
		if n.NextSibling() != nil {
			buf.WriteString("\n")
		}

	case *east.TableHeader, *east.TableRow, *east.TableCell:
		// Handled by Table

	case *ast.String:
		buf.WriteString(html.EscapeString(string(n.Value)))

	case *ast.RawHTML:
		// Skip raw HTML for security

	default:
		// Check for strikethrough extension
		if n.Kind().String() == "Strikethrough" {
			buf.WriteString("<s>")
			r.renderChildren(buf, n, source, depth)
			buf.WriteString("</s>")
		} else {
			r.renderChildren(buf, n, source, depth)
		}
	}
}

func (r *MarkdownRenderer) renderChildren(buf *bytes.Buffer, node ast.Node, source []byte, depth int) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderNode(buf, child, source, depth+1)
	}
}

func (r *MarkdownRenderer) renderListItemContent(buf *bytes.Buffer, item *ast.ListItem, source []byte, depth int) {
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			r.renderChildren(buf, para, source, depth)
		} else {
			r.renderNode(buf, child, source, depth)
		}
	}
}

func (r *MarkdownRenderer) renderBlockquoteContent(buf *bytes.Buffer, quote *ast.Blockquote, source []byte, depth int) {
	for child := quote.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			r.renderChildren(buf, para, source, depth)
		} else {
			r.renderNode(buf, child, source, depth)
		}
	}
}

// Parse splits markdown into content parts (text and code blocks).
func (r *MarkdownRenderer) Parse(markdown string) []ContentPart {
	// First decode any HTML entities in the input
	markdown = html.UnescapeString(markdown)
	// Normalize common model output patterns
	markdown = normalizeMarkdown(markdown)

	source := []byte(markdown)
	reader := text.NewReader(source)
	doc := r.md.Parser().Parse(reader)

	var parts []ContentPart
	var textBuf bytes.Buffer

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.FencedCodeBlock:
			// Flush any accumulated text
			if textBuf.Len() > 0 {
				text := strings.TrimSpace(textBuf.String())
				if text != "" {
					parts = append(parts, ContentPart{
						Type:    "text",
						Content: text,
					})
				}
				textBuf.Reset()
			}

			// Extract code block
			var codeBuf bytes.Buffer
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				content := string(line.Value(source))
				if i == lines.Len()-1 {
					content = strings.TrimRight(content, "\n")
				}
				codeBuf.WriteString(content)
			}

			lang := ""
			if n.Info != nil {
				lang = string(n.Info.Segment.Value(source))
				// Remove any extra info after language
				if idx := strings.IndexByte(lang, ' '); idx > 0 {
					lang = lang[:idx]
				}
			}

			parts = append(parts, ContentPart{
				Type:     "code",
				Content:  codeBuf.String(),
				Language: lang,
			})

		case *ast.CodeBlock:
			// Flush any accumulated text
			if textBuf.Len() > 0 {
				text := strings.TrimSpace(textBuf.String())
				if text != "" {
					parts = append(parts, ContentPart{
						Type:    "text",
						Content: text,
					})
				}
				textBuf.Reset()
			}

			// Extract code block
			var codeBuf bytes.Buffer
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				content := string(line.Value(source))
				if i == lines.Len()-1 {
					content = strings.TrimRight(content, "\n")
				}
				codeBuf.WriteString(content)
			}

			parts = append(parts, ContentPart{
				Type:    "code",
				Content: codeBuf.String(),
			})

		default:
			// Render other nodes to text buffer
			r.renderNode(&textBuf, child, source, 0)
		}
	}

	// Flush remaining text
	if textBuf.Len() > 0 {
		text := strings.TrimSpace(textBuf.String())
		text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
		if text != "" {
			parts = append(parts, ContentPart{
				Type:    "text",
				Content: text,
			})
		}
	}

	return parts
}
