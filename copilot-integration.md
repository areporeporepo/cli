# Plan: GitHub Copilot CLI Agent Integration

## Context

The Entire CLI currently integrates with Claude Code (stable) and Gemini CLI (preview) via a generic Agent interface. GitHub Copilot CLI is a standalone terminal agent (`copilot` command) that supports hooks similar to Claude/Gemini. This plan adds Copilot as a third agent integration (preview status).

**Good news**: All Copilot hooks include common fields `sessionId`, `transcript_path`, and `hookEventName` in their stdin JSON. Additionally, `agentStop` provides TurnEnd support. This means the integration follows the same pattern as Claude/Gemini.

---

## Hook Inventory (8 hooks)

| Copilot Hook | CLI Subcommand | Normalized Event | Notes |
|---|---|---|---|
| `sessionStart` | `session-start` | `SessionStart` | Session begins or resumes |
| `sessionEnd` | `session-end` | `SessionEnd` | Session completes/terminated |
| `userPromptSubmitted` | `user-prompt-submitted` | `TurnStart` | User submits prompt |
| `agentStop` | `agent-stop` | `TurnEnd` | Agent finished responding |
| `subagentStop` | `subagent-stop` | `SubagentEnd` | Subagent completed |
| `preToolUse` | `pre-tool-use` | `nil` | Pass-through |
| `postToolUse` | `post-tool-use` | `nil` | Pass-through |
| `errorOccurred` | `error-occurred` | `nil` | Logging only |

**Common stdin fields** (all hooks): `timestamp`, `cwd`, `sessionId`, `hookEventName`, `transcript_path`

**Hook-specific fields**: `source`/`initialPrompt` (sessionStart), `reason` (sessionEnd), `prompt` (userPromptSubmitted), `toolName`/`toolArgs` (preToolUse/postToolUse), `toolResult` (postToolUse), `error` (errorOccurred)

---

## Step 0: Verify Hook stdin Format

**Before writing code**, capture actual stdin JSON to confirm the common fields and `agentStop`/`subagentStop` format:

```json
{
  "version": 1,
  "hooks": {
    "sessionStart": [{"type": "command", "bash": "cat > /tmp/copilot-session-start.json", "timeoutSec": 10}],
    "agentStop": [{"type": "command", "bash": "cat > /tmp/copilot-agent-stop.json", "timeoutSec": 10}],
    "userPromptSubmitted": [{"type": "command", "bash": "cat > /tmp/copilot-user-prompt.json", "timeoutSec": 10}]
  }
}
```

Place at `.github/hooks/copilot-setup.json`, run `copilot`, check `/tmp/copilot-*.json`. Confirm `sessionId` and `transcript_path` are present.

---

## Step 1: Add Registry Constants

**File**: `cmd/entire/cli/agent/registry.go`

```go
// In AgentName constants:
AgentNameCopilot AgentName = "github-copilot"

// In AgentType constants:
AgentTypeCopilot AgentType = "GitHub Copilot"
```

---

## Step 2: Create `cmd/entire/cli/agent/copilot/types.go`

Follow `geminicli/types.go` pattern. Copilot uses camelCase JSON fields.

**Hook config types:**
```go
type CopilotHookConfig struct {
    Version int                           `json:"version"`
    Hooks   map[string][]CopilotHookEntry `json:"hooks"`
}

type CopilotHookEntry struct {
    Type       string `json:"type"`
    Bash       string `json:"bash"`
    TimeoutSec int    `json:"timeoutSec,omitempty"`
}
```

**Common stdin base (embedded in all hook input structs):**
```go
type hookBase struct {
    SessionID      string `json:"sessionId"`
    TranscriptPath string `json:"transcript_path"`
    Cwd            string `json:"cwd"`
    HookEventName  string `json:"hookEventName"`
    Timestamp      string `json:"timestamp"`
}
```

**Hook-specific stdin types:**
- `sessionStartRaw` - embeds `hookBase` + `Source string`, `InitialPrompt string`
- `sessionEndRaw` - embeds `hookBase` + `Reason string`
- `userPromptRaw` - embeds `hookBase` + `Prompt string`
- `agentStopRaw` - embeds `hookBase` (no extra fields expected)
- `subagentStopRaw` - embeds `hookBase` (+ possibly subagent metadata)
- `toolUseRaw` - embeds `hookBase` + `ToolName string`, `ToolArgs json.RawMessage`, `ToolResult json.RawMessage`
- `errorRaw` - embeds `hookBase` + `Error json.RawMessage`

**Tool name constants:**
```go
const (
    ToolEdit   = "edit"
    ToolCreate = "create"
)
var FileModificationTools = []string{ToolEdit, ToolCreate}
```

---

## Step 3: Create `cmd/entire/cli/agent/copilot/copilot.go`

Core agent. Follow `geminicli/gemini.go` as template.

**Self-registration:**
```go
func init() {
    agent.Register(agent.AgentNameCopilot, NewCopilotAgent)
}
```

