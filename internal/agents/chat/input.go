package chat

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// InputModel wraps a textarea for multi-line chat input.
// Shift+Enter inserts a newline, Enter sends the message.
type InputModel struct {
	textarea    textarea.Model
	placeholder string
	submitted   bool
	focused     bool
	maxHeight   int // Maximum height the textarea can expand to
}

// NewInputModel creates a new multi-line input component.
func NewInputModel(placeholder string) InputModel {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000
	ta.SetWidth(80)
	ta.SetHeight(1) // Start with single line, expands dynamically
	ta.Prompt = ""  // Remove default prompt

	// Disable the default InsertNewline binding - we handle enter/shift+enter ourselves
	km := textarea.DefaultKeyMap()
	km.InsertNewline = key.NewBinding() // Unbind enter from inserting newlines
	ta.KeyMap = km

	// Use minimal styling - no background color, just text
	styles := textarea.Styles{}
	styles.Focused.Base = lipgloss.NewStyle()
	styles.Focused.Text = lipgloss.NewStyle()
	styles.Focused.Placeholder = lipgloss.NewStyle().Faint(true)
	styles.Blurred.Base = lipgloss.NewStyle()
	styles.Blurred.Text = lipgloss.NewStyle()
	styles.Blurred.Placeholder = lipgloss.NewStyle().Faint(true)
	ta.SetStyles(styles)

	ta.Focus()

	return InputModel{
		textarea:    ta,
		placeholder: placeholder,
		focused:     true,
		maxHeight:   10, // Allow up to 10 lines
	}
}

// Init initializes the input model.
func (m InputModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the input component.
// Only KeyPressMsg and PasteMsg should be passed to this function.
func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Use msg.String() for proper key detection with keyboard enhancements
		keyStr := msg.String()
		k := msg.Key()

		// Filter out terminal escape sequence responses
		if IsEscapeSequence(keyStr) {
			return m, nil
		}

		// Check for enter key (with or without modifiers)
		// Handle multiple scenarios:
		// 1. Kitty keyboard protocol: msg.String() == "shift+enter"
		// 2. Direct modifier check: k.Mod&tea.ModShift != 0
		// 3. Terminal fallback: ctrl+j (some terminals report shift+enter as ctrl+j)
		if k.Code == tea.KeyEnter || keyStr == "enter" || keyStr == "shift+enter" || keyStr == "ctrl+j" {
			// Shift+Enter inserts newline
			// Also treat ctrl+j as newline (common terminal behavior for shift+enter)
			if keyStr == "shift+enter" || keyStr == "ctrl+j" || k.Mod&tea.ModShift != 0 {
				m.textarea.InsertString("\n")
				m.updateHeight()
				return m, nil
			}
			// Plain Enter submits
			m.submitted = true
			return m, nil
		}

		// Ctrl+U clears the input
		if keyStr == "ctrl+u" {
			m.textarea.SetValue("")
			m.textarea.SetHeight(1)
			m.updateHeight()
			return m, nil
		}

		// Let parent handle Ctrl+C and Ctrl+D
		if keyStr == "ctrl+c" || keyStr == "ctrl+d" {
			return m, nil
		}

		// Pass KeyPressMsg to textarea for normal character input
		m.textarea, cmd = m.textarea.Update(msg)
		m.updateHeight()
		return m, cmd

	case tea.PasteMsg:
		// Pass paste messages to textarea (already filtered by chat.go)
		m.textarea, cmd = m.textarea.Update(msg)
		m.updateHeight()
		return m, cmd
	}

	// Ignore all other message types - they may contain escape sequences
	return m, nil
}

// CleanValue removes any escape sequences from the textarea value.
// Uses the canonical CleanEscapeSequences function from filter.go.
func (m *InputModel) CleanValue() {
	value := m.textarea.Value()
	if cleaned := CleanEscapeSequences(value); cleaned != value {
		m.textarea.SetValue(cleaned)
	}
}

// updateHeight adjusts textarea height based on content.
func (m *InputModel) updateHeight() {
	content := m.textarea.Value()

	// Count actual lines (newlines + 1)
	newlineCount := strings.Count(content, "\n") + 1

	// Also account for line wrapping based on textarea width
	width := m.textarea.Width()
	if width <= 0 {
		width = 80
	}

	wrappedLines := 0
	for _, line := range strings.Split(content, "\n") {
		if len(line) == 0 {
			wrappedLines++
		} else {
			// Estimate wrapped lines (each line takes ceil(len/width) rows)
			wrappedLines += (len(line) + width - 1) / width
		}
	}

	// Use the larger of newline count or wrapped lines
	lines := newlineCount
	if wrappedLines > lines {
		lines = wrappedLines
	}

	// Clamp between 1 and maxHeight
	if lines < 1 {
		lines = 1
	}
	if lines > m.maxHeight {
		lines = m.maxHeight
	}

	m.textarea.SetHeight(lines)
}

// View renders the input component.
func (m InputModel) View() string {
	return m.textarea.View()
}

// Value returns the current input value.
func (m InputModel) Value() string {
	return m.textarea.Value()
}

// SetValue sets the input value.
func (m *InputModel) SetValue(s string) {
	m.textarea.SetValue(s)
}

// Reset clears the input and resets the submitted state.
func (m *InputModel) Reset() {
	m.textarea.SetValue("")
	m.textarea.Reset()
	m.textarea.SetHeight(1) // Reset to single line
	m.submitted = false
}

// Submitted returns true if the user pressed Enter to submit.
func (m InputModel) Submitted() bool {
	return m.submitted
}

// ClearSubmitted resets the submitted flag.
func (m *InputModel) ClearSubmitted() {
	m.submitted = false
}

// Focus focuses the input.
func (m *InputModel) Focus() tea.Cmd {
	m.focused = true
	return m.textarea.Focus()
}

// Blur removes focus from the input.
func (m *InputModel) Blur() {
	m.focused = false
	m.textarea.Blur()
}

// Focused returns whether the input is focused.
func (m InputModel) Focused() bool {
	return m.focused
}

// SetWidth sets the width of the textarea.
func (m *InputModel) SetWidth(w int) {
	m.textarea.SetWidth(w)
}

// SetHeight sets the height of the textarea.
func (m *InputModel) SetHeight(h int) {
	m.textarea.SetHeight(h)
}

// SetPlaceholder sets the placeholder text.
func (m *InputModel) SetPlaceholder(s string) {
	m.placeholder = s
	m.textarea.Placeholder = s
}

// Height returns the current height of the textarea.
func (m InputModel) Height() int {
	return m.textarea.Height()
}

// SetMaxHeight sets the maximum height the textarea can expand to.
func (m *InputModel) SetMaxHeight(h int) {
	m.maxHeight = h
}
