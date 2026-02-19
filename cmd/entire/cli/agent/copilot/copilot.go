// Package copilot implements the Agent interface for GitHub Copilot CLI.
package copilot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/entireio/cli/cmd/entire/cli/agent"
	"github.com/entireio/cli/cmd/entire/cli/paths"
)

//nolint:gochecknoinits // Agent self-registration is the intended pattern
func init() {
	agent.Register(agent.AgentNameCopilot, NewCopilotAgent)
}

// CopilotAgent implements the Agent interface for GitHub Copilot CLI.
//
//nolint:revive // CopilotAgent is clearer than Agent in this context
type CopilotAgent struct{}

// NewCopilotAgent creates a new CopilotAgent.
func NewCopilotAgent() agent.Agent {
	return &CopilotAgent{}
}

// Name returns the agent registry key.
func (c *CopilotAgent) Name() agent.AgentName {
	return agent.AgentNameCopilot
}

// Type returns the agent type identifier.
func (c *CopilotAgent) Type() agent.AgentType {
	return agent.AgentTypeCopilot
}

// Description returns a human-readable description.
func (c *CopilotAgent) Description() string {
	return "GitHub Copilot - GitHub's AI coding assistant"
}

// IsPreview returns true as the Copilot integration is in preview.
func (c *CopilotAgent) IsPreview() bool { return true }

// DetectPresence checks if GitHub Copilot is configured in the repository.
func (c *CopilotAgent) DetectPresence() (bool, error) {
	repoRoot, err := paths.RepoRoot()
	if err != nil {
		repoRoot = "."
	}

	// Check if our hooks are installed
	if c.AreHooksInstalled() {
		return true, nil
	}

	// Check for .github/hooks directory with copilot config
	configPath := filepath.Join(repoRoot, ".github", "hooks", CopilotConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return true, nil
	}

	return false, nil
}

// GetHookConfigPath returns the path to Copilot's hook config file.
func (c *CopilotAgent) GetHookConfigPath() string {
	return filepath.Join(".github", "hooks", CopilotConfigFileName)
}

// SupportsHooks returns true as GitHub Copilot supports lifecycle hooks.
func (c *CopilotAgent) SupportsHooks() bool {
	return true
}

// ParseHookInput parses Copilot hook input from stdin.
func (c *CopilotAgent) ParseHookInput(hookType agent.HookType, reader io.Reader) (*agent.HookInput, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("empty input")
	}

	input := &agent.HookInput{
		HookType:  hookType,
		Timestamp: time.Now(),
		RawData:   make(map[string]any),
	}

	switch hookType {
	case agent.HookSessionStart:
		var raw sessionStartRaw
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse hook input: %w", err)
		}
		input.RawData["cwd"] = raw.Cwd
		input.RawData["source"] = raw.Source

	case agent.HookSessionEnd:
		var raw sessionEndRaw
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse hook input: %w", err)
		}
		input.RawData["cwd"] = raw.Cwd
		input.RawData["reason"] = raw.Reason

	case agent.HookStop:
		// agentStop is the only hook with sessionId and transcriptPath
		var raw agentStopRaw
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse hook input: %w", err)
		}
		input.SessionID = raw.SessionID
		input.SessionRef = raw.TranscriptPath
		input.RawData["cwd"] = raw.Cwd
		input.RawData["stopReason"] = raw.StopReason

	case agent.HookUserPromptSubmit:
		var raw userPromptRaw
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse hook input: %w", err)
		}
		input.RawData["cwd"] = raw.Cwd
		if raw.Prompt != "" {
			input.UserPrompt = raw.Prompt
			input.RawData["prompt"] = raw.Prompt
		}

	case agent.HookPreToolUse, agent.HookPostToolUse:
		var raw toolUseRaw
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse tool hook input: %w", err)
		}
		input.ToolName = raw.ToolName
		input.ToolInput = raw.ToolArgs
		if hookType == agent.HookPostToolUse {
			input.ToolResponse = raw.ToolResult
		}
		input.RawData["cwd"] = raw.Cwd
	}

	return input, nil
}

// GetSessionID extracts the session ID from hook input.
func (c *CopilotAgent) GetSessionID(input *agent.HookInput) string {
	return input.SessionID
}

// ProtectedDirs returns directories that Copilot uses.
// .github is a shared directory (not solely ours), so we don't protect it.
func (c *CopilotAgent) ProtectedDirs() []string { return []string{} }

// GetSessionDir returns the directory where Copilot stores session transcripts.
func (c *CopilotAgent) GetSessionDir(_ string) (string, error) {
	if override := os.Getenv("ENTIRE_TEST_COPILOT_SESSION_DIR"); override != "" {
		return override, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".copilot", "session-state"), nil
}

// ResolveSessionFile returns the path to a Copilot session file.
func (c *CopilotAgent) ResolveSessionFile(sessionDir, agentSessionID string) string {
	return filepath.Join(sessionDir, agentSessionID, "events.jsonl")
}

// ReadSession reads a session from Copilot's storage.
func (c *CopilotAgent) ReadSession(input *agent.HookInput) (*agent.AgentSession, error) {
	if input.SessionRef == "" {
		return nil, errors.New("session reference (transcript path) is required")
	}

	data, err := os.ReadFile(input.SessionRef)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}

	return &agent.AgentSession{
		SessionID:  input.SessionID,
		AgentName:  c.Name(),
		SessionRef: input.SessionRef,
		StartTime:  time.Now(),
		NativeData: data,
	}, nil
}

// WriteSession writes a session to Copilot's storage.
func (c *CopilotAgent) WriteSession(session *agent.AgentSession) error {
	if session == nil {
		return errors.New("session is nil")
	}

	if session.AgentName != "" && session.AgentName != c.Name() {
		return fmt.Errorf("session belongs to agent %q, not %q", session.AgentName, c.Name())
	}

	if session.SessionRef == "" {
		return errors.New("session reference (transcript path) is required")
	}

	if len(session.NativeData) == 0 {
		return errors.New("session has no native data to write")
	}

	if err := os.WriteFile(session.SessionRef, session.NativeData, 0o600); err != nil {
		return fmt.Errorf("failed to write transcript: %w", err)
	}

	return nil
}

// FormatResumeCommand returns the command to resume a Copilot session.
func (c *CopilotAgent) FormatResumeCommand(_ string) string {
	return "copilot --resume"
}
