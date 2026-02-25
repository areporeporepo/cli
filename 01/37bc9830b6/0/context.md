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

yes, fix it

### Prompt 3

kick off a e2e in CI please

### Prompt 4

yes

### Prompt 5

are the last round of failures legit? are there some cases where we expect the shadow branch to hang around? /debug-e2e

### Prompt 6

[Request interrupted by user for tool use]

### Prompt 7

oh, I meant the local run

### Prompt 8

this one TestTrailerRemovalSkipsCondensation let's just fix the test - the shadow branch is legitimately hanging around because of the skipped condensation

### Prompt 9

they others are legit bugs correct?

### Prompt 10

okay, let's commit and push

