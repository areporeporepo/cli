# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Restructure Agent-Integration Skill for E2E-First TDD

## Context

The agent-integration skill claims to be E2E-driven but interleaves unit test writing after each E2E tier ("After passing, write unit tests"). When the skill ran for the Kiro integration, it wrote unit tests during development but never ran E2E tests — so E2E tests fail when actually run. The fix enforces strict discipline: E2E tests drive development, unit tests are written last.

**Core...

### Prompt 2

add section to .claude/skills/agent-integration/implementer.md after step 12 to run run all e2e tests for agent and verify they pass and fix any that fail the commadn should be mise run test:e2e --agent $AGENT_SLUG

### Prompt 3

## Context

- Current git status: On branch alisha/kiro-oneshot
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
	modified:   .claude/plugins/agent-integration/commands/implement.md
	modified:   .claude/plugins/agent-integration/commands/write-tests.md
	modified:   .claude/skills/agent-integration/SKILL.md
	modified:   .claude/skills/agent-integration/implementer.md
	modified:   .claude/skills/agent-integration/test-writer.md

Changes not staged for commit:
  (use ...

