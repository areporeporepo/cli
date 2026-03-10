# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Restructure E2E triage/debug into a plugin

## Context

The `e2e-triage` and `debug-e2e` skills are standalone skills that should be restructured into a plugin (like `agent-integration`) with clear command separation: triage-ci (report), debug (deep-dive), and implement (apply fixes + verify).

## Current Structure (to be deleted)

```
.claude/skills/e2e-triage/SKILL.md
.claude/skills/e2e-triage/README.md
.claude/skills/debug-e2e/SKILL.md
```

## Target ...

### Prompt 2

commit

