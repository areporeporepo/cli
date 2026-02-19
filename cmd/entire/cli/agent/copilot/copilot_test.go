package copilot

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/entireio/cli/cmd/entire/cli/agent"
)

func TestNewCopilotAgent(t *testing.T) {
	t.Parallel()

	ag := NewCopilotAgent()
	if ag == nil {
		t.Fatal("NewCopilotAgent() returned nil")
	}

	cp, ok := ag.(*CopilotAgent)
	if !ok {
		t.Fatal("NewCopilotAgent() didn't return *CopilotAgent")
	}
	if cp == nil {
		t.Fatal("NewCopilotAgent() returned nil agent")
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	if name := ag.Name(); name != agent.AgentNameCopilot {
		t.Errorf("Name() = %q, want %q", name, agent.AgentNameCopilot)
	}
}

func TestType(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	if tp := ag.Type(); tp != agent.AgentTypeCopilot {
		t.Errorf("Type() = %q, want %q", tp, agent.AgentTypeCopilot)
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	desc := ag.Description()
	if desc == "" {
		t.Error("Description() returned empty string")
	}
}

func TestIsPreview(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	if !ag.IsPreview() {
		t.Error("IsPreview() = false, want true")
	}
}

func TestDetectPresence_NoConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	ag := &CopilotAgent{}
	present, err := ag.DetectPresence()
	if err != nil {
		t.Fatalf("DetectPresence() error = %v", err)
	}
	if present {
		t.Error("DetectPresence() = true, want false")
	}
}

func TestDetectPresence_WithConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	// Create .github/hooks/copilot-setup.json
	hooksDir := filepath.Join(tempDir, ".github", "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("failed to create hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, CopilotConfigFileName), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	ag := &CopilotAgent{}
	present, err := ag.DetectPresence()
	if err != nil {
		t.Fatalf("DetectPresence() error = %v", err)
	}
	if !present {
		t.Error("DetectPresence() = false, want true")
	}
}

func TestGetHookConfigPath(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	path := ag.GetHookConfigPath()
	if path != filepath.Join(".github", "hooks", CopilotConfigFileName) {
		t.Errorf("GetHookConfigPath() = %q, want .github/hooks/%s", path, CopilotConfigFileName)
	}
}

func TestSupportsHooks(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	if !ag.SupportsHooks() {
		t.Error("SupportsHooks() = false, want true")
	}
}

func TestProtectedDirs(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	dirs := ag.ProtectedDirs()
	if len(dirs) != 0 {
		t.Errorf("ProtectedDirs() = %v, want empty", dirs)
	}
}

func TestGetSessionID(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	input := &agent.HookInput{SessionID: "test-session-456"}

	id := ag.GetSessionID(input)
	if id != "test-session-456" {
		t.Errorf("GetSessionID() = %q, want test-session-456", id)
	}
}

func TestGetSessionDir(t *testing.T) {
	ag := &CopilotAgent{}

	t.Setenv("ENTIRE_TEST_COPILOT_SESSION_DIR", "/test/override")

	dir, err := ag.GetSessionDir("/some/repo")
	if err != nil {
		t.Fatalf("GetSessionDir() error = %v", err)
	}
	if dir != "/test/override" {
		t.Errorf("GetSessionDir() = %q, want /test/override", dir)
	}
}

