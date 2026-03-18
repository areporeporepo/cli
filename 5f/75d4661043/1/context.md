# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Add CI artifact download to e2e-triage skill

## Context

When the user says "latest CI run", the e2e-triage skill should use `scripts/download-e2e-artifacts.sh` to download artifacts from the most recent failed CI E2E run. Currently, the skill's Step L1 only mentions local artifact paths and test names — it has no guidance for fetching CI artifacts, which led to ad-hoc `gh run list` queries instead of using the existing script.

## Change

**File:** `.c...

### Prompt 2

[Request interrupted by user]

### Prompt 3

commit

