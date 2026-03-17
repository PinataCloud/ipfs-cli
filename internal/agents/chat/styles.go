// Package chat provides a TUI for chatting with Pinata AI agents.
package chat

import (
	"os"

	"charm.land/lipgloss/v2"
)

// NoColor returns true if colors should be disabled (NO_COLOR env var is set).
func NoColor() bool {
	_, exists := os.LookupEnv("NO_COLOR")
	return exists
}

// Styles holds all the lipgloss styles for the chat TUI.
type Styles struct {
	// Message role styles
	UserStyle      lipgloss.Style
	AssistantStyle lipgloss.Style
	ToolStyle      lipgloss.Style
	SystemStyle    lipgloss.Style

	// Component styles
	BorderStyle  lipgloss.Style
	SpinnerStyle lipgloss.Style
	InputStyle   lipgloss.Style
	HelpStyle    lipgloss.Style

	// Status styles
	SuccessStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	PendingStyle lipgloss.Style
	RunningStyle lipgloss.Style

	// Label styles
	UserLabelStyle      lipgloss.Style
	AssistantLabelStyle lipgloss.Style
	ToolLabelStyle      lipgloss.Style
}

// DefaultStyles returns the default color scheme for the chat TUI.
func DefaultStyles() Styles {
	noColor := NoColor()

	if noColor {
		return Styles{
			UserStyle:           lipgloss.NewStyle(),
			AssistantStyle:      lipgloss.NewStyle(),
			ToolStyle:           lipgloss.NewStyle(),
			SystemStyle:         lipgloss.NewStyle().Italic(true),
			BorderStyle:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
			SpinnerStyle:        lipgloss.NewStyle(),
			InputStyle:          lipgloss.NewStyle(),
			HelpStyle:           lipgloss.NewStyle().Faint(true),
			SuccessStyle:        lipgloss.NewStyle(),
			ErrorStyle:          lipgloss.NewStyle(),
			PendingStyle:        lipgloss.NewStyle(),
			RunningStyle:        lipgloss.NewStyle(),
			UserLabelStyle:      lipgloss.NewStyle().Bold(true),
			AssistantLabelStyle: lipgloss.NewStyle().Bold(true),
			ToolLabelStyle:      lipgloss.NewStyle().Bold(true),
		}
	}

	// Color definitions
	userColor := lipgloss.Color("12")      // Blue
	assistantColor := lipgloss.Color("10") // Green
	toolColor := lipgloss.Color("11")      // Yellow
	systemColor := lipgloss.Color("8")     // Gray
	errorColor := lipgloss.Color("9")      // Red
	successColor := lipgloss.Color("10")   // Green
	pendingColor := lipgloss.Color("3")    // Olive/Yellow
	runningColor := lipgloss.Color("14")   // Cyan

	return Styles{
		UserStyle:      lipgloss.NewStyle().Foreground(userColor),
		AssistantStyle: lipgloss.NewStyle().Foreground(assistantColor),
		ToolStyle:      lipgloss.NewStyle().Foreground(toolColor),
		SystemStyle:    lipgloss.NewStyle().Foreground(systemColor).Italic(true),

		BorderStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(toolColor).
			Padding(0, 1),

		SpinnerStyle: lipgloss.NewStyle().Foreground(assistantColor),
		InputStyle:   lipgloss.NewStyle().BorderForeground(userColor),
		HelpStyle:    lipgloss.NewStyle().Foreground(systemColor).Faint(true),

		SuccessStyle: lipgloss.NewStyle().Foreground(successColor),
		ErrorStyle:   lipgloss.NewStyle().Foreground(errorColor),
		PendingStyle: lipgloss.NewStyle().Foreground(pendingColor),
		RunningStyle: lipgloss.NewStyle().Foreground(runningColor),

		UserLabelStyle:      lipgloss.NewStyle().Foreground(userColor).Bold(true),
		AssistantLabelStyle: lipgloss.NewStyle().Foreground(assistantColor).Bold(true),
		ToolLabelStyle:      lipgloss.NewStyle().Foreground(toolColor).Bold(true),
	}
}

// ToolBoxStyle returns a styled border for tool call boxes.
func (s Styles) ToolBoxStyle(width int) lipgloss.Style {
	return s.BorderStyle.Width(width - 4) // Account for border padding
}

// InputPromptStyle returns the style for the input prompt.
func (s Styles) InputPromptStyle() lipgloss.Style {
	return s.UserLabelStyle
}
