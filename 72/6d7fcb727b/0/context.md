# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Create Example External Agent "test-agent"

## Context

The codebase has a new external agent protocol (`cmd/entire/cli/agent/external/`) that discovers binaries named `entire-agent-<name>` from `$PATH` and integrates them as agents. We need an example implementation called `entire-agent-test` to serve as a reference for building external agents.

## Approach

Create a minimal Go binary at `cmd/entire-agent-test/` that implements the external agent proto...

