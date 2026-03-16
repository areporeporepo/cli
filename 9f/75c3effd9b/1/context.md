# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: E2E Triage Skill — Add Local Mode + CI Re-runs

## Context

The current e2e-triage skill is CI-only: it downloads artifacts from a failed GitHub Actions run and classifies failures based on static artifact analysis. This has two problems:

1. **CI mode lacks re-run verification** — classification relies on heuristic signals (timeout patterns, error strings) rather than actually re-running the test to see if it reproduces.
2. **No local triage path** — wh...

### Prompt 2

Tool loaded.

### Prompt 3

commit

