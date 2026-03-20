package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// WebSocket timeouts
	wsConnectTimeout = 30 * time.Second
	wsReadTimeout    = 60 * time.Second
	wsPongWait       = 60 * time.Second
)

// StreamEventType represents the type of streaming event.
type StreamEventType string

const (
	StreamEventToken      StreamEventType = "token"
	StreamEventToolCall   StreamEventType = "tool_call"
	StreamEventToolResult StreamEventType = "tool_result"
	StreamEventDone       StreamEventType = "done"
	StreamEventError      StreamEventType = "error"
)

// StreamEvent represents an event from the WebSocket stream.
type StreamEvent struct {
	Type       StreamEventType
	Token      string      // For token events
	ToolCall   *ToolCall   // For tool_call events
	ToolResult *ToolResult // For tool_result events
	Error      error       // For error events
}

// ChatMessage represents a message in the conversation.
type ChatMessage struct {
	Role       string      `json:"role"` // user, assistant, tool
	Content    string      `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// OpenClaw protocol message types
type OpenClawMessage struct {
	Type    string          `json:"type"`              // "req", "res", "event"
	ID      string          `json:"id,omitempty"`      // Request/response ID
	Method  string          `json:"method,omitempty"`  // For requests
	Params  json.RawMessage `json:"params,omitempty"`  // For requests
	Event   string          `json:"event,omitempty"`   // For events
	Payload json.RawMessage `json:"payload,omitempty"` // For events/responses
	OK      *bool           `json:"ok,omitempty"`      // For responses
	Error   *OpenClawError  `json:"error,omitempty"`   // For error responses
}

// OpenClawError represents an error in the OpenClaw protocol.
type OpenClawError struct {
	Code    interface{} `json:"code"` // Can be string or int
	Message string      `json:"message"`
}

// ConnectChallenge is the server's challenge payload
type ConnectChallenge struct {
	Nonce string `json:"nonce"`
	TS    int64  `json:"ts"`
}

// ConnectParams for the connect request
type ConnectParams struct {
	MinProtocol int              `json:"minProtocol"`
	MaxProtocol int              `json:"maxProtocol"`
	Client      ConnectClient    `json:"client"`
	Role        string           `json:"role"`
	Scopes      []string         `json:"scopes"`
	Caps        []string         `json:"caps"`
	Commands    []string         `json:"commands"`
	Permissions map[string]any   `json:"permissions"`
	Auth        ConnectAuth      `json:"auth"`
}

// ConnectClient identifies the client
type ConnectClient struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Mode     string `json:"mode"`
}

// ConnectAuth contains authentication details.
type ConnectAuth struct {
	Token string `json:"token"`
}

// ChatSendParams for the chat.send method
type ChatSendParams struct {
	Message        string `json:"message"`
	IdempotencyKey string `json:"idempotencyKey"`
	SessionKey     string `json:"sessionKey"`
}

// AgentEventPayload represents an agent event payload
type AgentEventPayload struct {
	Stream     string                 `json:"stream"`     // "assistant", "lifecycle", "tool", etc.
	RunID      string                 `json:"runId"`
	SessionKey string                 `json:"sessionKey"`
	Seq        int                    `json:"seq"`
	TS         int64                  `json:"ts"`
	Data       map[string]interface{} `json:"data"`
}

// ChatEventPayload represents a chat event payload
type ChatEventPayload struct {
	State      string                 `json:"state"` // "delta", "final"
	RunID      string                 `json:"runId"`
	SessionKey string                 `json:"sessionKey"`
	Seq        int                    `json:"seq"`
	Message    map[string]interface{} `json:"message"`
}

// BuildGatewayURL constructs the WebSocket URL for an agent's gateway.
func BuildGatewayURL(agentID string) string {
	return fmt.Sprintf("wss://%s.agents.pinata.cloud", agentID)
}

// StreamChat connects to the agent's gateway via WebSocket and streams responses.
func StreamChat(ctx context.Context, agentID, token, model, session string, messages []ChatMessage) <-chan StreamEvent {
	events := make(chan StreamEvent, 100)

	go func() {
		defer close(events)

		// Build gateway URL
		gatewayURL := BuildGatewayURL(agentID)

		// Connect to WebSocket with proper headers and timeout
		header := http.Header{}
		header.Set("Origin", "https://agents.pinata.cloud")
		if token != "" {
			header.Set("Authorization", "Bearer "+token)
		}

		// Create dialer with timeout
		dialer := websocket.Dialer{
			HandshakeTimeout: wsConnectTimeout,
		}

		conn, _, err := dialer.DialContext(ctx, gatewayURL, header)
		if err != nil {
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("failed to connect: %w", err)}
			return
		}
		defer conn.Close()

		// Set up context cancellation to close the connection
		go func() {
			<-ctx.Done()
			conn.Close()
		}()

		// Step 1: Wait for connect.challenge event
		conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
		var challengeMsg OpenClawMessage
		if err := conn.ReadJSON(&challengeMsg); err != nil {
			if ctx.Err() != nil {
				return // Context cancelled, exit silently
			}
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("failed to read challenge: %w", err)}
			return
		}

		if challengeMsg.Type != "event" || challengeMsg.Event != "connect.challenge" {
			if challengeMsg.Type == "event" && challengeMsg.Event == "error" {
				events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("server error: %s", string(challengeMsg.Payload))}
				return
			}
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("expected connect.challenge, got: type=%s event=%s", challengeMsg.Type, challengeMsg.Event)}
			return
		}

		// Step 2: Send connect request
		connectReq := OpenClawMessage{
			Type:   "req",
			ID:     "connect-1",
			Method: "connect",
			Params: mustMarshal(ConnectParams{
				MinProtocol: 3,
				MaxProtocol: 3,
				Client: ConnectClient{
					ID:       "cli",
					Version:  "1.0.0",
					Platform: "darwin",
					Mode:     "cli",
				},
				Role:        "operator",
				Scopes:      []string{"operator.read", "operator.write"},
				Caps:        []string{},
				Commands:    []string{},
				Permissions: map[string]any{},
				Auth:        ConnectAuth{Token: token},
			}),
		}

		if err := conn.WriteJSON(connectReq); err != nil {
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("failed to send connect: %w", err)}
			return
		}

		// Step 3: Wait for connect response
		conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
		var connectResp OpenClawMessage
		if err := conn.ReadJSON(&connectResp); err != nil {
			if ctx.Err() != nil {
				return
			}
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("failed to read connect response: %w", err)}
			return
		}

		if connectResp.Type != "res" || connectResp.ID != "connect-1" {
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("unexpected response: type=%s id=%s", connectResp.Type, connectResp.ID)}
			return
		}

		if connectResp.OK != nil && !*connectResp.OK {
			errMsg := "connect failed"
			if connectResp.Error != nil {
				errMsg = connectResp.Error.Message
			}
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf(errMsg)}
			return
		}

		// Step 4: Build the message from the last user message
		var userMessage string
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				userMessage = messages[i].Content
				break
			}
		}

		if userMessage == "" {
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("no user message found")}
			return
		}

		// Step 5: Send chat.send request
		// Use provided session or default to CLI-specific session
		sessionKey := session
		if sessionKey == "" {
			sessionKey = "agent:main:cli"
		}
		chatReq := OpenClawMessage{
			Type:   "req",
			ID:     "chat-1",
			Method: "chat.send",
			Params: mustMarshal(ChatSendParams{
				Message:        userMessage,
				IdempotencyKey: uuid.New().String(),
				SessionKey:     sessionKey,
			}),
		}

		if err := conn.WriteJSON(chatReq); err != nil {
			events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("failed to send request: %w", err)}
			return
		}

		// Step 6: Read streaming responses
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Set read deadline for each read - allows context cancellation to work
				conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
				var msg OpenClawMessage
				if err := conn.ReadJSON(&msg); err != nil {
					if ctx.Err() != nil {
						return // Context cancelled, exit silently
					}
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
						// Check if it's a timeout - just continue if so
						if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
							continue
						}
						events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("read error: %w", err)}
					}
					return
				}

				switch msg.Type {
				case "event":
					switch msg.Event {
					case "agent":
						// Parse agent event
						var payload AgentEventPayload
						if err := json.Unmarshal(msg.Payload, &payload); err != nil {
							continue
						}

						switch payload.Stream {
						case "assistant":
							// Text delta
							if delta, ok := payload.Data["delta"].(string); ok {
								events <- StreamEvent{Type: StreamEventToken, Token: delta}
							}

						case "tool":
							// Tool events
							if phase, ok := payload.Data["phase"].(string); ok {
								switch phase {
								case "start":
									// Tool call starting
									toolName, _ := payload.Data["name"].(string)
									toolID, _ := payload.Data["id"].(string)
									var args map[string]interface{}
									if input, ok := payload.Data["input"].(map[string]interface{}); ok {
										args = input
									}
									events <- StreamEvent{
										Type: StreamEventToolCall,
										ToolCall: &ToolCall{
											ID:        toolID,
											Name:      toolName,
											Arguments: args,
											Status:    ToolStatusRunning,
										},
									}
								case "end":
									// Tool result
									toolID, _ := payload.Data["id"].(string)
									result, _ := payload.Data["result"].(string)
									isError, _ := payload.Data["isError"].(bool)
									events <- StreamEvent{
										Type: StreamEventToolResult,
										ToolResult: &ToolResult{
											ToolCallID: toolID,
											Content:    result,
											IsError:    isError,
										},
									}
								}
							}

						case "lifecycle":
							// Lifecycle events
							if phase, ok := payload.Data["phase"].(string); ok {
								if phase == "end" {
									events <- StreamEvent{Type: StreamEventDone}
									return
								}
							}

						case "error":
							// Error events
							errMsg := "agent error"
							if reason, ok := payload.Data["reason"].(string); ok {
								errMsg = reason
							}
							events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf(errMsg)}
							return
						}

					case "chat":
						// Chat events - check for final state
						var payload ChatEventPayload
						if err := json.Unmarshal(msg.Payload, &payload); err != nil {
							continue
						}
						if payload.State == "final" {
							events <- StreamEvent{Type: StreamEventDone}
							return
						}

					case "tick", "health", "presence":
						// Ignore these events
						continue

					case "error":
						// Global error event
						events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf("server error")}
						return
					}

				case "res":
					// Handle response to our chat request
					if msg.ID == "chat-1" {
						if msg.OK != nil && !*msg.OK {
							errMsg := "request failed"
							if msg.Error != nil {
								errMsg = msg.Error.Message
							}
							events <- StreamEvent{Type: StreamEventError, Error: fmt.Errorf(errMsg)}
							return
						}
						// Request accepted, continue waiting for events
					}
				}
			}
		}()

		wg.Wait()
	}()

	return events
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// BuildEndpoint is kept for backwards compatibility but now returns the WebSocket URL.
func BuildEndpoint(gatewayURL, path string) string {
	// If gatewayURL looks like a WebSocket URL, use it directly
	if strings.HasPrefix(gatewayURL, "ws://") || strings.HasPrefix(gatewayURL, "wss://") {
		return gatewayURL
	}
	// Otherwise, assume it's an agent ID and build the URL
	return BuildGatewayURL(gatewayURL)
}
