# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Kiro E2E Auth — Use `kiro-cli-chat` binary

## Context

Kiro E2E tests fail immediately (0.5s) with `OAuth error: Failed to open browser`. The root cause is that our CI uses `kiro-cli` (the desktop app wrapper) which ignores `AMAZON_Q_SIGV4` env var and always forces browser OAuth. The standalone `kiro-cli-chat` binary is the one that supports headless SIGV4 auth.

**Verified locally:**
- `AMAZON_Q_SIGV4=1 kiro-cli chat --no-interactive "hello"` → tries b...

### Prompt 2

commit and push