**Identity:**
- `Name()` → `agent.AgentNameCopilot`
- `Type()` → `agent.AgentTypeCopilot`
- `Description()` → `"GitHub Copilot - GitHub's AI coding assistant"`
- `IsPreview()` → `true`
- `ProtectedDirs()` → `[]string{}` (`.github` is shared, not ours)

**Detection** (`DetectPresence()`):
- Check for `.github/hooks/` with an Entire copilot hook file present (AreHooksInstalled)
- OR check `~/.copilot/` directory existence
- Use `paths.RepoRoot()` for repo-relative checks

**Session storage:**
- `GetHookConfigPath()` → `.github/hooks/copilot-setup.json`
- `GetSessionDir(repoPath)` → `~/.copilot/session-state/` (with `ENTIRE_TEST_COPILOT_SESSION_DIR` override)
- `ResolveSessionFile(dir, sessionID)` → `filepath.Join(dir, sessionID, "events.jsonl")`
- `FormatResumeCommand(sessionID)` → `copilot --resume`

**Transcript storage** (JSONL - reuse existing helpers):
- `ReadTranscript(sessionRef)` → `os.ReadFile(sessionRef)`
- `ChunkTranscript(content, maxSize)` → `agent.ChunkJSONL(content, maxSize)`
- `ReassembleTranscript(chunks)` → `agent.ReassembleJSONL(chunks)`

**Legacy methods:** ParseHookInput, GetSessionID, ReadSession, WriteSession (follow Gemini pattern)

---

## Step 4: Create `cmd/entire/cli/agent/copilot/lifecycle.go`

Event parsing. Follow `geminicli/lifecycle.go` pattern.

**Compile-time assertions:**
```go
var _ agent.TranscriptAnalyzer = (*CopilotAgent)(nil)
```

**`ParseHookEvent` switch:**
```go
case HookNameSessionStart  → parseSessionStart(stdin)  → agent.SessionStart
case HookNameUserPromptSubmitted → parseTurnStart(stdin) → agent.TurnStart (with Prompt)
case HookNameAgentStop     → parseTurnEnd(stdin)        → agent.TurnEnd
case HookNameSessionEnd    → parseSessionEnd(stdin)     → agent.SessionEnd
case HookNameSubagentStop  → parseSubagentEnd(stdin)    → agent.SubagentEnd
case HookNamePreToolUse, HookNamePostToolUse, HookNameErrorOccurred → nil
```

**Internal parse functions** (use `agent.ReadAndParseHookInput[T]` generic helper):
- `parseSessionStart` → reads `sessionStartRaw`, returns Event{SessionStart, SessionID, SessionRef}
- `parseTurnStart` → reads `userPromptRaw`, returns Event{TurnStart, SessionID, SessionRef, Prompt}
- `parseTurnEnd` → reads `agentStopRaw`, returns Event{TurnEnd, SessionID, SessionRef}
- `parseSessionEnd` → reads `sessionEndRaw`, returns Event{SessionEnd, SessionID, SessionRef}
- `parseSubagentEnd` → reads `subagentStopRaw`, returns Event{SubagentEnd, SessionID, SessionRef}

**TranscriptAnalyzer methods:**
- `GetTranscriptPosition(path)` → count lines in JSONL file
- `ExtractModifiedFilesFromOffset(path, offset)` → parse JSONL from line offset, extract file paths from edit/create events
- `ExtractPrompts(sessionRef, fromOffset)` → extract user message content from JSONL
- `ExtractSummary(sessionRef)` → last assistant message from JSONL

---

## Step 5: Create `cmd/entire/cli/agent/copilot/hooks.go`

Hook installation into `.github/hooks/copilot-setup.json`. Simpler than Claude/Gemini since config is a flat JSON file (no matchers).

**Constants:**
```go
const (
    HookNameSessionStart       = "session-start"
    HookNameSessionEnd         = "session-end"
    HookNameUserPromptSubmitted = "user-prompt-submitted"
    HookNameAgentStop          = "agent-stop"
    HookNameSubagentStop       = "subagent-stop"
    HookNamePreToolUse         = "pre-tool-use"
    HookNamePostToolUse        = "post-tool-use"
    HookNameErrorOccurred      = "error-occurred"
)

// Maps CLI subcommand names to Copilot's native hook names (camelCase in config)
var hookNameToConfigKey = map[string]string{
    HookNameSessionStart:        "sessionStart",
    HookNameSessionEnd:          "sessionEnd",
    HookNameUserPromptSubmitted: "userPromptSubmitted",
    HookNameAgentStop:           "agentStop",
    HookNameSubagentStop:        "subagentStop",
    HookNamePreToolUse:          "preToolUse",
    HookNamePostToolUse:         "postToolUse",
    HookNameErrorOccurred:       "errorOccurred",
}
```

