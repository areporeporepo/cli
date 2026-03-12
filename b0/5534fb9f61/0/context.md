# Session Context

## User Prompts

### Prompt 1

in the claude e2e tests - how exactly are we setting up the claude settings json? what are we doing with UserDir in the claude eBootstrap()?

### Prompt 2

Base directory for this skill: /Users/alex/workspace/cli/.claude/skills/debug-e2e

# Debug Entire CLI via E2E Artifacts

Diagnose Entire CLI bugs using captured artifacts from the E2E test suite. Artifacts are written to `e2e/artifacts/` locally or downloaded from CI via GitHub Actions.

## Inputs

The user provides either:
- **A test run directory:** `e2e/artifacts/{timestamp}/` — triage all failures
- **A specific test directory:** `e2e/artifacts/{timestamp}/{TestName}-{agent}/` — debug one...

### Prompt 3

/Users/alex/workspace/cli/e2e/artifacts/2026-03-05T10-50-34

### Prompt 4

also our 'mise run test:e2e' without arguments seems to be broken now
```
> mise run test:e2e
[build] $ ~/workspace/cli/mise-tasks/build
artifacts: /Users/alex/workspace/cli/e2e/artifacts/2026-03-05T11-02-52

=== Failed
=== FAIL: e2e/tests  (0.00s)
preflight: missing required binaries: [agent]
```

### Prompt 5

oh, do I not have factory droid installed?

### Prompt 6

can we improve the preflight error to at least let the user know what's missing?

### Prompt 7

how about just the agent name in the brackets?

### Prompt 8

commit, push, PR

