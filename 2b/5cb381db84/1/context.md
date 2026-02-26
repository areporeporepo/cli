# Session Context

## User Prompts

### Prompt 1

Base directory for this skill: /Users/alex/workspace/cli/.claude/skills/debug-e2e

# Debug Entire CLI via E2E Artifacts

Diagnose Entire CLI bugs using captured artifacts from the E2E test suite. Artifacts are written to `e2e/artifacts/` locally or downloaded from CI via GitHub Actions.

## Inputs

The user provides either:
- **A test run directory:** `e2e/artifacts/{timestamp}/` — triage all failures
- **A specific test directory:** `e2e/artifacts/{timestamp}/{TestName}-{agent}/` — debug one...

### Prompt 2

is the claude invocation in the e2e's picking up user settings? we shouldn't be, right?

### Prompt 3

could we enable sandboxing? https://code.claude.com/docs/en/settings#sandbox-settings

### Prompt 4

the 'auth' fixes we did yesterday to use keychain based auth - do we even need the claude dir to copy anything?

### Prompt 5

let's try it

### Prompt 6

Base directory for this skill: /Users/alex/.claude/skills/tdd-dev

# TDD Developer

Implement features using Test-Driven Development and Clean Code principles.

## TDD Cycle

1. **Red** - Write failing test first
2. **Green** - Minimal code to pass
3. **Refactor** - Clean up, tests stay green

## Clean Code Standards

- **Names reveal intent** - Variables, functions, classes
- **Small functions** - One responsibility each
- **DRY** - Extract duplication
- **SOLID** - Single responsibility, Op...

### Prompt 7

but keep the comment for DISABLE_NONESSENTIAL

### Prompt 8

commit this, then let's have a look at the latest test failures /Users/alex/workspace/cli/e2e/artifacts/2026-02-26T12-25-30

