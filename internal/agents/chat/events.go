package chat

import (
	"encoding/json"
	"fmt"
	"os"
)

// Event types for JSONL output (Anthropic-compatible naming)
const (
	EventTypeMessageStart  = "message_start"
	EventTypeContentDelta  = "content_delta"
	EventTypeToolUse       = "tool_use"
	EventTypeToolResult    = "tool_result"
	EventTypeMessageStop   = "message_stop"
	EventTypeError         = "error"
	EventTypeInputRequired = "input_required" // For tool approval in bidirectional mode
)

// JSONLEvent is the base structure for all JSONL events
type JSONLEvent struct {
	Type string `json:"type"`
}

// MessageStartEvent signals the start of a response
type MessageStartEvent struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id,omitempty"`
	Model     string `json:"model,omitempty"`
}

// ContentDelta contains a text token or content block
type ContentDelta struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// ContentDeltaEvent wraps a content delta
type ContentDeltaEvent struct {
	Type  string       `json:"type"`
	Index int          `json:"index,omitempty"`
	Delta ContentDelta `json:"delta"`
}

// ToolUseEvent represents a tool call request
type ToolUseEvent struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResultEvent represents the result of a tool execution
type ToolResultEvent struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// MessageStopEvent signals the end of a response
type MessageStopEvent struct {
	Type       string `json:"type"`
	StopReason string `json:"stop_reason"` // "end_turn", "tool_use", "max_tokens"
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ErrorEvent represents an error
type ErrorEvent struct {
	Type  string      `json:"type"`
	Error ErrorDetail `json:"error"`
}

// InputRequiredEvent signals that input is needed (for bidirectional tool approval)
type InputRequiredEvent struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	ToolName  string `json:"tool_name"`
	Message   string `json:"message"`
}

// ToolApprovalInput is the expected input format for tool approval
type ToolApprovalInput struct {
	Approve bool `json:"approve"`
}

// JSONLWriter handles writing JSONL events to stdout
type JSONLWriter struct{}

// NewJSONLWriter creates a new JSONL writer
func NewJSONLWriter() *JSONLWriter {
	return &JSONLWriter{}
}

// Write outputs a JSONL event to stdout
func (w *JSONLWriter) Write(event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

// WriteMessageStart emits a message_start event
func (w *JSONLWriter) WriteMessageStart(messageID, model string) error {
	return w.Write(MessageStartEvent{
		Type:      EventTypeMessageStart,
		MessageID: messageID,
		Model:     model,
	})
}

// WriteContentDelta emits a content_delta event
func (w *JSONLWriter) WriteContentDelta(text string, index int) error {
	return w.Write(ContentDeltaEvent{
		Type:  EventTypeContentDelta,
		Index: index,
		Delta: ContentDelta{
			Type: "text",
			Text: text,
		},
	})
}

// WriteToolUse emits a tool_use event
func (w *JSONLWriter) WriteToolUse(id, name string, input map[string]interface{}) error {
	return w.Write(ToolUseEvent{
		Type:  EventTypeToolUse,
		ID:    id,
		Name:  name,
		Input: input,
	})
}

// WriteToolResult emits a tool_result event
func (w *JSONLWriter) WriteToolResult(toolUseID, content string, isError bool) error {
	return w.Write(ToolResultEvent{
		Type:      EventTypeToolResult,
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	})
}

// WriteMessageStop emits a message_stop event
func (w *JSONLWriter) WriteMessageStop(stopReason string) error {
	return w.Write(MessageStopEvent{
		Type:       EventTypeMessageStop,
		StopReason: stopReason,
	})
}

// WriteError emits an error event
func (w *JSONLWriter) WriteError(errType, message string) error {
	return w.Write(ErrorEvent{
		Type: EventTypeError,
		Error: ErrorDetail{
			Type:    errType,
			Message: message,
		},
	})
}

// WriteInputRequired emits an input_required event for tool approval
func (w *JSONLWriter) WriteInputRequired(toolUseID, toolName, message string) error {
	return w.Write(InputRequiredEvent{
		Type:      EventTypeInputRequired,
		ToolUseID: toolUseID,
		ToolName:  toolName,
		Message:   message,
	})
}
