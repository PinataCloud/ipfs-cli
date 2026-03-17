package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strings"

	"pinata/internal/common"
	"pinata/internal/config"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// State represents the current state of the chat TUI.
type State int

const (
	StateIdle        State = iota // User can type, history visible
	StateStreaming                // Receiving response, spinner shown
	StateToolPending              // Tool call waiting for permission
	StateToolRunning              // Tool executing
)

// ChatConfig holds configuration for the chat TUI.
type ChatConfig struct {
	AgentID     string
	AgentName   string // Display name for the agent
	AgentEmoji  string // Emoji for the agent
	AgentStatus string // Current status (running, etc.)
	GatewayURL  string // Base URL for the chat endpoint (e.g., "https://gateway.pinata.cloud")
	Token       string // API token for authentication
	Model       string // Model override (empty for agent default)
	Session     string // Session key for conversation context
}

// ChatModel is the main bubbletea model for the chat TUI.
type ChatModel struct {
	// Configuration
	config ChatConfig

	// State
	state   State
	err     error
	ready   bool
	quiting bool

	// Components
	input    InputModel
	viewport viewport.Model
	spinner  spinner.Model
	styles   Styles

	// Messages
	messages    []ChatMessage
	currentResp *strings.Builder

	// Tool handling
	pendingTool *ToolCall
	permissions *ToolPermissions

	// Streaming
	streamCtx    context.Context
	streamCancel context.CancelFunc
	programRef   *tea.Program // Reference to program for event injection

	// Message queue for messages submitted during streaming
	messageQueue []string

	// Viewport content hash to avoid redundant updates
	lastContentHash uint64

	// Dimensions
	width  int
	height int
}

// programReadyMsg is sent when the program is ready to receive events.
type programReadyMsg struct {
	program *tea.Program
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// streamToProgram reads from the event channel and injects events into bubbletea.
// This runs in a separate goroutine to avoid blocking the event loop.
func streamToProgram(ctx context.Context, p *tea.Program, events <-chan StreamEvent) {
	for {
		select {
		case event, ok := <-events:
			if !ok {
				// Channel closed
				p.Send(streamEventMsg{event: StreamEvent{Type: StreamEventDone}})
				return
			}
			p.Send(streamEventMsg{event: event})
		case <-ctx.Done():
			return
		}
	}
}

// NewChatModel creates a new chat TUI model.
func NewChatModel(cfg ChatConfig) ChatModel {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Initialize styles
	styles := DefaultStyles()
	s.Style = styles.SpinnerStyle

	// Initialize input
	input := NewInputModel("Type a message...")

	return ChatModel{
		config:      cfg,
		state:       StateIdle,
		input:       input,
		spinner:     s,
		styles:      styles,
		messages:    []ChatMessage{},
		currentResp: &strings.Builder{},
		permissions: NewToolPermissions(),
		width:       80,
		height:      24,
	}
}

// streamEventMsg wraps a StreamEvent for the bubbletea message system.
type streamEventMsg struct {
	event StreamEvent
}

// sendQueuedMsg triggers sending a queued message after stream completion.
type sendQueuedMsg struct {
	text string
}

// Init initializes the chat model.
func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		m.spinner.Tick,
	)
}

