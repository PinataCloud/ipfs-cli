package chat

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// IsPipedInput returns true if stdin is piped (not a terminal).
func IsPipedInput() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// OneShotConfig holds configuration for one-shot mode.
type OneShotConfig struct {
	AgentID          string
	Model            string
	Prompt           string
	GatewayURL       string
	Token            string
	Session          string // Session key for conversation context
	NoColor          bool
	JSONOutput       bool // Force JSONL output
	TextOutput       bool // Force plain text output (overrides JSONL for non-TTY)
	AutoApprove      bool // Auto-approve tool calls (--yes)
	ConversationMode bool // Multi-turn conversation mode
}

// RunOneShot executes a single chat message and streams the response to stdout.
// This is used for piped input mode where the CLI should exit after one response.
// It reads any additional input from stdin and appends it to the prompt.
//
// Output modes:
// - If JSONOutput is true or stdout is not a TTY: JSONL format
// - Otherwise: Human-readable text with markdown rendering
func RunOneShot(cfg OneShotConfig) error {
	// Determine output mode
	// JSON if: explicit --json OR (non-TTY AND not --text)
	useJSON := cfg.JSONOutput || (!IsTerminal() && !cfg.TextOutput)

	// Read stdin if there's piped input
	var stdinContent string
	if IsPipedInput() {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		stdinContent = string(data)
	}

	// Combine prompt with stdin content
	fullPrompt := cfg.Prompt
	if stdinContent != "" {
		if fullPrompt != "" {
			fullPrompt = fullPrompt + "\n\n" + stdinContent
		} else {
			fullPrompt = stdinContent
		}
	}

	if fullPrompt == "" {
		return fmt.Errorf("no input provided")
	}

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Build the messages array with single user message
	messages := []ChatMessage{
		{
			Role:    "user",
			Content: fullPrompt,
		},
	}

	// Start streaming via WebSocket
	events := StreamChat(ctx, cfg.AgentID, cfg.Token, cfg.Model, cfg.Session, messages)

	if useJSON {
		return runOneShotJSON(ctx, events, cfg.AutoApprove)
	}
	return runOneShotText(ctx, events)
}

// runOneShotJSON processes events and outputs JSONL format
func runOneShotJSON(ctx context.Context, events <-chan StreamEvent, autoApprove bool) error {
	writer := NewJSONLWriter()
	tokenIndex := 0

	// Emit message_start
	writer.WriteMessageStart("", "")

	for event := range events {
		switch event.Type {
		case StreamEventToken:
			writer.WriteContentDelta(event.Token, tokenIndex)
			tokenIndex++

		case StreamEventToolCall:
			if event.ToolCall != nil {
				writer.WriteToolUse(event.ToolCall.ID, event.ToolCall.Name, event.ToolCall.Arguments)

				// Handle tool approval
				if !autoApprove {
					// In non-auto mode, emit input_required and wait for approval
					// For now, we skip the tool if not auto-approved
					writer.WriteToolResult(event.ToolCall.ID, "Tool execution skipped: --yes flag not provided", true)
				}
				// If autoApprove is true, the server handles execution and sends tool_result
			}

		case StreamEventToolResult:
			if event.ToolResult != nil {
				writer.WriteToolResult(event.ToolResult.ToolCallID, event.ToolResult.Content, event.ToolResult.IsError)
			}

		case StreamEventDone:
			writer.WriteMessageStop("end_turn")
			return nil

		case StreamEventError:
			if event.Error != nil {
				if ctx.Err() != nil {
					writer.WriteMessageStop("interrupted")
					return nil
				}
				writer.WriteError("stream_error", event.Error.Error())
				return event.Error
			}
		}
	}

	writer.WriteMessageStop("end_turn")
	return nil
}

// runOneShotText processes events and outputs human-readable text
func runOneShotText(ctx context.Context, events <-chan StreamEvent) error {
	styles := DefaultStyles()

	for event := range events {
		switch event.Type {
		case StreamEventToken:
			fmt.Print(event.Token)

		case StreamEventToolCall:
			if event.ToolCall != nil {
				fmt.Println()
				fmt.Println(RenderToolCall(*event.ToolCall, 80, styles))
			}

		case StreamEventToolResult:
			if event.ToolResult != nil {
				fmt.Println(RenderToolResult(*event.ToolResult, 80, styles))
			}

		case StreamEventDone:
			fmt.Println()
			return nil

		case StreamEventError:
			if event.Error != nil {
				if ctx.Err() != nil {
					fmt.Println("\n[interrupted]")
					return nil
				}
				return fmt.Errorf("stream error: %w", event.Error)
			}
		}
	}

	fmt.Println()
	return nil
}