**`InstallHooks(localDev, force)`:**
1. `paths.RepoRoot()` to find repo root
2. Read/create `.github/hooks/copilot-setup.json`
3. Parse into `CopilotHookConfig` (or create with `version: 1`)
4. For idempotency: check if existing commands match expected
5. Remove existing Entire hooks if force or mode change
6. For each of 8 hooks, add entry:
   ```json
   {"type": "command", "bash": "entire hooks github-copilot <subcommand>", "timeoutSec": 10}
   ```
7. Create `.github/hooks/` dir if needed
8. Write back with `json.MarshalIndent`
9. Return count (8)

**`UninstallHooks()`:** Remove entries matching `entireHookPrefixes`, remove config key if array empty

**`AreHooksInstalled()`:** Check if any entry matches `entireHookPrefixes`

**`entireHookPrefixes`:** `["entire ", "go run ${COPILOT_PROJECT_DIR}/cmd/entire/main.go "]`

---

## Step 6: Create `cmd/entire/cli/agent/copilot/transcript.go`

Transcript parsing for `events.jsonl`. Basic implementation, refined after verifying actual format in Step 0.

Since Copilot uses JSONL (like Claude Code), the transcript parsing follows the Claude pattern - line-based offsets, JSONL line splitting, etc. The exact event schema will be adapted based on the actual `events.jsonl` content discovered in Step 0.

**Key functions:**
- Parse JSONL events line by line
- Extract file paths from tool call events (edit, create tools)
- Extract user prompts from user message events
- Extract assistant responses for summaries
- Token usage calculation (if token fields are present in events)

---

## Step 7: Update Existing Files

### `cmd/entire/cli/hooks_cmd.go`
Add blank import:
```go
_ "github.com/entireio/cli/cmd/entire/cli/agent/copilot"
```

### `cmd/entire/cli/hook_registry.go`
Add import and update `getHookType()`:
```go
import "github.com/entireio/cli/cmd/entire/cli/agent/copilot"

// In getHookType():
case copilot.HookNamePreToolUse, copilot.HookNamePostToolUse:
    return "tool"
case copilot.HookNameSubagentStop:
    return "subagent"
case copilot.HookNameErrorOccurred:
    return "error"
```

### `cmd/entire/cli/setup_test.go`
Add blank import: `_ "github.com/entireio/cli/cmd/entire/cli/agent/copilot"`

### `cmd/entire/cli/strategy/rewind_test.go`
Add blank import: `_ "github.com/entireio/cli/cmd/entire/cli/agent/copilot"`

---

## Step 8: Write Tests

All tests use `t.Parallel()`. Follow `geminicli/*_test.go` patterns.

- **`copilot_test.go`**: Identity, detection, session storage, transcript chunk/reassemble
- **`hooks_test.go`**: Install (fresh, localDev, idempotent, force, preserve user hooks), uninstall, AreHooksInstalled
- **`lifecycle_test.go`**: ParseHookEvent for each hook → correct EventType, nil for pass-throughs, invalid/empty input errors
- **`transcript_test.go`**: JSONL parsing, file extraction, prompt extraction, summary extraction

---

## Files Summary

### New files (`cmd/entire/cli/agent/copilot/`):
| File | Template | Purpose |
|---|---|---|
| `types.go` | `geminicli/types.go` | Type definitions |
| `copilot.go` | `geminicli/gemini.go` | Core agent impl |
| `lifecycle.go` | `geminicli/lifecycle.go` | Event parsing |
| `hooks.go` | `geminicli/hooks.go` | Hook install/uninstall (simpler config) |
| `transcript.go` | `geminicli/transcript.go` | JSONL transcript parsing |
| `copilot_test.go` | `geminicli/gemini_test.go` | Core tests |
| `hooks_test.go` | `geminicli/hooks_test.go` | Hook tests |
| `lifecycle_test.go` | `geminicli/lifecycle_test.go` | Lifecycle tests |
| `transcript_test.go` | `geminicli/transcript_test.go` | Transcript tests |

### Modified files:
| File | Change |
|---|---|
| `cmd/entire/cli/agent/registry.go` | Add `AgentNameCopilot`, `AgentTypeCopilot` constants |
| `cmd/entire/cli/hooks_cmd.go` | Add copilot blank import |
| `cmd/entire/cli/hook_registry.go` | Add copilot import + update `getHookType()` |
| `cmd/entire/cli/setup_test.go` | Add copilot blank import |
| `cmd/entire/cli/strategy/rewind_test.go` | Add copilot blank import |

---

## Verification

1. **Lint + format**: `mise run fmt && mise run lint`
2. **Tests**: `mise run test:ci` (unit + integration)
3. **Manual - hook install**: `entire enable --agent github-copilot` → verify `.github/hooks/copilot-setup.json` created
4. **Manual - commands exist**: `entire hooks github-copilot --help` → shows all 8 subcommands
5. **Manual - agent detection**: Verify `github-copilot` appears in agent list
6. **Manual - with Copilot CLI**: Start `copilot` session → hooks fire → make commit → checkpoint created → `entire session log` shows data
