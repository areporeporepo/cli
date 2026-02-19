package copilot

import (
	"os"
	"path/filepath"
	"testing"
)

// sampleEventsJSONL contains a realistic Copilot events.jsonl with various event types.
// Based on actual captured data from Copilot CLI v0.0.411.
const sampleEventsJSONL = `{"type":"session.start","data":{"sessionId":"733bfbbd-bd9b-4750-a55b-3c8cc42629de","version":1,"producer":"copilot-cli","copilotVersion":"0.0.411","startTime":"2025-06-19T18:21:24.159Z","context":{"cwd":"/Users/test/project","gitRoot":"/Users/test/project","branch":"main","repository":"test/project"}},"id":"e1","timestamp":"2025-06-19T18:21:24.159Z","parentId":null}
{"type":"session.model_change","data":{"model":"claude-sonnet-4-20250514"},"id":"e2","timestamp":"2025-06-19T18:21:24.160Z","parentId":"e1"}
{"type":"user.message","data":{"content":"Read the file main.go","transformedContent":"(context injected) Read the file main.go"},"id":"e3","timestamp":"2025-06-19T18:21:24.161Z","parentId":"e1"}
{"type":"assistant.turn_start","data":{},"id":"e4","timestamp":"2025-06-19T18:21:25.000Z","parentId":"e3"}
{"type":"assistant.message","data":{"messageId":"msg1","content":"","toolRequests":[{"toolCallId":"tc1","name":"view","arguments":{"path":"/Users/test/project/main.go"},"type":"function"}]},"id":"e5","timestamp":"2025-06-19T18:21:26.000Z","parentId":"e4"}
{"type":"tool.execution_start","data":{"toolCallId":"tc1","toolName":"view","arguments":{"path":"/Users/test/project/main.go"}},"id":"e6","timestamp":"2025-06-19T18:21:26.001Z","parentId":"e5"}
{"type":"tool.execution_complete","data":{"toolCallId":"tc1","success":true,"result":{"content":"package main..."}},"id":"e7","timestamp":"2025-06-19T18:21:26.100Z","parentId":"e6"}
{"type":"assistant.message","data":{"messageId":"msg2","content":"I've read the file main.go. It contains a basic Go program.","toolRequests":[]},"id":"e8","timestamp":"2025-06-19T18:21:27.000Z","parentId":"e7"}
{"type":"assistant.turn_end","data":{"stopReason":"end_turn"},"id":"e9","timestamp":"2025-06-19T18:21:27.100Z","parentId":"e8"}
{"type":"user.message","data":{"content":"thanks","transformedContent":"thanks"},"id":"e10","timestamp":"2025-06-19T18:21:30.000Z","parentId":"e9"}
{"type":"assistant.message","data":{"messageId":"msg3","content":"You're welcome!","toolRequests":[]},"id":"e11","timestamp":"2025-06-19T18:21:31.000Z","parentId":"e10"}`

// sampleEventsWithCreate contains events.jsonl data with file modification tools (create, edit).
const sampleEventsWithCreate = `{"type":"session.start","data":{"sessionId":"10855a12-xxxx","version":1},"id":"e1","timestamp":"2025-06-19T18:25:00.000Z","parentId":null}
{"type":"user.message","data":{"content":"Create a test file","transformedContent":"Create a test file"},"id":"e2","timestamp":"2025-06-19T18:25:01.000Z","parentId":"e1"}
{"type":"assistant.turn_start","data":{},"id":"e3","timestamp":"2025-06-19T18:25:02.000Z","parentId":"e2"}
{"type":"assistant.message","data":{"messageId":"msg1","content":"","toolRequests":[{"toolCallId":"tc1","name":"create","arguments":{"path":"/tmp/test.txt","content":"hello"},"type":"function"}]},"id":"e4","timestamp":"2025-06-19T18:25:03.000Z","parentId":"e3"}
{"type":"tool.execution_start","data":{"toolCallId":"tc1","toolName":"create","arguments":{"path":"/tmp/test.txt","content":"hello"}},"id":"e5","timestamp":"2025-06-19T18:25:03.001Z","parentId":"e4"}
{"type":"tool.execution_complete","data":{"toolCallId":"tc1","success":true,"result":{"content":"Created /tmp/test.txt"}},"id":"e6","timestamp":"2025-06-19T18:25:03.100Z","parentId":"e5"}
{"type":"assistant.message","data":{"messageId":"msg2","content":"","toolRequests":[{"toolCallId":"tc2","name":"edit","arguments":{"file_path":"/tmp/other.go","old_string":"foo","new_string":"bar"},"type":"function"}]},"id":"e7","timestamp":"2025-06-19T18:25:04.000Z","parentId":"e6"}
{"type":"tool.execution_start","data":{"toolCallId":"tc2","toolName":"edit","arguments":{"file_path":"/tmp/other.go","old_string":"foo","new_string":"bar"}},"id":"e8","timestamp":"2025-06-19T18:25:04.001Z","parentId":"e7"}
{"type":"tool.execution_complete","data":{"toolCallId":"tc2","success":true,"result":{"content":"Edited /tmp/other.go"}},"id":"e9","timestamp":"2025-06-19T18:25:04.100Z","parentId":"e8"}
{"type":"assistant.message","data":{"messageId":"msg3","content":"Done! Created test.txt and edited other.go.","toolRequests":[]},"id":"e10","timestamp":"2025-06-19T18:25:05.000Z","parentId":"e9"}`

