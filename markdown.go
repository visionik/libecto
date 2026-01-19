package libecto

import (
	"github.com/russross/blackfriday/v2"
)

// MarkdownToHTML converts markdown content to HTML.
// It uses blackfriday with common extensions including:
//   - CommonExtensions (tables, fenced code, autolinks, strikethrough, etc.)
//   - AutoHeadingIDs (generates IDs for headings)
//   - NoEmptyLineBeforeBlock (allows blocks without preceding blank lines)
//
// The function is safe to use with untrusted input as it does not execute
// any embedded code, though the output should still be sanitized if
// displayed in a web context.
func MarkdownToHTML(md []byte) string {
	html := blackfriday.Run(md,
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions|
				blackfriday.AutoHeadingIDs|
				blackfriday.NoEmptyLineBeforeBlock,
		),
	)
	return string(html)
}

// MarkdownStringToHTML is a convenience function that converts a markdown
// string to HTML. It is equivalent to calling MarkdownToHTML with the
// string converted to bytes.
func MarkdownStringToHTML(md string) string {
	return MarkdownToHTML([]byte(md))
}
