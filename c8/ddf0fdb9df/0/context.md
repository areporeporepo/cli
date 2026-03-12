# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Remove CI mode from e2e-triage skill

## Context
The e2e-triage skill currently has two modes: Local and CI. We want to remove the CI mode (branch creation, PRs, `triage-summary.json`, CI re-runs via `gh workflow run`) while keeping the skill useful for local debugging of CI failures (downloading artifacts, analyzing them, running tests locally).

## Files to modify

### 1. `.claude/skills/e2e-triage/SKILL.md`
- Update description (line 3) to remove CI r...

### Prompt 2

commit

