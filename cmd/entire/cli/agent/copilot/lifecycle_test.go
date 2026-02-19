package copilot

import (
	"strings"
	"testing"

	"github.com/entireio/cli/cmd/entire/cli/agent"
)

func TestParseHookEvent_SessionStart(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot sessionStart stdin format (no sessionId/transcriptPath)
	input := `{"timestamp":1771465283476,"cwd":"/Users/test/project","source":"new","initialPrompt":"hello"}`

	event, err := ag.ParseHookEvent(HookNameSessionStart, strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != agent.SessionStart {
		t.Errorf("expected event type %v, got %v", agent.SessionStart, event.Type)
	}
	// sessionStart does NOT provide sessionId or transcriptPath
	if event.SessionID != "" {
		t.Errorf("expected empty session_id, got %q", event.SessionID)
	}
	if event.SessionRef != "" {
		t.Errorf("expected empty session_ref, got %q", event.SessionRef)
	}
	if event.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParseHookEvent_TurnStart(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot userPromptSubmitted stdin format
	input := `{"timestamp":1771465282293,"cwd":"/Users/test/project","prompt":"Hello Copilot"}`

	event, err := ag.ParseHookEvent(HookNameUserPromptSubmitted, strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != agent.TurnStart {
		t.Errorf("expected event type %v, got %v", agent.TurnStart, event.Type)
	}
	// userPromptSubmitted does NOT provide sessionId
	if event.SessionID != "" {
		t.Errorf("expected empty session_id, got %q", event.SessionID)
	}
	if event.Prompt != "Hello Copilot" {
		t.Errorf("expected prompt 'Hello Copilot', got %q", event.Prompt)
	}
}

func TestParseHookEvent_TurnEnd(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot agentStop stdin format - the ONLY hook with sessionId and transcriptPath
	input := `{
		"timestamp": 1771465289990,
		"cwd": "/Users/test/project",
		"sessionId": "733bfbbd-bd9b-4750-a55b-3c8cc42629de",
		"transcriptPath": "/Users/test/.copilot/session-state/733bfbbd/events.jsonl",
		"stopReason": "end_turn"
	}`

	event, err := ag.ParseHookEvent(HookNameAgentStop, strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != agent.TurnEnd {
		t.Errorf("expected event type %v, got %v", agent.TurnEnd, event.Type)
	}
	if event.SessionID != "733bfbbd-bd9b-4750-a55b-3c8cc42629de" {
		t.Errorf("expected session_id, got %q", event.SessionID)
	}
	if event.SessionRef != "/Users/test/.copilot/session-state/733bfbbd/events.jsonl" {
		t.Errorf("expected session_ref, got %q", event.SessionRef)
	}
}

func TestParseHookEvent_SessionEnd(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot sessionEnd stdin format
	input := `{"timestamp":1771465290000,"cwd":"/Users/test/project","reason":"complete"}`

	event, err := ag.ParseHookEvent(HookNameSessionEnd, strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != agent.SessionEnd {
		t.Errorf("expected event type %v, got %v", agent.SessionEnd, event.Type)
	}
	// sessionEnd does NOT provide sessionId
	if event.SessionID != "" {
		t.Errorf("expected empty session_id, got %q", event.SessionID)
	}
}

func TestParseHookEvent_SubagentEnd(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	input := `{"timestamp":1771465290000,"cwd":"/Users/test/project","sessionId":"sub-123","transcriptPath":"/tmp/sub.jsonl"}`

	event, err := ag.ParseHookEvent(HookNameSubagentStop, strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if event.Type != agent.SubagentEnd {
		t.Errorf("expected event type %v, got %v", agent.SubagentEnd, event.Type)
	}
}

func TestParseHookEvent_PassThroughHooks_ReturnNil(t *testing.T) {
	t.Parallel()

	passThroughHooks := []string{
		HookNamePreToolUse,
		HookNamePostToolUse,
		HookNameErrorOccurred,
	}

	ag := &CopilotAgent{}
	input := `{"timestamp":1771465282293,"cwd":"/Users/test/project"}`

	for _, hookName := range passThroughHooks {
		t.Run(hookName, func(t *testing.T) {
			t.Parallel()

			event, err := ag.ParseHookEvent(hookName, strings.NewReader(input))

			if err != nil {
				t.Fatalf("unexpected error for %s: %v", hookName, err)
			}
			if event != nil {
				t.Errorf("expected nil event for %s, got %+v", hookName, event)
			}
		})
	}
}

func TestParseHookEvent_UnknownHook_ReturnsNil(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	input := `{"timestamp":1771465282293,"cwd":"/Users/test/project"}`

	event, err := ag.ParseHookEvent("unknown-hook-name", strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event != nil {
		t.Errorf("expected nil event for unknown hook, got %+v", event)
	}
}

func TestParseHookEvent_EmptyInput(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	_, err := ag.ParseHookEvent(HookNameSessionStart, strings.NewReader(""))

	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), "empty hook input") {
		t.Errorf("expected 'empty hook input' error, got: %v", err)
	}
}

func TestParseHookEvent_MalformedJSON(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	input := `{"timestamp": INVALID}`

	_, err := ag.ParseHookEvent(HookNameSessionStart, strings.NewReader(input))

	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse hook input") {
		t.Errorf("expected 'failed to parse hook input' error, got: %v", err)
	}
}

func TestParseHookEvent_AllLifecycleHooks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		hookName      string
		expectedType  agent.EventType
		expectNil     bool
		inputTemplate string
	}{
		{
			hookName:      HookNameSessionStart,
			expectedType:  agent.SessionStart,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","source":"new"}`,
		},
		{
			hookName:      HookNameUserPromptSubmitted,
			expectedType:  agent.TurnStart,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","prompt":"hi"}`,
		},
		{
			hookName:      HookNameAgentStop,
			expectedType:  agent.TurnEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","sessionId":"s3","transcriptPath":"/t","stopReason":"end_turn"}`,
		},
		{
			hookName:      HookNameSessionEnd,
			expectedType:  agent.SessionEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","reason":"complete"}`,
		},
		{
			hookName:      HookNameSubagentStop,
			expectedType:  agent.SubagentEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp"}`,
		},
		{
			hookName:      HookNamePreToolUse,
			expectNil:     true,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp"}`,
		},
		{
			hookName:      HookNamePostToolUse,
			expectNil:     true,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp"}`,
		},
		{
			hookName:      HookNameErrorOccurred,
			expectNil:     true,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.hookName, func(t *testing.T) {
			t.Parallel()

			ag := &CopilotAgent{}
			event, err := ag.ParseHookEvent(tc.hookName, strings.NewReader(tc.inputTemplate))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectNil {
				if event != nil {
					t.Errorf("expected nil event, got %+v", event)
				}
				return
			}

			if event == nil {
				t.Fatal("expected event, got nil")
			}
			if event.Type != tc.expectedType {
				t.Errorf("expected event type %v, got %v", tc.expectedType, event.Type)
			}
		})
	}
}