// Update handles messages and updates the model.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Initialize viewport if not ready
		if !m.ready {
			m.viewport = viewport.New(viewport.WithWidth(msg.Width), viewport.WithHeight(10))
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
		}

		// Update input width and max height
		m.input.SetWidth(msg.Width - 2)
		maxInputHeight := (msg.Height - 8) / 2 // Allow input to use up to half the remaining space
		if maxInputHeight < 1 {
			maxInputHeight = 1
		}
		if maxInputHeight > 10 {
			maxInputHeight = 10
		}
		m.input.SetMaxHeight(maxInputHeight)

		// Update viewport height based on current input
		m.updateViewportHeight()
		return m, nil

	case tea.KeyPressMsg:
		// Filter terminal escape sequence responses before handling
		keyStr := msg.String()
		if IsEscapeSequence(keyStr) {
			return m, nil
		}
		return m.handleKeyPress(msg)

	case tea.MouseWheelMsg, tea.MouseClickMsg, tea.MouseMotionMsg, tea.MouseReleaseMsg:
		// Ignore mouse events - mouse mode is disabled to prevent escape sequence issues
		return m, nil

	case tea.FocusMsg, tea.BlurMsg:
		// Handle focus changes - clean input when focus returns
		if _, ok := msg.(tea.FocusMsg); ok {
			m.input.CleanValue()
		}
		return m, nil

	case tea.BackgroundColorMsg, tea.ForegroundColorMsg, tea.CursorColorMsg,
		tea.CursorPositionMsg, tea.ColorProfileMsg, tea.CapabilityMsg,
		tea.ModeReportMsg, tea.KeyboardEnhancementsMsg:
		// Ignore terminal query responses - these should not affect UI
		return m, nil

	case tea.PasteMsg:
		// Filter paste messages that contain terminal escape sequences
		if IsEscapeSequence(msg.Content) {
			return m, nil
		}
		// Pass clean paste to input (only when idle or streaming)
		if m.state == StateIdle || m.state == StateStreaming {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.updateViewportHeight()
			return m, cmd
		}
		return m, nil

	case tea.PasteStartMsg, tea.PasteEndMsg:
		// Ignore paste boundary markers
		return m, nil

	case spinner.TickMsg:
		if m.state == StateStreaming || m.state == StateToolRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case streamEventMsg:
		return m.handleStreamEvent(msg.event)

	case programReadyMsg:
		m.programRef = msg.program
		return m, nil

	case sendQueuedMsg:
		// Send the queued message - don't reset input (user might be typing)
		return m.sendMessage(msg.text, true)
	}

	// Note: KeyPressMsg is handled in handleKeyPress(), PasteMsg is handled above
	// Do NOT pass other message types to input - they may contain escape sequences

	// Handle viewport updates
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updateViewportHeight adjusts viewport height based on input size.
// Layout: header(1) + divider(1) + viewport + divider(1) + [status(1) +] input + help(1) + newlines
func (m *ChatModel) updateViewportHeight() {
	// Fixed chrome: header(1) + top divider(1) + bottom divider(1) + help(1) + newlines/padding(2) = 6 lines
	// Plus input height (dynamic)
	// Plus status line (1) during streaming
	inputHeight := m.input.Height()
	chrome := 6
	if m.state == StateStreaming || m.state == StateToolRunning {
		chrome = 7 // Extra line for status during streaming
	}
	newHeight := m.height - chrome - inputHeight
	if newHeight < 3 {
		newHeight = 3 // Minimum viewport height
	}
	if m.viewport.Height() != newHeight {
		m.viewport.SetHeight(newHeight)
		m.setViewportContent(m.renderMessages())
	}
}

// setViewportContent updates viewport content only if it changed.
// Uses FNV hash to avoid redundant re-renders during streaming.
func (m *ChatModel) setViewportContent(content string) {
	h := fnv.New64a()
	h.Write([]byte(content))
	hash := h.Sum64()

	if hash != m.lastContentHash {
		m.lastContentHash = hash
		m.viewport.SetContent(content)
	}
}

