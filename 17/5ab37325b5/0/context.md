# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Remove backward-compatibility fallbacks for unknown agent types

## Context

Agent type tracking was added on Jan 9, 2026 (6 days after repo creation) and shipped in v0.3.5 (Jan 15). There are no realistic production sessions without agent type info. The codebase has several backward-compat patterns (`DefaultAgentType`, `isSpecificAgentType`, `ResolveAgentForRewind` fallback, backfill logic in hooks) that add complexity for a case that no longer exists. ...

### Prompt 2

commit it