func Test_countLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want int
	}{
		{"empty", "", 0},
		{"single line", `{"type":"session.start","data":{}}`, 1},
		{"two lines", `{"type":"session.start","data":{}}
{"type":"user.message","data":{}}`, 2},
		{"trailing newline", `{"type":"session.start","data":{}}
`, 1},
		{"blank lines ignored", `{"type":"session.start","data":{}}

{"type":"user.message","data":{}}
`, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countLines([]byte(tt.data))
			if got != tt.want {
				t.Errorf("countLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func Test_extractModifiedFilesFromLines(t *testing.T) {
	t.Parallel()

	files := extractModifiedFilesFromLines([]byte(sampleEventsWithCreate), 0)

	if len(files) != 2 {
		t.Fatalf("extractModifiedFilesFromLines() got %d files, want 2: %v", len(files), files)
	}

	expectedFiles := map[string]bool{"/tmp/test.txt": true, "/tmp/other.go": true}
	for _, f := range files {
		if !expectedFiles[f] {
			t.Errorf("unexpected file: %s", f)
		}
	}
}

func Test_extractModifiedFilesFromLines_ViewNotIncluded(t *testing.T) {
	t.Parallel()

	// sampleEventsJSONL has a "view" tool call but no create/edit
	files := extractModifiedFilesFromLines([]byte(sampleEventsJSONL), 0)

	if len(files) != 0 {
		t.Errorf("extractModifiedFilesFromLines() got %d files, want 0 (view is not a modification tool): %v", len(files), files)
	}
}

func Test_extractModifiedFilesFromLines_WithOffset(t *testing.T) {
	t.Parallel()

	// sampleEventsWithCreate: create tool is at line index 4, edit tool is at line index 7
	// Skip first 6 lines (indices 0-5) to only get the edit tool
	files := extractModifiedFilesFromLines([]byte(sampleEventsWithCreate), 6)

	if len(files) != 1 {
		t.Fatalf("extractModifiedFilesFromLines() got %d files, want 1: %v", len(files), files)
	}
	if files[0] != "/tmp/other.go" {
		t.Errorf("expected /tmp/other.go, got %s", files[0])
	}
}

func Test_extractModifiedFilesFromLines_Empty(t *testing.T) {
	t.Parallel()

	files := extractModifiedFilesFromLines([]byte(""), 0)
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty data, got %d", len(files))
	}
}

func Test_extractModifiedFilesFromLines_Dedup(t *testing.T) {
	t.Parallel()

	// Two create events for the same file
	data := `{"type":"tool.execution_start","data":{"toolCallId":"tc1","toolName":"create","arguments":{"path":"/tmp/same.txt"}},"id":"e1","timestamp":"2025-01-01T00:00:00Z","parentId":null}
{"type":"tool.execution_start","data":{"toolCallId":"tc2","toolName":"create","arguments":{"path":"/tmp/same.txt"}},"id":"e2","timestamp":"2025-01-01T00:00:01Z","parentId":"e1"}`

	files := extractModifiedFilesFromLines([]byte(data), 0)

	if len(files) != 1 {
		t.Errorf("expected 1 deduplicated file, got %d: %v", len(files), files)
	}
}

func Test_extractPromptsFromData(t *testing.T) {
	t.Parallel()

	prompts, err := extractPromptsFromData([]byte(sampleEventsJSONL), 0)
	if err != nil {
		t.Fatalf("extractPromptsFromData() error = %v", err)
	}

	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d: %v", len(prompts), prompts)
	}
	if prompts[0] != "Read the file main.go" {
		t.Errorf("first prompt = %q, want 'Read the file main.go'", prompts[0])
	}
	if prompts[1] != "thanks" {
		t.Errorf("second prompt = %q, want 'thanks'", prompts[1])
	}
}

func Test_extractPromptsFromData_WithOffset(t *testing.T) {
	t.Parallel()

	// The second user.message is at line index 9 (0-based) in sampleEventsJSONL
	// Skip first 5 lines to skip the first user.message
	prompts, err := extractPromptsFromData([]byte(sampleEventsJSONL), 5)
	if err != nil {
		t.Fatalf("extractPromptsFromData() error = %v", err)
	}

	if len(prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d: %v", len(prompts), prompts)
	}
	if prompts[0] != "thanks" {
		t.Errorf("prompt = %q, want 'thanks'", prompts[0])
	}
}

func Test_extractLastAssistantMessage(t *testing.T) {
	t.Parallel()

	result := extractLastAssistantMessage([]byte(sampleEventsJSONL))
	if result != "You're welcome!" {
		t.Errorf("extractLastAssistantMessage() = %q, want 'You're welcome!'", result)
	}
}

func Test_extractLastAssistantMessage_SkipsEmptyContent(t *testing.T) {
	t.Parallel()

	// sampleEventsWithCreate has assistant messages with empty content (tool-only) and one final message
	result := extractLastAssistantMessage([]byte(sampleEventsWithCreate))
	if result != "Done! Created test.txt and edited other.go." {
		t.Errorf("extractLastAssistantMessage() = %q, want 'Done! Created test.txt and edited other.go.'", result)
	}
}

func Test_extractLastAssistantMessage_Empty(t *testing.T) {
	t.Parallel()

	result := extractLastAssistantMessage([]byte(""))
	if result != "" {
		t.Errorf("expected empty string for empty data, got %q", result)
	}
}

func Test_extractLastAssistantMessage_NoAssistant(t *testing.T) {
	t.Parallel()

	data := `{"type":"user.message","data":{"content":"hello"},"id":"e1","timestamp":"2025-01-01T00:00:00Z","parentId":null}`
	result := extractLastAssistantMessage([]byte(data))
	if result != "" {
		t.Errorf("expected empty string when no assistant messages, got %q", result)
	}
}

func Test_sliceFromLine(t *testing.T) {
	t.Parallel()

	data := []byte(`{"type":"session.start","data":{},"id":"e1","timestamp":"T","parentId":null}
{"type":"user.message","data":{"content":"hi"},"id":"e2","timestamp":"T","parentId":"e1"}
{"type":"assistant.message","data":{"content":"hello"},"id":"e3","timestamp":"T","parentId":"e2"}`)

	t.Run("start from 0", func(t *testing.T) {
		t.Parallel()
		result := sliceFromLine(data, 0)
		if string(result) != string(data) {
			t.Errorf("sliceFromLine(0) should return original data")
		}
	})

	t.Run("start from 1", func(t *testing.T) {
		t.Parallel()
		result := sliceFromLine(data, 1)
		if result == nil {
			t.Fatal("sliceFromLine(1) returned nil")
		}
		lines := countLines(result)
		if lines != 2 {
			t.Errorf("sliceFromLine(1) has %d lines, want 2", lines)
		}
	})

	t.Run("start past end", func(t *testing.T) {
		t.Parallel()
		result := sliceFromLine(data, 10)
		if result != nil {
			t.Errorf("sliceFromLine(10) should return nil, got %q", result)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		t.Parallel()
		result := sliceFromLine([]byte(""), 0)
		if len(result) != 0 {
			t.Errorf("sliceFromLine on empty data should return empty")
		}
	})
}

func TestGetTranscriptPosition(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	transcriptPath := filepath.Join(tempDir, "events.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(sampleEventsJSONL), 0o644); err != nil {
		t.Fatal(err)
	}

	ag := &CopilotAgent{}
	pos, err := ag.GetTranscriptPosition(transcriptPath)
	if err != nil {
		t.Fatalf("GetTranscriptPosition() error = %v", err)
	}
	if pos != 11 {
		t.Errorf("GetTranscriptPosition() = %d, want 11", pos)
	}
}

func TestGetTranscriptPosition_EmptyPath(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	pos, err := ag.GetTranscriptPosition("")
	if err != nil {
		t.Fatalf("GetTranscriptPosition() error = %v", err)
	}
	if pos != 0 {
		t.Errorf("GetTranscriptPosition('') = %d, want 0", pos)
	}
}

func TestGetTranscriptPosition_NonexistentFile(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	pos, err := ag.GetTranscriptPosition("/nonexistent/file.jsonl")
	if err != nil {
		t.Fatalf("GetTranscriptPosition() error = %v", err)
	}
	if pos != 0 {
		t.Errorf("GetTranscriptPosition() = %d, want 0 for nonexistent file", pos)
	}
}

func TestExtractModifiedFilesFromOffset(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	transcriptPath := filepath.Join(tempDir, "events.jsonl")

	if err := os.WriteFile(transcriptPath, []byte(sampleEventsWithCreate), 0o644); err != nil {
		t.Fatal(err)
	}

	ag := &CopilotAgent{}
	files, pos, err := ag.ExtractModifiedFilesFromOffset(transcriptPath, 0)
	if err != nil {
		t.Fatalf("ExtractModifiedFilesFromOffset() error = %v", err)
	}
	if pos != 10 {
		t.Errorf("position = %d, want 10", pos)
	}
	if len(files) != 2 {
		t.Errorf("files = %d, want 2: %v", len(files), files)
	}
}

func TestExtractModifiedFilesFromOffset_EmptyPath(t *testing.T) {
	t.Parallel()

	ag := &CopilotAgent{}
	files, pos, err := ag.ExtractModifiedFilesFromOffset("", 0)
	if err != nil {
		t.Fatalf("ExtractModifiedFilesFromOffset() error = %v", err)
	}
	if pos != 0 || len(files) != 0 {
		t.Errorf("expected empty results for empty path, got pos=%d files=%v", pos, files)
	}
}

func TestParseTranscript(t *testing.T) {
	t.Parallel()

	events, err := ParseTranscript([]byte(sampleEventsJSONL))
	if err != nil {
		t.Fatalf("ParseTranscript() error = %v", err)
	}

	if len(events) != 11 {
		t.Fatalf("ParseTranscript() returned %d events, want 11", len(events))
	}

	// Verify event types (using string literals for types without constants)
	expectedTypes := []string{
		"session.start",
		"session.model_change",
		EventUserMessage,
		"assistant.turn_start",
		EventAssistantMessage,
		EventToolExecStart,
		"tool.execution_complete",
		EventAssistantMessage,
		"assistant.turn_end",
		EventUserMessage,
		EventAssistantMessage,
	}

	for i, expected := range expectedTypes {
		if events[i].Type != expected {
			t.Errorf("events[%d].Type = %q, want %q", i, events[i].Type, expected)
		}
	}
}

func TestParseTranscript_Empty(t *testing.T) {
	t.Parallel()

	events, err := ParseTranscript([]byte(""))
	if err != nil {
		t.Fatalf("ParseTranscript() error = %v", err)
	}
	if len(events) != 0 {
		t.Errorf("ParseTranscript() returned %d events for empty input, want 0", len(events))
	}
}

func TestParseTranscript_MalformedLines(t *testing.T) {
	t.Parallel()

	// Malformed lines should be skipped
	data := `{"type":"user.message","data":{"content":"hello"},"id":"e1","timestamp":"T","parentId":null}
not valid json
{"type":"assistant.message","data":{"content":"hi"},"id":"e2","timestamp":"T","parentId":"e1"}`

	events, err := ParseTranscript([]byte(data))
	if err != nil {
		t.Fatalf("ParseTranscript() error = %v", err)
	}
	if len(events) != 2 {
		t.Errorf("ParseTranscript() returned %d events, want 2 (skipping malformed line)", len(events))
	}
}
