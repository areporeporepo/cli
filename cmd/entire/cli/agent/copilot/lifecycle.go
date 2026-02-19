package copilot

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/entireio/cli/cmd/entire/cli/agent"
)

// Compile-time interface assertions
var _ agent.TranscriptAnalyzer = (*CopilotAgent)(nil)

// HookNames returns the hook verbs Copilot supports.
func (c *CopilotAgent) HookNames() []string {
	return []string{
		HookNameSessionStart,
		HookNameSessionEnd,
		HookNameUserPromptSubmitted,
		HookNameAgentStop,
		HookNameSubagentStop,
		HookNamePreToolUse,
		HookNamePostToolUse,
		HookNameErrorOccurred,
	}
}

// ParseHookEvent translates a Copilot hook into a normalized lifecycle Event.
// Returns nil if the hook has no lifecycle significance (e.g., pass-through hooks).
//
// Note: In Copilot CLI, only agentStop provides sessionId and transcriptPath.
// Other hooks (sessionStart, userPromptSubmitted, sessionEnd) do not include
// these fields, so the resulting Event will have empty SessionID/SessionRef.
func (c *CopilotAgent) ParseHookEvent(hookName string, stdin io.Reader) (*agent.Event, error) {
	switch hookName {
	case HookNameSessionStart:
		return c.parseSessionStart(stdin)
	case HookNameUserPromptSubmitted:
		return c.parseTurnStart(stdin)
	case HookNameAgentStop:
		return c.parseTurnEnd(stdin)
	case HookNameSessionEnd:
		return c.parseSessionEnd(stdin)
	case HookNameSubagentStop:
		return c.parseSubagentEnd(stdin)
	case HookNamePreToolUse, HookNamePostToolUse, HookNameErrorOccurred:
		return nil, nil //nolint:nilnil // nil event = no lifecycle action
	default:
		return nil, nil //nolint:nilnil // Unknown hooks have no lifecycle action
	}
}

// ReadTranscript reads the raw JSONL transcript bytes for a session.
func (c *CopilotAgent) ReadTranscript(sessionRef string) ([]byte, error) {
	data, err := os.ReadFile(sessionRef) //nolint:gosec // Path comes from agent hook input
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}
	return data, nil
}

// ChunkTranscript splits a JSONL transcript into chunks.
func (c *CopilotAgent) ChunkTranscript(content []byte, maxSize int) ([][]byte, error) {
	chunks, err := agent.ChunkJSONL(content, maxSize)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk JSONL: %w", err)
	}
	return chunks, nil
}

// ReassembleTranscript combines JSONL chunks back into a single transcript.
func (c *CopilotAgent) ReassembleTranscript(chunks [][]byte) ([]byte, error) {
	return agent.ReassembleJSONL(chunks), nil
}

// ExtractPrompts extracts user prompts from the transcript starting at the given line offset.
func (c *CopilotAgent) ExtractPrompts(sessionRef string, fromOffset int) ([]string, error) {
	data, err := os.ReadFile(sessionRef) //nolint:gosec // Path comes from agent hook input
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}

	return extractPromptsFromData(data, fromOffset)
}

// ExtractSummary extracts the last assistant message as a session summary.
func (c *CopilotAgent) ExtractSummary(sessionRef string) (string, error) {
	data, err := os.ReadFile(sessionRef) //nolint:gosec // Path comes from agent hook input
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}
	return extractLastAssistantMessage(data), nil
}

// GetTranscriptPosition returns the current line count of the JSONL transcript.
func (c *CopilotAgent) GetTranscriptPosition(path string) (int, error) {
	if path == "" {
		return 0, nil
	}

	data, err := os.ReadFile(path) //nolint:gosec // Reading from controlled transcript path
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read transcript: %w", err)
	}

	return countLines(data), nil
}

// ExtractModifiedFilesFromOffset extracts files modified since a given line offset.
func (c *CopilotAgent) ExtractModifiedFilesFromOffset(path string, startOffset int) (files []string, currentPosition int, err error) {
	if path == "" {
		return nil, 0, nil
	}

	data, readErr := os.ReadFile(path) //nolint:gosec // Reading from controlled transcript path
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to read transcript: %w", readErr)
	}

	if len(data) == 0 {
		return nil, 0, nil
	}

	totalLines := countLines(data)
	extractedFiles := extractModifiedFilesFromLines(data, startOffset)

	return extractedFiles, totalLines, nil
}

// --- Internal hook parsing functions ---

// parseSessionStart handles the sessionStart hook.
// Note: Copilot's sessionStart stdin does NOT include sessionId or transcriptPath.
func (c *CopilotAgent) parseSessionStart(stdin io.Reader) (*agent.Event, error) {
	_, err := agent.ReadAndParseHookInput[sessionStartRaw](stdin)
	if err != nil {
		return nil, err
	}
	return &agent.Event{
		Type:      agent.SessionStart,
		Timestamp: time.Now(),
	}, nil
}

// parseTurnStart handles the userPromptSubmitted hook.
// Note: Copilot's userPromptSubmitted stdin does NOT include sessionId or transcriptPath.
func (c *CopilotAgent) parseTurnStart(stdin io.Reader) (*agent.Event, error) {
	raw, err := agent.ReadAndParseHookInput[userPromptRaw](stdin)
	if err != nil {
		return nil, err
	}
	return &agent.Event{
		Type:      agent.TurnStart,
		Prompt:    raw.Prompt,
		Timestamp: time.Now(),
	}, nil
}

// parseTurnEnd handles the agentStop hook.
// This is the only hook that provides sessionId and transcriptPath.
func (c *CopilotAgent) parseTurnEnd(stdin io.Reader) (*agent.Event, error) {
	raw, err := agent.ReadAndParseHookInput[agentStopRaw](stdin)
	if err != nil {
		return nil, err
	}
	return &agent.Event{
		Type:       agent.TurnEnd,
		SessionID:  raw.SessionID,
		SessionRef: raw.TranscriptPath,
		Timestamp:  time.Now(),
	}, nil
}

// parseSessionEnd handles the sessionEnd hook.
// Note: Copilot's sessionEnd stdin does NOT include sessionId or transcriptPath.
func (c *CopilotAgent) parseSessionEnd(stdin io.Reader) (*agent.Event, error) {
	_, err := agent.ReadAndParseHookInput[sessionEndRaw](stdin)
	if err != nil {
		return nil, err
	}
	return &agent.Event{
		Type:      agent.SessionEnd,
		Timestamp: time.Now(),
	}, nil
}

// parseSubagentEnd handles the subagentStop hook.
func (c *CopilotAgent) parseSubagentEnd(stdin io.Reader) (*agent.Event, error) {
	raw, err := agent.ReadAndParseHookInput[subagentStopRaw](stdin)
	if err != nil {
		return nil, err
	}
	return &agent.Event{
		Type:       agent.SubagentEnd,
		SessionID:  raw.SessionID,
		SessionRef: raw.TranscriptPath,
		Timestamp:  time.Now(),
	}, nil
}
