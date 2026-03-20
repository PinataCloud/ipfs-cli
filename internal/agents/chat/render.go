package chat

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

// Cached renderers to avoid recreating on every render and terminal queries
var (
	rendererMu       sync.Mutex
	cachedRenderer   *glamour.TermRenderer
	cachedNoColor    *glamour.TermRenderer
	cachedWidth      int
	cachedNoColorVal bool
)

// getRenderer returns a cached glamour renderer, creating one if needed.
// Uses "dark" style by default to avoid terminal color queries.
func getRenderer(width int, noColor bool) *glamour.TermRenderer {
	rendererMu.Lock()
	defer rendererMu.Unlock()

	// Return cached renderer if width matches
	if noColor && cachedNoColor != nil && cachedWidth == width {
		return cachedNoColor
	}
	if !noColor && cachedRenderer != nil && cachedWidth == width {
		return cachedRenderer
	}

	var renderer *glamour.TermRenderer
	var err error

	if noColor {
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("notty"),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			cachedNoColor = renderer
		}
	} else {
		// Use "dark" style instead of WithAutoStyle() to avoid terminal queries
		// WithAutoStyle() triggers OSC 11 queries that cause escape sequence issues
		renderer, err = glamour.NewTermRenderer(
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			cachedRenderer = renderer
		}
	}

	cachedWidth = width
	return renderer
}

// RenderMarkdown renders markdown content with syntax highlighting.
// It respects the noColor flag and wraps content to the specified width.
func RenderMarkdown(content string, width int, noColor bool) string {
	if content == "" {
		return ""
	}

	// Ensure minimum width
	if width < 40 {
		width = 40
	}

	renderer := getRenderer(width, noColor)
	if renderer == nil {
		// Fall back to plain text if renderer creation fails
		return wrapText(content, width)
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		// Fall back to plain text if rendering fails
		return wrapText(content, width)
	}

	// Trim trailing newlines that glamour adds
	return strings.TrimSuffix(rendered, "\n")
}

// RenderPlainText renders content without markdown processing.
// Used for tool outputs and error messages.
func RenderPlainText(content string, width int) string {
	return wrapText(content, width)
}

// wrapText wraps text to the specified width.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Handle empty lines
		if len(line) == 0 {
			continue
		}

		// Wrap long lines
		for len(line) > width {
			// Find a good break point (space)
			breakPoint := width
			for j := width - 1; j > width/2; j-- {
				if line[j] == ' ' {
					breakPoint = j
					break
				}
			}

			result.WriteString(line[:breakPoint])
			result.WriteString("\n")
			line = strings.TrimLeft(line[breakPoint:], " ")
		}
		result.WriteString(line)
	}

	return result.String()
}

// TruncateText truncates text to the specified length, adding ellipsis if needed.
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return text[:maxLen]
	}
	return text[:maxLen-3] + "..."
}

// IndentText adds the specified prefix to each line of text.
func IndentText(text string, prefix string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(prefix)
		result.WriteString(line)
	}

	return result.String()
}