// handleKeyPress handles keyboard input.
func (m ChatModel) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Check for Ctrl+C
	if keyStr == "ctrl+c" {
		// Cancel streaming if active, otherwise quit
		if m.state == StateStreaming || m.state == StateToolRunning {
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.state = StateIdle
			m.setViewportContent(m.renderMessages())
			return m, m.input.Focus()
		}
		m.quiting = true
		return m, tea.Quit
	}

	// Check for Ctrl+D
	if keyStr == "ctrl+d" {
		m.quiting = true
		return m, tea.Quit
	}

	// Check for Escape
	if keyStr == "esc" {
		// Cancel current operation or quit
		if m.state == StateStreaming || m.state == StateToolRunning {
			if m.streamCancel != nil {
				m.streamCancel()
			}
			m.state = StateIdle
			m.setViewportContent(m.renderMessages())
			return m, m.input.Focus()
		}
		if m.state == StateToolPending {
			// Deny tool
			return m.denyTool()
		}
		// Quit when idle
		m.quiting = true
		return m, tea.Quit
	}

	// Handle scroll keys (page up/down) in any state
	switch keyStr {
	case "pgup":
		m.viewport.ScrollUp(m.viewport.Height())
		return m, nil
	case "pgdown":
		m.viewport.ScrollDown(m.viewport.Height())
		return m, nil
	}

	// Handle tool permission keys
	if m.state == StateToolPending && m.pendingTool != nil {
		switch keyStr {
		case "a": // Allow
			return m.allowTool(false)
		case "d": // Deny
			return m.denyTool()
		case "A": // Always allow
			return m.allowTool(true)
		case "N": // Never allow
			m.permissions.SetNeverAllow(m.pendingTool.Name)
			return m.denyTool()
		}
	}

	// Pass to input when idle or streaming (allow typing during streaming)
	if m.state == StateIdle || m.state == StateStreaming {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)

		// Update viewport height in case input height changed (e.g., shift+enter)
		m.updateViewportHeight()

		// Check if user submitted
		if m.input.Submitted() {
			m.input.ClearSubmitted()
			// Clean any escape sequences before getting the value
			m.input.CleanValue()
			if text := strings.TrimSpace(m.input.Value()); text != "" {
				if m.state == StateStreaming {
					// Queue message for after streaming completes
					m.messageQueue = append(m.messageQueue, text)
					m.input.Reset()
					return m, cmd
				}
				return m.sendMessage(text)
			}
		}

		return m, cmd
	}

	return m, nil
}

// sendMessage sends a message to the chat endpoint.
// If fromQueue is true, the input won't be reset (preserving what user is typing).
func (m ChatModel) sendMessage(text string, fromQueue ...bool) (tea.Model, tea.Cmd) {
	// Add user message to history
	m.messages = append(m.messages, ChatMessage{
		Role:    "user",
		Content: text,
	})

	// Clear input unless this is a queued message (user might be typing next message)
	if len(fromQueue) == 0 || !fromQueue[0] {
		m.input.Reset()
	}

	// Update state
	m.state = StateStreaming
	m.currentResp.Reset()

	// Create stream context
	m.streamCtx, m.streamCancel = context.WithCancel(context.Background())

	// Create the stream and start the goroutine to inject events
	events := StreamChat(m.streamCtx, m.config.AgentID, m.config.Token, m.config.Model, m.config.Session, m.messages)

	// Start goroutine to inject events via Program.Send() - non-blocking!
	if m.programRef != nil {
		go streamToProgram(m.streamCtx, m.programRef, events)
	}

	// Update viewport
	m.setViewportContent(m.renderMessages())
	m.viewport.GotoBottom()

	// Return only spinner tick - events will arrive via Program.Send()
	return m, m.spinner.Tick
}

