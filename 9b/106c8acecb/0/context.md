# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: Factory AI Droid TTY hang during git commit

## Context

When Factory AI Droid is told to commit code, `git commit -m "..."` triggers the `prepare-commit-msg` hook, which calls `hasTTY()` to decide whether to show an interactive y/n/a prompt. Droid's subprocess can open `/dev/tty` (so `hasTTY()` returns `true`), but Droid can't actually respond to the prompt — causing a hang/timeout.

This is the **exact same issue** Gemini CLI had, fixed by checking `GEM...

### Prompt 2

## Context

- Current git status: On branch alisha/factoryai-agent
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   cmd/entire/cli/strategy/manual_commit_hooks.go

no changes added to commit (use "git add" and/or "git commit -a")
- Current git diff (staged and unstaged changes): diff --git a/cmd/entire/cli/strategy/manual_commit_hooks.go b/cmd/entire/cli/strategy/man...

