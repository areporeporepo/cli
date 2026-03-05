# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Switch Kiro E2E to `--no-interactive` mode

## Context

Kiro E2E tests fail in CI with "You are not logged in" because kiro-cli's **interactive TUI doesn't support SIGV4 env-var auth**. The `setup-kiro-action` reference confirms SIGV4 is designed for `--no-interactive` only — all CI examples use it.

**Verified:** `--no-interactive` **fires hooks** (the comment at `kiro.go:127` was wrong):
```
✓ 1 of 1 hooks finished in 0.09 s  (×3 hooks)
```

## Changes
...

### Prompt 2

commit and push

