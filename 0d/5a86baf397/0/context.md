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

1. is that possible?

### Prompt 3

run the opencode e2es locally

### Prompt 4

this is new, the tests ran fine before

### Prompt 5

ah, nice. I guess that locking hopefully won't happen all that often?

### Prompt 6

yes