func TestGetSessionDir_DefaultPath(t *testing.T) {
	ag := &CopilotAgent{}

	t.Setenv("ENTIRE_TEST_COPILOT_SESSION_DIR", "")

	dir, err := ag.GetSessionDir("/some/repo")
	if err != nil {
		t.Fatalf("GetSessionDir() error = %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("GetSessionDir() should return absolute path, got %q", dir)
	}
}

func TestResolveSessionFile(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	result := ag.ResolveSessionFile("/tmp/sessions", "abc123")
	expected := filepath.Join("/tmp/sessions", "abc123", "events.jsonl")
	if result != expected {
		t.Errorf("ResolveSessionFile() = %q, want %q", result, expected)
	}
}

func TestFormatResumeCommand(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	cmd := ag.FormatResumeCommand("abc123")
	if cmd != "copilot --resume" {
		t.Errorf("FormatResumeCommand() = %q, want %q", cmd, "copilot --resume")
	}
}

func TestParseHookInput_SessionStart(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot sessionStart stdin format (no sessionId/transcriptPath)
	input := `{"timestamp":1771465283476,"cwd":"/project","source":"new","initialPrompt":"hello"}`

	hookInput, err := ag.ParseHookInput(agent.HookSessionStart, bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}

	// sessionStart does NOT provide sessionId or transcriptPath
	if hookInput.SessionID != "" {
		t.Errorf("SessionID = %q, want empty (sessionStart has no sessionId)", hookInput.SessionID)
	}
	if hookInput.SessionRef != "" {
		t.Errorf("SessionRef = %q, want empty (sessionStart has no transcriptPath)", hookInput.SessionRef)
	}
	if hookInput.RawData["cwd"] != "/project" {
		t.Errorf("RawData[cwd] = %q, want /project", hookInput.RawData["cwd"])
	}
	if hookInput.RawData["source"] != "new" {
		t.Errorf("RawData[source] = %q, want new", hookInput.RawData["source"])
	}
}

func TestParseHookInput_UserPrompt(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot userPromptSubmitted stdin format (no sessionId)
	input := `{"timestamp":1771465282293,"cwd":"/project","prompt":"Fix the bug"}`

	hookInput, err := ag.ParseHookInput(agent.HookUserPromptSubmit, bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}

	if hookInput.UserPrompt != "Fix the bug" {
		t.Errorf("UserPrompt = %q, want 'Fix the bug'", hookInput.UserPrompt)
	}
	// userPromptSubmitted does NOT provide sessionId
	if hookInput.SessionID != "" {
		t.Errorf("SessionID = %q, want empty", hookInput.SessionID)
	}
}

func TestParseHookInput_ToolUse(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot preToolUse stdin format
	input := `{"timestamp":1771465284000,"cwd":"/project","toolName":"edit","toolArgs":{"file_path":"test.go"}}`

	hookInput, err := ag.ParseHookInput(agent.HookPreToolUse, bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}

	if hookInput.ToolName != "edit" {
		t.Errorf("ToolName = %q, want edit", hookInput.ToolName)
	}
	if hookInput.ToolInput == nil {
		t.Error("ToolInput is nil")
	}
}

func TestParseHookInput_AgentStop(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// agentStop is the ONLY hook with sessionId and transcriptPath
	input := `{"timestamp":1771465289990,"cwd":"/project","sessionId":"733bfbbd-bd9b-4750-a55b-3c8cc42629de","transcriptPath":"/home/user/.copilot/session-state/733bfbbd/events.jsonl","stopReason":"end_turn"}`

	hookInput, err := ag.ParseHookInput(agent.HookStop, bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}

	if hookInput.SessionID != "733bfbbd-bd9b-4750-a55b-3c8cc42629de" {
		t.Errorf("SessionID = %q, want 733bfbbd-bd9b-4750-a55b-3c8cc42629de", hookInput.SessionID)
	}
	if hookInput.SessionRef != "/home/user/.copilot/session-state/733bfbbd/events.jsonl" {
		t.Errorf("SessionRef = %q, want transcript path", hookInput.SessionRef)
	}
	if hookInput.RawData["stopReason"] != "end_turn" {
		t.Errorf("RawData[stopReason] = %q, want end_turn", hookInput.RawData["stopReason"])
	}
}

func TestParseHookInput_SessionEnd(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	// Actual Copilot sessionEnd stdin format (no sessionId)
	input := `{"timestamp":1771465290000,"cwd":"/project","reason":"complete"}`

	hookInput, err := ag.ParseHookInput(agent.HookSessionEnd, bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseHookInput() error = %v", err)
	}

	// sessionEnd does NOT provide sessionId
	if hookInput.SessionID != "" {
		t.Errorf("SessionID = %q, want empty", hookInput.SessionID)
	}
	if hookInput.RawData["reason"] != "complete" {
		t.Errorf("RawData[reason] = %q, want complete", hookInput.RawData["reason"])
	}
}

func TestParseHookInput_Empty(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	_, err := ag.ParseHookInput(agent.HookSessionStart, bytes.NewReader([]byte("")))
	if err == nil {
		t.Error("ParseHookInput() should error on empty input")
	}
}

func TestParseHookInput_InvalidJSON(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	_, err := ag.ParseHookInput(agent.HookSessionStart, bytes.NewReader([]byte("{not json}")))
	if err == nil {
		t.Error("ParseHookInput() should error on invalid JSON")
	}
}

func TestReadSession(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	transcriptPath := filepath.Join(tempDir, "transcript.jsonl")
	content := `{"role":"user","content":"hello"}`
	if err := os.WriteFile(transcriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write transcript: %v", err)
	}

	ag := &CopilotAgent{}
	input := &agent.HookInput{
		SessionID:  "test-session",
		SessionRef: transcriptPath,
	}

	session, err := ag.ReadSession(input)
	if err != nil {
		t.Fatalf("ReadSession() error = %v", err)
	}

	if session.SessionID != "test-session" {
		t.Errorf("SessionID = %q, want test-session", session.SessionID)
	}
	if session.AgentName != agent.AgentNameCopilot {
		t.Errorf("AgentName = %q, want %q", session.AgentName, agent.AgentNameCopilot)
	}
	if len(session.NativeData) == 0 {
		t.Error("NativeData is empty")
	}
}

func TestReadSession_NoSessionRef(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	input := &agent.HookInput{SessionID: "test-session"}

	_, err := ag.ReadSession(input)
	if err == nil {
		t.Error("ReadSession() should error when SessionRef is empty")
	}
}

func TestWriteSession(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	transcriptPath := filepath.Join(tempDir, "transcript.jsonl")

	ag := &CopilotAgent{}
	session := &agent.AgentSession{
		SessionID:  "test-session",
		AgentName:  agent.AgentNameCopilot,
		SessionRef: transcriptPath,
		NativeData: []byte(`{"role":"user","content":"hello"}`),
	}

	if err := ag.WriteSession(session); err != nil {
		t.Fatalf("WriteSession() error = %v", err)
	}

	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		t.Fatalf("failed to read transcript: %v", err)
	}
	if string(data) != `{"role":"user","content":"hello"}` {
		t.Errorf("unexpected content: %s", data)
	}
}

