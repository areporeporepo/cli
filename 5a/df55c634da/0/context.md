# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Isolate integration tests from user git config

## Context

The user's `~/.gitconfig` sets `core.excludesFile` to `~/.gitignore-global`, which contains `*.log`. This caused `TestShadow_UntrackedFilePreservation` to fail because `git ls-files --others --exclude-standard` (used by `collectUntrackedFiles()` during rewind) treated `temp-debug.log` as gitignored and invisible.

The first commit fixed `cliEnv()` in `testenv.go` by adding `GIT_CONFIG_GLOBAL=/de...

### Prompt 2

[Request interrupted by user for tool use]

### Prompt 3

Commit with explanation.