// handleStreamEvent handles events from the stream.
// Events arrive via Program.Send() from the streamToProgram goroutine.
func (m ChatModel) handleStreamEvent(event StreamEvent) (tea.Model, tea.Cmd) {
	switch event.Type {
	case StreamEventToken:
		m.currentResp.WriteString(event.Token)
		m.setViewportContent(m.renderMessages())
		m.viewport.GotoBottom()
		// Events arrive via Program.Send() - just keep spinner active
		return m, m.spinner.Tick

	case StreamEventToolCall:
		if event.ToolCall != nil {
			// Display tool call info (tools run server-side in OpenClaw)
			m.pendingTool = event.ToolCall
			m.state = StateToolRunning
			m.setViewportContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
		// Events arrive via Program.Send() - just keep spinner active
		return m, m.spinner.Tick

	case StreamEventToolResult:
		if event.ToolResult != nil {
			if m.pendingTool != nil {
				m.pendingTool.Status = ToolStatusDone
			}
			// Resume streaming after tool result
			m.state = StateStreaming
			m.pendingTool = nil
			m.setViewportContent(m.renderMessages())
		}
		// Events arrive via Program.Send() - just keep spinner active
		return m, m.spinner.Tick

	case StreamEventDone:
		// Finalize assistant message
		if m.currentResp.Len() > 0 {
			m.messages = append(m.messages, ChatMessage{
				Role:    "assistant",
				Content: m.currentResp.String(),
			})
			m.currentResp.Reset()
		}

		// Clean up old stream context
		if m.streamCancel != nil {
			m.streamCancel()
			m.streamCancel = nil
		}
		m.streamCtx = nil

		m.setViewportContent(m.renderMessages())
		m.viewport.GotoBottom()

		// Check for queued messages
		if len(m.messageQueue) > 0 {
			text := m.messageQueue[0]
			m.messageQueue = m.messageQueue[1:]
			// Small delay to ensure clean state before starting new stream
			return m, tea.Batch(
				m.input.Focus(),
				func() tea.Msg {
					return sendQueuedMsg{text: text}
				},
			)
		}

		m.state = StateIdle
		return m, m.input.Focus()

	case StreamEventError:
		m.err = event.Error
		m.state = StateIdle

		// Add error as system message
		if event.Error != nil {
			m.messages = append(m.messages, ChatMessage{
				Role:    "system",
				Content: fmt.Sprintf("Error: %v", event.Error),
			})
		}

		m.setViewportContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, m.input.Focus()
	}

	return m, nil
}


// allowTool approves the pending tool call.
func (m ChatModel) allowTool(always bool) (tea.Model, tea.Cmd) {
	if m.pendingTool == nil {
		m.state = StateIdle
		return m, m.input.Focus()
	}

	if always {
		m.permissions.SetAlwaysAllow(m.pendingTool.Name)
	}

	m.pendingTool.Status = ToolStatusRunning
	m.state = StateToolRunning
	m.setViewportContent(m.renderMessages())

	// In a real implementation, you would send the tool approval to the server
	// and continue the stream. For now, we simulate completion.
	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			// Simulate tool completion
			return streamEventMsg{event: StreamEvent{
				Type: StreamEventToolResult,
				ToolResult: &ToolResult{
					ToolCallID: m.pendingTool.ID,
					Content:    "[Tool executed - result would appear here]",
				},
			}}
		},
	)
}

// denyTool denies the pending tool call.
func (m ChatModel) denyTool() (tea.Model, tea.Cmd) {
	if m.pendingTool != nil {
		m.pendingTool.Status = ToolStatusDenied
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Tool '%s' was denied", m.pendingTool.Name),
		})
	}

	m.pendingTool = nil
	m.state = StateIdle
	m.setViewportContent(m.renderMessages())
	return m, m.input.Focus()
}

