package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// ToolStatus represents the execution status of a tool call.
type ToolStatus string

const (
	ToolStatusPending ToolStatus = "pending"
	ToolStatusRunning ToolStatus = "running"
	ToolStatusDone    ToolStatus = "done"
	ToolStatusError   ToolStatus = "error"
	ToolStatusDenied  ToolStatus = "denied"
)

// ToolCall represents a tool invocation request from the assistant.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
	Status    ToolStatus             `json:"status"`
}

// ToolResult represents the result of a tool execution.
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

// RenderToolCall renders a tool call as a bordered box with tool name and parameters.
func RenderToolCall(tool ToolCall, width int, styles Styles) string {
	// Determine status style and indicator
	var statusIndicator string
	var statusStyle lipgloss.Style

	switch tool.Status {
	case ToolStatusPending:
		statusIndicator = "[pending]"
		statusStyle = styles.PendingStyle
	case ToolStatusRunning:
		statusIndicator = "[running]"
		statusStyle = styles.RunningStyle
	case ToolStatusDone:
		statusIndicator = "[done]"
		statusStyle = styles.SuccessStyle
	case ToolStatusError:
		statusIndicator = "[error]"
		statusStyle = styles.ErrorStyle
	case ToolStatusDenied:
		statusIndicator = "[denied]"
		statusStyle = styles.ErrorStyle
	default:
		statusIndicator = ""
		statusStyle = styles.ToolStyle
	}

	// Build header
	header := styles.ToolLabelStyle.Render("Tool: " + tool.Name)
	if statusIndicator != "" {
		header += " " + statusStyle.Render(statusIndicator)
	}

	// Format arguments
	var argsLines []string
	if len(tool.Arguments) > 0 {
		argsLines = formatArguments(tool.Arguments, width-6, "") // Account for border padding
	}

	// Build content
	var content strings.Builder
	content.WriteString(header)
	if len(argsLines) > 0 {
		content.WriteString("\n")
		for _, line := range argsLines {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Apply border style
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	borderStyle := styles.BorderStyle.Width(boxWidth)
	return borderStyle.Render(content.String())
}

// RenderToolResult renders a tool result as a bordered box.
func RenderToolResult(result ToolResult, width int, styles Styles) string {
	// Determine header style based on error status
	var header string
	if result.IsError {
		header = styles.ErrorStyle.Render("Tool Error")
	} else {
		header = styles.SuccessStyle.Render("Tool Result")
	}

	// Truncate long content
	content := result.Content
	maxContentLen := 500
	if len(content) > maxContentLen {
		content = content[:maxContentLen] + "\n... (truncated)"
	}

	// Wrap content
	wrappedContent := wrapText(content, width-6)

	// Build box content
	var boxContent strings.Builder
	boxContent.WriteString(header)
	boxContent.WriteString("\n")
	boxContent.WriteString(wrappedContent)

	// Apply border style
	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}

	borderStyle := styles.BorderStyle.Width(boxWidth)
	return borderStyle.Render(boxContent.String())
}

// RenderPermissionPrompt renders the tool permission prompt.
func RenderPermissionPrompt(tool ToolCall, styles Styles) string {
	return styles.HelpStyle.Render("[a]llow  [d]eny  [A]lways allow  [N]ever allow")
}

// formatArguments formats tool arguments for display.
func formatArguments(args map[string]interface{}, width int, indent string) []string {
	var lines []string

	for key, value := range args {
		formattedValue := formatValue(value, width-len(key)-4, indent+"  ")
		line := fmt.Sprintf("%s%s: %s", indent, key, formattedValue)
		lines = append(lines, line)
	}

	return lines
}

// formatValue formats a single argument value for display.
func formatValue(value interface{}, width int, indent string) string {
	switch v := value.(type) {
	case string:
		// Quote and potentially truncate long strings
		if len(v) > 100 {
			v = v[:97] + "..."
		}
		return fmt.Sprintf("%q", v)

	case float64:
		// Format numbers nicely
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)

	case bool:
		return fmt.Sprintf("%t", v)

	case nil:
		return "null"

	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		// Format as JSON array for readability
		jsonBytes, _ := json.MarshalIndent(v, indent, "  ")
		return string(jsonBytes)

	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		// Format as JSON object for readability
		jsonBytes, _ := json.MarshalIndent(v, indent, "  ")
		return string(jsonBytes)

	default:
		// Fall back to JSON encoding
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}

// ToolPermissions manages tool permission settings for the session.
type ToolPermissions struct {
	alwaysAllow map[string]bool // Tools that are always allowed
	neverAllow  map[string]bool // Tools that are never allowed
}

// NewToolPermissions creates a new tool permissions manager.
func NewToolPermissions() *ToolPermissions {
	return &ToolPermissions{
		alwaysAllow: make(map[string]bool),
		neverAllow:  make(map[string]bool),
	}
}

// Check checks if a tool has a stored permission.
// Returns (allowed, hasPermission).
func (p *ToolPermissions) Check(toolName string) (bool, bool) {
	if p.alwaysAllow[toolName] {
		return true, true
	}
	if p.neverAllow[toolName] {
		return false, true
	}
	return false, false
}

// SetAlwaysAllow marks a tool as always allowed.
func (p *ToolPermissions) SetAlwaysAllow(toolName string) {
	p.alwaysAllow[toolName] = true
	delete(p.neverAllow, toolName)
}

// SetNeverAllow marks a tool as never allowed.
func (p *ToolPermissions) SetNeverAllow(toolName string) {
	p.neverAllow[toolName] = true
	delete(p.alwaysAllow, toolName)
}

// Clear removes all stored permissions.
func (p *ToolPermissions) Clear() {
	p.alwaysAllow = make(map[string]bool)
	p.neverAllow = make(map[string]bool)
}
