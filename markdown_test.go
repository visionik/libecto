package libecto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "paragraph",
			input:    "Hello, world!",
			contains: []string{"<p>", "Hello, world!", "</p>"},
		},
		{
			name:     "heading h1",
			input:    "# Heading 1",
			contains: []string{"<h1", "Heading 1", "</h1>"},
		},
		{
			name:     "heading h2",
			input:    "## Heading 2",
			contains: []string{"<h2", "Heading 2", "</h2>"},
		},
		{
			name:     "heading h3",
			input:    "### Heading 3",
			contains: []string{"<h3", "Heading 3", "</h3>"},
		},
		{
			name:     "bold text",
			input:    "This is **bold** text",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "italic text",
			input:    "This is *italic* text",
			contains: []string{"<em>italic</em>"},
		},
		{
			name:     "link",
			input:    "[Example](https://example.com)",
			contains: []string{`<a href="https://example.com"`, "Example", "</a>"},
		},
		{
			name:     "unordered list",
			input:    "- Item 1\n- Item 2\n- Item 3",
			contains: []string{"<ul>", "<li>Item 1</li>", "<li>Item 2</li>", "<li>Item 3</li>", "</ul>"},
		},
		{
			name:     "ordered list",
			input:    "1. First\n2. Second\n3. Third",
			contains: []string{"<ol>", "<li>First</li>", "<li>Second</li>", "</ol>"},
		},
		{
			name:     "code inline",
			input:    "Use `code` here",
			contains: []string{"<code>code</code>"},
		},
		{
			name:     "code block",
			input:    "```\ncode block\n```",
			contains: []string{"<pre>", "<code>", "code block", "</code>", "</pre>"},
		},
		{
			name:     "code block with language",
			input:    "```go\nfunc main() {}\n```",
			contains: []string{"<pre>", "<code", "func main()", "</code>", "</pre>"},
		},
		{
			name:     "blockquote",
			input:    "> This is a quote",
			contains: []string{"<blockquote>", "This is a quote", "</blockquote>"},
		},
		{
			name:     "horizontal rule",
			input:    "Above\n\n---\n\nBelow",
			contains: []string{"<hr"},
		},
		{
			name:     "image",
			input:    "![Alt text](https://example.com/image.jpg)",
			contains: []string{`<img`, `src="https://example.com/image.jpg"`, `alt="Alt text"`},
		},
		{
			name:     "table",
			input:    "| Header |\n|--------|\n| Cell   |",
			contains: []string{"<table>", "<thead>", "<th>Header</th>", "<td>Cell</td>", "</table>"},
		},
		{
			name:     "strikethrough",
			input:    "This is ~~deleted~~ text",
			contains: []string{"<del>deleted</del>"},
		},
		{
			name:     "autolink",
			input:    "Visit https://example.com for more",
			contains: []string{`<a href="https://example.com"`, "</a>"},
		},
		{
			name:     "nested formatting",
			input:    "This is ***bold and italic*** text",
			contains: []string{"<strong>", "<em>", "bold and italic", "</em>", "</strong>"},
		},
		{
			name:     "empty input",
			input:    "",
			contains: []string{},
		},
		{
			name:     "whitespace only",
			input:    "   \n   \n   ",
			contains: []string{},
		},
		{
			name:     "multiple paragraphs",
			input:    "First paragraph.\n\nSecond paragraph.",
			contains: []string{"<p>First paragraph.</p>", "<p>Second paragraph.</p>"},
		},
		{
			name:  "heading with auto ID",
			input: "# My Heading",
			// AutoHeadingIDs should generate an id attribute
			contains: []string{`id="my-heading"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownToHTML([]byte(tt.input))
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "expected %q in output", expected)
			}
		})
	}
}

func TestMarkdownStringToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "simple paragraph",
			input:    "Hello world",
			contains: []string{"<p>Hello world</p>"},
		},
		{
			name:     "heading",
			input:    "# Title",
			contains: []string{"<h1", "Title", "</h1>"},
		},
		{
			name:     "empty string",
			input:    "",
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownStringToHTML(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestMarkdownToHTML_Consistency(t *testing.T) {
	// Ensure byte slice and string versions produce same output
	input := "# Test\n\nThis is a **test** with [a link](http://example.com)."

	byteResult := MarkdownToHTML([]byte(input))
	stringResult := MarkdownStringToHTML(input)

	assert.Equal(t, byteResult, stringResult)
}

func TestMarkdownToHTML_LargeInput(t *testing.T) {
	// Test with large input to ensure no panics or excessive memory usage
	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		builder.WriteString("# Heading\n\n")
		builder.WriteString("This is paragraph number ")
		builder.WriteString(string(rune('0' + (i % 10))))
		builder.WriteString(" with **bold** and *italic* text.\n\n")
		builder.WriteString("- List item 1\n")
		builder.WriteString("- List item 2\n\n")
	}

	input := builder.String()
	result := MarkdownToHTML([]byte(input))

	// Just verify it doesn't panic and produces output
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "<h1")
	assert.Contains(t, result, "<strong>bold</strong>")
}

func TestMarkdownToHTML_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		notContains []string
	}{
		{
			name:     "HTML entities",
			input:    "Less than < and greater than >",
			contains: []string{"&lt;", "&gt;"},
		},
		{
			name:     "ampersand",
			input:    "Tom & Jerry",
			contains: []string{"&amp;"},
		},
		{
			name:     "unicode",
			input:    "Hello 世界 🌍",
			contains: []string{"世界", "🌍"},
		},
		{
			name:     "escaped markdown",
			input:    `This is \*not italic\*`,
			notContains: []string{"<em>not italic</em>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownToHTML([]byte(tt.input))
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
			for _, notExpected := range tt.notContains {
				assert.NotContains(t, result, notExpected)
			}
		})
	}
}

func TestMarkdownToHTML_NestedStructures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "nested lists",
			input:    "- Parent\n  - Child\n    - Grandchild",
			contains: []string{"<ul>", "<li>Parent", "<li>Child", "<li>Grandchild"},
		},
		{
			name:     "list with paragraphs",
			input:    "- Item 1\n\n  Continued paragraph\n\n- Item 2",
			contains: []string{"<li>", "Item 1", "Item 2"},
		},
		{
			name:     "blockquote with formatting",
			input:    "> This is **bold** in a quote",
			contains: []string{"<blockquote>", "<strong>bold</strong>", "</blockquote>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarkdownToHTML([]byte(tt.input))
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// Fuzz tests

func FuzzMarkdownToHTML(f *testing.F) {
	// Seed with various markdown patterns
	seeds := []string{
		"# Heading",
		"**bold**",
		"*italic*",
		"[link](http://example.com)",
		"![image](http://example.com/img.png)",
		"```\ncode\n```",
		"- list\n- items",
		"> quote",
		"| table | header |\n|-------|--------|\n| cell  | cell   |",
		"",
		"   ",
		"\n\n\n",
		"# Multiple\n## Headings\n### Here",
		"***bold italic***",
		"~~strikethrough~~",
		"Normal text",
		"Text with `inline code` here",
		"---",
		"Multiple\n\nparagraphs\n\nhere",
		"<script>alert('xss')</script>",
		"&amp; entities &lt; here &gt;",
		"Unicode: 日本語 🎉 émojis",
		strings.Repeat("*", 100),
		strings.Repeat("#", 100),
		strings.Repeat("`", 100),
		"[[[nested]]]",
		"((((parentheses))))",
		"\\*escaped\\*",
		"http://autolink.example.com",
		"user@email.com",
	}

	// Add more fuzz seeds for coverage
	for i := 0; i < 30; i++ {
		seeds = append(seeds, strings.Repeat("# ", i)+"Heading")
		seeds = append(seeds, strings.Repeat("- ", i)+"List item")
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic
		result := MarkdownToHTML([]byte(input))
		// Result should be a string (not nil or invalid)
		_ = len(result)
	})
}

// Benchmarks

func BenchmarkMarkdownToHTML_Simple(b *testing.B) {
	input := []byte("# Hello World\n\nThis is a simple paragraph.")
	for i := 0; i < b.N; i++ {
		_ = MarkdownToHTML(input)
	}
}

func BenchmarkMarkdownToHTML_Complex(b *testing.B) {
	input := []byte(`# Main Title

This is a paragraph with **bold**, *italic*, and ***both***.

## Lists

- Item 1
- Item 2
  - Nested item
- Item 3

## Code

` + "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```" + `

## Table

| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |

> A blockquote with [a link](https://example.com)
`)
	for i := 0; i < b.N; i++ {
		_ = MarkdownToHTML(input)
	}
}

func BenchmarkMarkdownStringToHTML(b *testing.B) {
	input := "# Test\n\nParagraph with **formatting**."
	for i := 0; i < b.N; i++ {
		_ = MarkdownStringToHTML(input)
	}
}