// renderMessages renders all messages for the viewport.
func (m ChatModel) renderMessages() string {
	var sb strings.Builder

	for _, msg := range m.messages {
		sb.WriteString(m.renderMessage(msg))
		sb.WriteString("\n")
	}

	// Render current streaming response - use raw text for performance
	// Markdown rendering happens only when the message is finalized
	if m.state == StateStreaming && m.currentResp.Len() > 0 {
		agentLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		border := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("│")
		// Use agent name if available
		streamingLabel := "Agent"
		if m.config.AgentName != "" {
			streamingLabel = m.config.AgentName
		}
		if m.config.AgentEmoji != "" {
			streamingLabel = m.config.AgentEmoji + " " + streamingLabel
		}
		sb.WriteString(agentLabel.Render(streamingLabel))
		sb.WriteString("\n")
		// Show raw text during streaming for performance
		for _, line := range strings.Split(m.currentResp.String(), "\n") {
			sb.WriteString(border)
			sb.WriteString(" ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	// Render pending tool
	if m.state == StateToolPending && m.pendingTool != nil {
		sb.WriteString(RenderToolCall(*m.pendingTool, m.width, m.styles))
		sb.WriteString("\n")
		sb.WriteString(RenderPermissionPrompt(*m.pendingTool, m.styles))
		sb.WriteString("\n")
	}

	// Render running tool
	if m.state == StateToolRunning && m.pendingTool != nil {
		sb.WriteString(RenderToolCall(*m.pendingTool, m.width, m.styles))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderMessage renders a single message.
func (m ChatModel) renderMessage(msg ChatMessage) string {
	var sb strings.Builder

	// Role indicator styles - colored labels to distinguish roles
	youLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	agentLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	toolLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	systemLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

	// Border character - subtle gray, same for all messages
	border := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("│")

	switch msg.Role {
	case "user":
		sb.WriteString(youLabel.Render("You"))
		sb.WriteString("\n")
		for _, line := range strings.Split(msg.Content, "\n") {
			sb.WriteString(border)
			sb.WriteString(" ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}

	case "assistant":
		// Use agent name if available, otherwise "Agent"
		agentName := "Agent"
		if m.config.AgentName != "" {
			agentName = m.config.AgentName
		}
		if m.config.AgentEmoji != "" {
			agentName = m.config.AgentEmoji + " " + agentName
		}
		sb.WriteString(agentLabel.Render(agentName))
		sb.WriteString("\n")
		content := RenderMarkdown(msg.Content, m.width-4, NoColor())
		for _, line := range strings.Split(strings.TrimSuffix(content, "\n"), "\n") {
			sb.WriteString(border)
			sb.WriteString(" ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}

	case "tool":
		sb.WriteString(toolLabel.Render("Tool"))
		sb.WriteString("\n")
		sb.WriteString(border)
		sb.WriteString(" ")
		sb.WriteString(msg.Content)

	case "system":
		sb.WriteString(systemLabel.Render("  " + msg.Content))
	}

	return sb.String()
}

// View renders the chat TUI.
func (m ChatModel) View() tea.View {
	if m.quiting {
		return tea.NewView("")
	}

	if !m.ready {
		return tea.NewView("\n  Connecting...\n")
	}

	var sb strings.Builder

	// Header - show agent name, emoji, and status
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		PaddingLeft(1)
	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	// Build agent display: emoji + name (or just ID if no name)
	agentDisplay := m.config.AgentID
	if m.config.AgentName != "" {
		agentDisplay = m.config.AgentName
	}
	if m.config.AgentEmoji != "" {
		agentDisplay = m.config.AgentEmoji + " " + agentDisplay
	}

	// Status indicator
	statusColor := "8" // gray default
	statusText := m.config.AgentStatus
	if m.config.AgentStatus == "running" {
		statusColor = "10" // green
		statusText = "●"
	} else if m.config.AgentStatus == "starting" {
		statusColor = "11" // yellow
		statusText = "○"
	} else if m.config.AgentStatus == "not_running" {
		statusColor = "9" // red
		statusText = "○"
	}
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))

	sb.WriteString(nameStyle.Render(agentDisplay))
	sb.WriteString(" ")
	sb.WriteString(statusStyle.Render(statusText))
	sb.WriteString(headerStyle.Render(fmt.Sprintf(" (%s)", m.config.AgentID)))
	if m.config.Session != "" && m.config.Session != "agent:main:cli" {
		sb.WriteString(headerStyle.Render(fmt.Sprintf(" · %s", m.config.Session)))
	}
	sb.WriteString("\n")

	// Divider
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Repeat("─", m.width))
	sb.WriteString(divider)
	sb.WriteString("\n")

	// Viewport (message history)
	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	// Bottom divider
	sb.WriteString(divider)
	sb.WriteString("\n")

	// Status line and/or input
	if m.state == StateStreaming || m.state == StateToolRunning {
		sb.WriteString(m.renderStatusLine())
		// Show prominent queue indicator if messages are queued
		if len(m.messageQueue) > 0 {
			queueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")).
				Bold(true)
			sb.WriteString("  ")
			sb.WriteString(queueStyle.Render(fmt.Sprintf("[%d queued]", len(m.messageQueue))))
		}
		sb.WriteString("\n")
		// Show input during streaming so user can type next message
		promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		sb.WriteString(promptStyle.Render("> "))
		sb.WriteString(m.input.View())
	} else if m.state == StateToolPending {
		sb.WriteString(m.renderStatusLine())
	} else {
		// Input prompt
		promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		sb.WriteString(promptStyle.Render("> "))
		sb.WriteString(m.input.View())
	}

	// Help line
	sb.WriteString("\n")
	sb.WriteString(m.renderHelpLine())

	v := tea.NewView(sb.String())
	v.AltScreen = true
	// Disable mouse mode - it causes too many escape sequence issues
	// Users can still scroll with pgup/pgdn
	v.KeyboardEnhancements = tea.KeyboardEnhancements{}
	return v
}

// renderStatusLine renders the status indicator.
func (m ChatModel) renderStatusLine() string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	var status string
	switch m.state {
	case StateStreaming:
		status = m.spinner.View() + " Thinking..."
	case StateToolPending:
		if m.pendingTool != nil {
			status = "Tool permission: " + toolStyle.Render(m.pendingTool.Name)
		} else {
			status = "Tool permission required"
		}
	case StateToolRunning:
		if m.pendingTool != nil {
			toolInfo := m.pendingTool.Name
			// Show truncated arguments if available
			if len(m.pendingTool.Arguments) > 0 {
				// Format first key=value pair for preview
				for k, v := range m.pendingTool.Arguments {
					args := truncateString(fmt.Sprintf("%s=%v", k, v), 25)
					toolInfo += " " + args
					break // Just show first argument
				}
			}
			status = m.spinner.View() + " " + toolStyle.Render(toolInfo)
		} else {
			status = m.spinner.View() + " Running tool..."
		}
	default:
		status = ""
	}
	// Pad to full width to overwrite any previous content
	if m.width > len(status) {
		status = status + strings.Repeat(" ", m.width-len(status)-2)
	}
	return statusStyle.Render(status)
}

// renderHelpLine renders the help text.
func (m ChatModel) renderHelpLine() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Faint(true)
	var help string

	switch m.state {
	case StateIdle:
		help = "enter send · shift+enter newline · ctrl+u clear · esc quit"
	case StateStreaming, StateToolRunning:
		if len(m.messageQueue) > 0 {
			help = fmt.Sprintf("%d queued · type more · esc cancel", len(m.messageQueue))
		} else {
			help = "type to queue · enter queue · esc cancel"
		}
	case StateToolPending:
		help = "a allow · d deny · A always · N never"
	}

	// Add scroll hint if not already shown
	if m.state == StateIdle && m.viewport.TotalLineCount() > m.viewport.Height() {
		help += " · pgup/pgdn scroll"
	}

	return helpStyle.Render(help)
}

// StartChat launches the interactive chat TUI.
// gatewayURL is the base URL for the chat endpoint.
// token is the API token for authentication.
// model is optional - if empty, the agent's default model is used.
//
// Example gatewayURL: "https://gateway.pinata.cloud" or from agent.GatewayToken
// StartChat starts a chat session with an agent.
//
// Output mode selection:
//   - If jsonOutput is true or stdout is not a TTY: JSONL streaming mode
//   - Otherwise: Interactive TUI mode
//
// The autoApprove flag controls tool execution:
//   - If true: Tools are automatically approved without prompting
//   - If false: Interactive mode prompts for permission, non-interactive mode skips tools
//
// If gatewayURL or token are empty, they are fetched from the agent details.
func StartChat(agentID, gatewayURL, token, model string, jsonOutput, textOutput, conversationMode, autoApprove bool, prompt, session string) error {
	// Always fetch agent info for metadata (name, emoji, status)
	agent, err := FetchAgentForChat(agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent details: %w", err)
	}

	if token == "" {
		token = agent.GatewayToken
	}
	if gatewayURL == "" {
		gatewayURL = GetDefaultGatewayURL()
	}

	cfg := ChatConfig{
		AgentID:     agentID,
		AgentName:   agent.Name,
		AgentEmoji:  agent.Emoji,
		AgentStatus: agent.Status,
		GatewayURL:  gatewayURL,
		Token:       token,
		Model:       model,
		Session:     session,
	}

	// Build one-shot config for non-interactive modes
	oneShotCfg := OneShotConfig{
		AgentID:          agentID,
		GatewayURL:       gatewayURL,
		Token:            token,
		Model:            model,
		Prompt:           prompt,
		Session:          session,
		NoColor:          NoColor(),
		JSONOutput:       jsonOutput,
		TextOutput:       textOutput,
		AutoApprove:      autoApprove,
		ConversationMode: conversationMode,
	}

	// Multi-turn conversation mode
	if conversationMode {
		return RunConversation(oneShotCfg)
	}

	// Determine if we should use non-interactive mode
	// Non-interactive mode when: stdout is not a TTY, --json flag is set, piped input, or prompt provided
	useNonInteractive := jsonOutput || textOutput || !IsTerminal() || IsPipedInput() || prompt != ""

	if useNonInteractive {
		return RunOneShot(oneShotCfg)
	}

	// Run interactive TUI
	p := tea.NewProgram(NewChatModel(cfg))

	// Send program reference to model for event injection
	go func() {
		p.Send(programReadyMsg{program: p})
	}()

	_, runErr := p.Run()
	return runErr
}

// FetchAgentForChat fetches agent details needed for chat.
// This is a lightweight wrapper that imports from the agents package.
func FetchAgentForChat(agentID string) (*AgentInfo, error) {
	// We need to call the agents API to get the gateway token
	// This is implemented in the agents package
	return fetchAgentInfo(agentID)
}

// AgentInfo contains agent info needed for chat.
type AgentInfo struct {
	AgentID      string
	Name         string
	Description  string
	Emoji        string
	GatewayToken string
	Status       string
}

// GetDefaultGatewayURL returns the default gateway URL for chat.
func GetDefaultGatewayURL() string {
	// Check environment variable first
	if url := getEnv("PINATA_GATEWAY_URL", ""); url != "" {
		return url
	}
	// Default gateway URL - same as agents API
	return "https://agents.pinata.cloud"
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// fetchAgentInfo fetches agent details from the API.
func fetchAgentInfo(agentID string) (*AgentInfo, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	url := fmt.Sprintf("https://%s/v0/agents/%s", config.GetAgentsHost(), agentID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get agent: status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Agent struct {
			AgentID      string  `json:"agentId"`
			Name         string  `json:"name"`
			Description  *string `json:"description"`
			Emoji        *string `json:"emoji"`
			GatewayToken string  `json:"gatewayToken"`
			Status       string  `json:"status"`
		} `json:"agent"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	info := &AgentInfo{
		AgentID:      result.Agent.AgentID,
		Name:         result.Agent.Name,
		GatewayToken: result.Agent.GatewayToken,
		Status:       result.Agent.Status,
	}
	if result.Agent.Description != nil {
		info.Description = *result.Agent.Description
	}
	if result.Agent.Emoji != nil {
		info.Emoji = *result.Agent.Emoji
	}
	return info, nil
}

// Header style for the chat window
var headerStyle = lipgloss.NewStyle().
	Bold(true).
	Padding(0, 1)
