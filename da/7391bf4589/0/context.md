# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Droid E2E: Retry prompt on WaitFor timeout

## Context

Droid E2E tests sometimes time out waiting for the prompt pattern `⛬|│ >` after 60s. Instead of failing the test, we want to re-send the last prompt and wait again (one retry).

## Approach

Single-file change in `e2e/testutil/repo.go`. No interface changes, no new files.

## Changes to `e2e/testutil/repo.go`

### 1. Add `lastSentInput` field to `RepoState` (line 22)

```go
type RepoState struct {
    // ...

