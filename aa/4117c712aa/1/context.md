# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: CI — `mkdir -p ~/.local/bin` before copying kiro-cli-chat

## Context

CI fails with `cp: cannot create regular file '/home/runner/.local/bin/': Not a directory`. The `~/.local/bin` directory doesn't exist on the GitHub Actions runner by default. Our kiro install step tries to `cp` into it without creating it first.

## Changes

### 1. `.github/workflows/e2e.yml` — Add `mkdir -p`

Add `mkdir -p ~/.local/bin` before the `cp` command in the kiro case block ...

### Prompt 2

commit and push

