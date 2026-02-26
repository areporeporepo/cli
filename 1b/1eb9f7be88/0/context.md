# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Make Agent Integration Skill Resilient to Codebase Drift

## Context

The agent-integration skill (`.claude/skills/agent-integration/`) contains ~40 hardcoded file paths, enumerated names (event types, runner classes, test files), and inline code examples with method signatures. These already drifted once (PR #498 review caught 10 issues). The root cause is that the skill prompts duplicate information that lives in source code and docs, creating multiple...