// PrintStreamingToken prints a token for streaming output.
// In non-TTY environments, it buffers output appropriately.
func PrintStreamingToken(token string) {
	fmt.Print(token)
}

// PrintError prints an error message to stderr.
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// RunConversation executes a multi-turn conversation, reading messages from stdin line by line.
// Each line is treated as a separate user message. Responses are streamed to stdout.
// Message history is maintained for context across turns.
func RunConversation(cfg OneShotConfig) error {
	// Determine output mode
	useJSON := cfg.JSONOutput || (!IsTerminal() && !cfg.TextOutput)

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Message history for context
	messages := []ChatMessage{}

	// Read stdin line by line
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// Check for cancellation
		if ctx.Err() != nil {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Add user message to history
		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: line,
		})

		// Stream response
		events := StreamChat(ctx, cfg.AgentID, cfg.Token, cfg.Model, cfg.Session, messages)

		// Process events and collect response
		var response strings.Builder
		if useJSON {
			err := processConversationJSON(ctx, events, &response, cfg.AutoApprove)
			if err != nil {
				return err
			}
		} else {
			err := processConversationText(ctx, events, &response)
			if err != nil {
				return err
			}
		}

		// Add assistant response to history
		if response.Len() > 0 {
			messages = append(messages, ChatMessage{
				Role:    "assistant",
				Content: response.String(),
			})
		}
	}

	return scanner.Err()
}

// processConversationJSON processes events for conversation mode with JSONL output.
func processConversationJSON(ctx context.Context, events <-chan StreamEvent, response *strings.Builder, autoApprove bool) error {
	writer := NewJSONLWriter()
	tokenIndex := 0

	writer.WriteMessageStart("", "")

	for event := range events {
		switch event.Type {
		case StreamEventToken:
			writer.WriteContentDelta(event.Token, tokenIndex)
			response.WriteString(event.Token)
			tokenIndex++

		case StreamEventToolCall:
			if event.ToolCall != nil {
				writer.WriteToolUse(event.ToolCall.ID, event.ToolCall.Name, event.ToolCall.Arguments)
				if !autoApprove {
					writer.WriteToolResult(event.ToolCall.ID, "Tool execution skipped: --yes flag not provided", true)
				}
			}

		case StreamEventToolResult:
			if event.ToolResult != nil {
				writer.WriteToolResult(event.ToolResult.ToolCallID, event.ToolResult.Content, event.ToolResult.IsError)
			}

		case StreamEventDone:
			writer.WriteMessageStop("end_turn")
			return nil

		case StreamEventError:
			if event.Error != nil {
				if ctx.Err() != nil {
					writer.WriteMessageStop("interrupted")
					return nil
				}
				writer.WriteError("stream_error", event.Error.Error())
				return event.Error
			}
		}
	}

	writer.WriteMessageStop("end_turn")
	return nil
}

// processConversationText processes events for conversation mode with plain text output.
func processConversationText(ctx context.Context, events <-chan StreamEvent, response *strings.Builder) error {
	styles := DefaultStyles()

	for event := range events {
		switch event.Type {
		case StreamEventToken:
			fmt.Print(event.Token)
			response.WriteString(event.Token)

		case StreamEventToolCall:
			if event.ToolCall != nil {
				fmt.Println()
				fmt.Println(RenderToolCall(*event.ToolCall, 80, styles))
			}

		case StreamEventToolResult:
			if event.ToolResult != nil {
				fmt.Println(RenderToolResult(*event.ToolResult, 80, styles))
			}

		case StreamEventDone:
			fmt.Println()
			fmt.Println("---") // Delimiter between turns
			return nil

		case StreamEventError:
			if event.Error != nil {
				if ctx.Err() != nil {
					fmt.Println("\n[interrupted]")
					return nil
				}
				return fmt.Errorf("stream error: %w", event.Error)
			}
		}
	}

	fmt.Println()
	fmt.Println("---")
	return nil
}
