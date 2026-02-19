package copilot

import "encoding/json"

// hookConfig represents the .github/hooks/copilot-setup.json structure
type hookConfig struct {
	Version int                    `json:"version"`
	Hooks   map[string][]hookEntry `json:"hooks"`
}

// hookEntry represents a single hook command entry
type hookEntry struct {
	Type       string `json:"type"`
	Bash       string `json:"bash"`
	TimeoutSec int    `json:"timeoutSec,omitempty"`
}

// --- Hook stdin types (from actual Copilot CLI v0.0.411 capture) ---
//
// Common fields: only "timestamp" (unix milliseconds) and "cwd" are present in all hooks.
// "sessionId" and "transcriptPath" are only present in the agentStop hook.

// hookBase contains the only fields present in ALL Copilot hook stdin JSON.
type hookBase struct {
	Timestamp int64  `json:"timestamp"`
	Cwd       string `json:"cwd"`
}

// sessionStartRaw is the stdin JSON for the sessionStart hook.
// Example: {"timestamp":1771465283476,"cwd":"/path","source":"new","initialPrompt":"..."}
type sessionStartRaw struct {
	hookBase

	Source        string `json:"source,omitempty"`
	InitialPrompt string `json:"initialPrompt,omitempty"`
}

// sessionEndRaw is the stdin JSON for the sessionEnd hook.
// Example: {"timestamp":1771465290000,"cwd":"/path","reason":"complete"}
type sessionEndRaw struct {
	hookBase

	Reason string `json:"reason,omitempty"`
}

// userPromptRaw is the stdin JSON for the userPromptSubmitted hook.
// Example: {"timestamp":1771465282293,"cwd":"/path","prompt":"..."}
type userPromptRaw struct {
	hookBase

	Prompt string `json:"prompt,omitempty"`
}

// agentStopRaw is the stdin JSON for the agentStop hook.
// This is the only hook that includes sessionId and transcriptPath.
// Example: {"timestamp":...,"cwd":"/path","sessionId":"uuid","transcriptPath":"/path/events.jsonl","stopReason":"end_turn"}
type agentStopRaw struct {
	hookBase

	SessionID      string `json:"sessionId"`
	TranscriptPath string `json:"transcriptPath"`
	StopReason     string `json:"stopReason,omitempty"`
}

// subagentStopRaw is the stdin JSON for the subagentStop hook.
// Assumed to follow agentStop format (not yet captured in the wild).
type subagentStopRaw struct {
	hookBase

	SessionID      string `json:"sessionId,omitempty"`
	TranscriptPath string `json:"transcriptPath,omitempty"`
}

// toolUseRaw is the stdin JSON for preToolUse/postToolUse hooks.
// Note: toolArgs is a JSON string in preToolUse but an object in postToolUse.
// Example (pre):  {"timestamp":...,"cwd":"/path","toolName":"view","toolArgs":"{\"path\":\"/foo\"}"}
// Example (post): {"timestamp":...,"cwd":"/path","toolName":"view","toolArgs":{"path":"/foo"},"toolResult":{...}}
type toolUseRaw struct {
	hookBase

	ToolName   string          `json:"toolName"`
	ToolArgs   json.RawMessage `json:"toolArgs,omitempty"`
	ToolResult json.RawMessage `json:"toolResult,omitempty"`
}

// --- events.jsonl types (session transcript format) ---

// Event represents a single line in a Copilot events.jsonl transcript.
type Event struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	ID        string          `json:"id"`
	Timestamp string          `json:"timestamp"`
	ParentID  *string         `json:"parentId"`
}

// Event type constants for events.jsonl
//
// Known types not listed here (unused but documented):
//
//	session.end, session.mode_changed, session.model_change,
//	assistant.turn_start, assistant.turn_end, tool.execution_complete
const (
	EventUserMessage      = "user.message"
	EventAssistantMessage = "assistant.message"
	EventToolExecStart    = "tool.execution_start"
)

// userMessageData is the data field for user.message events.
type userMessageData struct {
	Content            string `json:"content"`
	TransformedContent string `json:"transformedContent"`
}

// assistantMessageData is the data field for assistant.message events.
type assistantMessageData struct {
	MessageID    string        `json:"messageId"`
	Content      string        `json:"content"`
	ToolRequests []toolRequest `json:"toolRequests"`
}

// toolRequest represents a tool call request within an assistant.message event.
type toolRequest struct {
	ToolCallID string         `json:"toolCallId"`
	Name       string         `json:"name"`
	Arguments  map[string]any `json:"arguments"`
	Type       string         `json:"type"`
}

// toolExecStartData is the data field for tool.execution_start events.
type toolExecStartData struct {
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Arguments  map[string]any `json:"arguments"`
}

// Tool names used in GitHub Copilot that modify files
const (
	ToolEdit   = "edit"
	ToolCreate = "create"
)

// fileModificationTools lists tools that create or modify files in GitHub Copilot
var fileModificationTools = []string{
	ToolEdit,
	ToolCreate,
}
