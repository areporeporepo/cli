# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Pre-seed cursor cli-config.json to avoid ENOENT race

## Context
TestResumeSquashMergeMultipleCheckpoints (cursor-cli) fails flakily on CI. Cursor's atomic config write (`cli-config.json.tmp` → `cli-config.json`) races when parallel tests trigger "Workspace Trust Required" simultaneously, causing ENOENT crashes.

## Change
**File:** `e2e/agents/cursor_cli.go` — `Bootstrap()` function (after `MkdirAll` on line 79)

Pre-seed `cli-config.json` with `{}` if i...

### Prompt 2

commit and push

