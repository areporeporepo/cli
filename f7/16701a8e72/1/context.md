# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Forward AWS SIGV4 env vars through tmux for Kiro E2E tests

## Context

All Kiro E2E tests in CI fail with a login prompt timeout:
```
waiting for kiro startup: timed out waiting for "!>" after 30s
--- pane content ---
You are not logged in. Login now?
```

**Root cause:** Kiro authenticates via AWS SIGV4 env vars (`AMAZON_Q_SIGV4`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`). These are set in the Go test process but *...

### Prompt 2

commit and push

