# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Use setup-kiro-action@v1 with v1.24.0 in e2e-isolated

## Context

Testing whether kiro-cli-chat v1.24.0 supports SIGV4 auth in CI. The `clouatre-labs/setup-kiro-action@v1` handles install, caching, and SIGV4 env var setup. Replace the manual kiro install block with the action.

## Changes — `.github/workflows/e2e-isolated.yml`

1. **Remove the kiro case block** (lines 45-52) from the "Install agent CLI" step — the action handles it
2. **Move the "Configure AW...

### Prompt 2

commit and push

### Prompt 3

update kiro version to 1.22 and commit and push