func TestWriteSession_Nil(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	if err := ag.WriteSession(nil); err == nil {
		t.Error("WriteSession(nil) should error")
	}
}

func TestWriteSession_WrongAgent(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	session := &agent.AgentSession{
		AgentName:  "claude-code",
		SessionRef: "/path/to/file",
		NativeData: []byte("{}"),
	}

	if err := ag.WriteSession(session); err == nil {
		t.Error("WriteSession() should error for wrong agent")
	}
}

func TestWriteSession_NoSessionRef(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	session := &agent.AgentSession{
		AgentName:  agent.AgentNameCopilot,
		NativeData: []byte("{}"),
	}

	if err := ag.WriteSession(session); err == nil {
		t.Error("WriteSession() should error when SessionRef is empty")
	}
}

func TestWriteSession_NoNativeData(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	session := &agent.AgentSession{
		AgentName:  agent.AgentNameCopilot,
		SessionRef: "/path/to/file",
	}

	if err := ag.WriteSession(session); err == nil {
		t.Error("WriteSession() should error when NativeData is empty")
	}
}

func TestGetSupportedHooks(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	hooks := ag.GetSupportedHooks()

	expected := []agent.HookType{
		agent.HookSessionStart,
		agent.HookSessionEnd,
		agent.HookStop,
		agent.HookUserPromptSubmit,
		agent.HookPreToolUse,
		agent.HookPostToolUse,
	}

	if len(hooks) != len(expected) {
		t.Errorf("GetSupportedHooks() returned %d hooks, want %d", len(hooks), len(expected))
	}

	for i, hook := range expected {
		if hooks[i] != hook {
			t.Errorf("GetSupportedHooks()[%d] = %v, want %v", i, hooks[i], hook)
		}
	}
}

func TestChunkTranscript_SmallContent(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	content := []byte(`{"role":"user","content":"hello"}
{"role":"assistant","content":"hi"}`)

	chunks, err := ag.ChunkTranscript(content, agent.MaxChunkSize)
	if err != nil {
		t.Fatalf("ChunkTranscript() error = %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk, got %d", len(chunks))
	}
}

func TestChunkTranscript_RoundTrip(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	original := `{"role":"user","content":"hello"}
{"role":"assistant","content":"hi there"}
{"role":"user","content":"thanks"}`

	chunks, err := ag.ChunkTranscript([]byte(original), 50)
	if err != nil {
		t.Fatalf("ChunkTranscript() error = %v", err)
	}

	reassembled, err := ag.ReassembleTranscript(chunks)
	if err != nil {
		t.Fatalf("ReassembleTranscript() error = %v", err)
	}

	if string(reassembled) != original {
		t.Errorf("Round-trip failed:\ngot:  %q\nwant: %q", string(reassembled), original)
	}
}
