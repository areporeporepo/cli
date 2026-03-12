# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Verify Cursor CLI E2E Changes

## Context

Branch `alisha/e2e-triage-local-only` has uncommitted changes to `e2e/agents/cursor_cli.go` that fix two cursor-cli E2E issues:
1. **Isolated config dirs** — `cursorConfigDir()` creates per-session `XDG_CONFIG_HOME` to prevent ENOENT race on parallel tests
2. **Changed wait pattern** — from `PromptPattern()` (`/ commands`) to `"Add a follow-up"` to avoid premature WaitFor settling during "Thinking" phase
3. **Cleanup ...

### Prompt 2

commit and push

