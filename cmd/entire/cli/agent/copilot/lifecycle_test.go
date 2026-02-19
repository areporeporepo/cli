package copilot

import (
	"errors"
	"strings"
	"testing"

	"github.com/entireio/cli/cmd/entire/cli/agent"
)

const (
	testSessionID      = "abc-123"
	testTranscriptPath = "/tmp/events.jsonl"
)

func TestParseHookEvent_SessionStart_WithoutSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Copilot sessionStart without sessionId (current behavior)
	input := `{"timestamp":1771465283476,"cwd":"/Users/test/project","source":"new","initialPrompt":"hello"}`

	_, err := ag.ParseHookEvent(HookNameSessionStart, strings.NewReader(input))

	if !errors.Is(err, ErrMissingSessionID) {
		t.Errorf("expected ErrMissingSessionID, got %v", err)
	}
}

func TestParseHookEvent_SessionStart_WithSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// sessionStart with sessionId (future Copilot behavior)
	// See: https://github.com/github/copilot-cli/issues/1425
	input := `{"timestamp":1771465283476,"cwd":"/Users/test/project","source":"new","initialPrompt":"hello","sessionId":"` + testSessionID + `","transcriptPath":"` + testTranscriptPath + `"}`

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
	if event.SessionID != testSessionID {
		t.Errorf("expected session_id %q, got %q", testSessionID, event.SessionID)
	}
	if event.SessionRef != testTranscriptPath {
		t.Errorf("expected session_ref %q, got %q", testTranscriptPath, event.SessionRef)
	}
	if event.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestParseHookEvent_TurnStart_WithoutSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Copilot userPromptSubmitted without sessionId (current behavior)
	input := `{"timestamp":1771465282293,"cwd":"/Users/test/project","prompt":"Hello Copilot"}`

	_, err := ag.ParseHookEvent(HookNameUserPromptSubmitted, strings.NewReader(input))

	if !errors.Is(err, ErrMissingSessionID) {
		t.Errorf("expected ErrMissingSessionID, got %v", err)
	}
}

func TestParseHookEvent_TurnStart_WithSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// userPromptSubmitted with sessionId (future Copilot behavior)
	// See: https://github.com/github/copilot-cli/issues/1425
	input := `{"timestamp":1771465282293,"cwd":"/Users/test/project","prompt":"Hello Copilot","sessionId":"` + testSessionID + `","transcriptPath":"` + testTranscriptPath + `"}`

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
	if event.SessionID != testSessionID {
		t.Errorf("expected session_id %q, got %q", testSessionID, event.SessionID)
	}
	if event.Prompt != "Hello Copilot" {
		t.Errorf("expected prompt %q, got %q", "Hello Copilot", event.Prompt)
	}
}

func TestParseHookEvent_TurnEnd(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
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

func TestParseHookEvent_SessionEnd_WithoutSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Copilot sessionEnd without sessionId (current behavior)
	input := `{"timestamp":1771465290000,"cwd":"/Users/test/project","reason":"complete"}`

	_, err := ag.ParseHookEvent(HookNameSessionEnd, strings.NewReader(input))

	if !errors.Is(err, ErrMissingSessionID) {
		t.Errorf("expected ErrMissingSessionID, got %v", err)
	}
}

func TestParseHookEvent_SessionEnd_WithSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// sessionEnd with sessionId (future Copilot behavior)
	// See: https://github.com/github/copilot-cli/issues/1425
	input := `{"timestamp":1771465290000,"cwd":"/Users/test/project","reason":"complete","sessionId":"` + testSessionID + `","transcriptPath":"` + testTranscriptPath + `"}`

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
	if event.SessionID != testSessionID {
		t.Errorf("expected session_id %q, got %q", testSessionID, event.SessionID)
	}
	if event.SessionRef != testTranscriptPath {
		t.Errorf("expected session_ref %q, got %q", testTranscriptPath, event.SessionRef)
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

func TestParseHookEvent_AllLifecycleHooks_WithSessionID(t *testing.T) {
	t.Parallel()

	// When sessionId is present, all lifecycle hooks should produce events.
	// See: https://github.com/github/copilot-cli/issues/1425
	testCases := []struct {
		hookName      string
		expectedType  agent.EventType
		expectNil     bool
		inputTemplate string
	}{
		{
			hookName:      HookNameSessionStart,
			expectedType:  agent.SessionStart,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","source":"new","sessionId":"s1","transcriptPath":"/t"}`,
		},
		{
			hookName:      HookNameUserPromptSubmitted,
			expectedType:  agent.TurnStart,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","prompt":"hi","sessionId":"s2","transcriptPath":"/t"}`,
		},
		{
			hookName:      HookNameAgentStop,
			expectedType:  agent.TurnEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","sessionId":"s3","transcriptPath":"/t","stopReason":"end_turn"}`,
		},
		{
			hookName:      HookNameSessionEnd,
			expectedType:  agent.SessionEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","reason":"complete","sessionId":"s4","transcriptPath":"/t"}`,
		},
		{
			hookName:      HookNameSubagentStop,
			expectedType:  agent.SubagentEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","sessionId":"s5","transcriptPath":"/t"}`,
		},
		{
			hookName:      HookNamePreToolUse,
			expectNil:     true,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","sessionId":"s6"}`,
		},
		{
			hookName:      HookNamePostToolUse,
			expectNil:     true,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","sessionId":"s6"}`,
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

func TestParseHookEvent_WithoutSessionID_ReturnsError(t *testing.T) {
	t.Parallel()

	// When sessionId is absent (current Copilot behavior), lifecycle hooks
	// should return ErrMissingSessionID.
	// See: https://github.com/github/copilot-cli/issues/1425
	hooks := []struct {
		hookName      string
		inputTemplate string
	}{
		{
			hookName:      HookNameSessionStart,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","source":"new"}`,
		},
		{
			hookName:      HookNameUserPromptSubmitted,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","prompt":"hi"}`,
		},
		{
			hookName:      HookNameSessionEnd,
			inputTemplate: `{"timestamp":1000,"cwd":"/tmp","reason":"complete"}`,
		},
	}

	for _, tc := range hooks {
		t.Run(tc.hookName, func(t *testing.T) {
			t.Parallel()

			ag := &CopilotAgent{}
			_, err := ag.ParseHookEvent(tc.hookName, strings.NewReader(tc.inputTemplate))

			if !errors.Is(err, ErrMissingSessionID) {
				t.Errorf("expected ErrMissingSessionID, got %v", err)
			}
		})
	}
}
