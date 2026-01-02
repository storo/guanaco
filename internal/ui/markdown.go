package ui

import (
	"bytes"
	"html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// MarkdownRenderer converts Markdown to Pango markup for GTK labels.
type MarkdownRenderer struct {
	md goldmark.Markdown
}

// NewMarkdownRenderer creates a new markdown renderer.
func NewMarkdownRenderer() *MarkdownRenderer {
	return &MarkdownRenderer{
		md: goldmark.New(
			goldmark.WithExtensions(extension.Strikethrough),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
		),
	}
}

// ToPango converts markdown text to Pango markup.
func (r *MarkdownRenderer) ToPango(markdown string) string {
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
					buf.WriteString(string(rune('1' + i)))
					buf.WriteString(". ")
				} else {
					buf.WriteString("• ")
				}
				r.renderListItemContent(buf, listItem, source, depth)
				if child.NextSibling() != nil {
					buf.WriteString("\n")
				}
			}
		}

	case *ast.ListItem:
		// Handled by List

	case *ast.Blockquote:
		buf.WriteString("<i>▎ ")
		r.renderBlockquoteContent(buf, n, source, depth)
		buf.WriteString("</i>")

	case *ast.ThematicBreak:
		buf.WriteString("───────────────────────────────────────")

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
