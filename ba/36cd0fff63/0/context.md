# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Local test script for Kiro SQLite auth injection

## Context

The workflow changes are already applied. This plan adds a local test script that mirrors the CI "Inject Kiro auth token" step so we can verify auth works before pushing.

## Approach

Create `scripts/test-kiro-auth.sh` that:

1. Reads auth values from the real macOS kiro DB (`~/Library/Application Support/kiro-cli/data.sqlite3`)
2. Backs up the real DB
3. Replaces it with a fresh DB built using the...

### Prompt 2

just commit whats already staged and push

