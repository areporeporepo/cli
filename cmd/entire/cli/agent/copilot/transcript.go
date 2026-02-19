package copilot

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
)

// Copilot events.jsonl transcript parsing.
//
// Copilot CLI stores session transcripts as JSONL files at:
//   ~/.copilot/session-state/<sessionID>/events.jsonl
//
// Each line is a JSON object with the schema:
//   {"type":"<event-type>", "data":{...}, "id":"uuid", "timestamp":"iso8601", "parentId":"uuid|null"}
//
// Known event types:
//   session.start, session.end, session.mode_changed, session.model_change
//   user.message, assistant.message, assistant.turn_start, assistant.turn_end
//   tool.execution_start, tool.execution_complete

// countLines counts the number of non-empty lines in JSONL content.
func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	count := 0
	lines := bytes.SplitSeq(data, []byte("\n"))
	for line := range lines {
		if len(bytes.TrimSpace(line)) > 0 {
			count++
		}
	}
	return count
}

// extractModifiedFilesFromLines extracts file paths modified by tool calls
// from events.jsonl lines starting at the given offset.
// Looks for tool.execution_start events where the tool name is a file modification tool.
func extractModifiedFilesFromLines(data []byte, startOffset int) []string {
	lines := bytes.Split(data, []byte("\n"))
	fileSet := make(map[string]bool)
	var files []string

	lineIndex := 0
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		if lineIndex < startOffset {
			lineIndex++
			continue
		}
		lineIndex++

		var event Event
		if err := json.Unmarshal(trimmed, &event); err != nil {
			continue
		}

		// Look for tool execution events that modify files
		if event.Type != EventToolExecStart {
			continue
		}

		var toolData toolExecStartData
		if err := json.Unmarshal(event.Data, &toolData); err != nil {
			continue
		}

		if !isFileModificationTool(toolData.ToolName) {
			continue
		}

		file := extractFilePathFromArgs(toolData.Arguments)
		if file != "" && !fileSet[file] {
			fileSet[file] = true
			files = append(files, file)
		}
	}

	return files
}

// extractPromptsFromData extracts user prompts from events.jsonl data
// starting at the given line offset.
// Looks for user.message events and extracts the raw content (not the transformed version
// which includes injected context like timestamps and reminders).
func extractPromptsFromData(data []byte, fromOffset int) ([]string, error) {
	lines := bytes.Split(data, []byte("\n"))
	var prompts []string

	lineIndex := 0
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		if lineIndex < fromOffset {
			lineIndex++
			continue
		}
		lineIndex++

		var event Event
		if err := json.Unmarshal(trimmed, &event); err != nil {
			continue
		}

		if event.Type != EventUserMessage {
			continue
		}

		var msgData userMessageData
		if err := json.Unmarshal(event.Data, &msgData); err != nil {
			continue
		}

		if msgData.Content != "" {
			prompts = append(prompts, msgData.Content)
		}
	}

	return prompts, nil
}

// extractLastAssistantMessage extracts the last non-empty assistant message
// from events.jsonl data. Searches backwards for the last assistant.message
// event with content.
func extractLastAssistantMessage(data []byte) string {
	lines := bytes.Split(data, []byte("\n"))

	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := bytes.TrimSpace(lines[i])
		if len(trimmed) == 0 {
			continue
		}

		var event Event
		if err := json.Unmarshal(trimmed, &event); err != nil {
			continue
		}

		if event.Type != EventAssistantMessage {
			continue
		}

		var msgData assistantMessageData
		if err := json.Unmarshal(event.Data, &msgData); err != nil {
			continue
		}

		if msgData.Content != "" {
			return msgData.Content
		}
	}

	return ""
}

// ParseTranscript parses a Copilot events.jsonl transcript into structured events.
// This is used by the summarizer to build a condensed view.
func ParseTranscript(data []byte) ([]Event, error) { //nolint:unparam // error kept for API consistency with other transcript parsers
	lines := bytes.Split(data, []byte("\n"))
	var events []Event

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		var event Event
		if err := json.Unmarshal(trimmed, &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

// isFileModificationTool checks if a tool name modifies files.
func isFileModificationTool(toolName string) bool {
	return slices.Contains(fileModificationTools, toolName)
}

// extractFilePathFromArgs extracts a file path from tool arguments map.
// Checks common field names used by Copilot tools: path, file_path, filePath.
func extractFilePathFromArgs(args map[string]any) string {
	for _, key := range []string{"path", "file_path", "filePath"} {
		if v, ok := args[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// sliceFromLine returns JSONL data scoped to lines starting from startLine.
// Returns the original data if startLine <= 0.
// Returns nil if startLine exceeds the number of lines.
func sliceFromLine(data []byte, startLine int) []byte {
	if len(data) == 0 || startLine <= 0 {
		return data
	}

	lines := bytes.Split(data, []byte("\n"))
	lineIndex := 0
	var result [][]byte

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}

		if lineIndex >= startLine {
			result = append(result, line)
		}
		lineIndex++
	}

	if len(result) == 0 {
		return nil
	}

	return []byte(strings.Join(byteSlicesToStrings(result), "\n"))
}

// byteSlicesToStrings converts [][]byte to []string.
func byteSlicesToStrings(slices [][]byte) []string {
	result := make([]string, len(slices))
	for i, s := range slices {
		result[i] = string(s)
	}
	return result
}
