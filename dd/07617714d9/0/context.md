# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Remove ActionUpdateLastInteraction from EventGitCommit transitions

## Context

`entire status` shows 57+ zombie sessions from January that appear "active 2h ago". Root cause: the PostCommit hook iterates ALL sessions for the current worktree (via `findSessionsForWorktree`), runs `TransitionAndLog(EventGitCommit)` on each, which includes `ActionUpdateLastInteraction`. This refreshes `LastInteractionTime` on every commit — even for sessions that have nothi...

