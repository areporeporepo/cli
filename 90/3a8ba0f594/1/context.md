# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: e2e.yml `matrix` not available in job-level `if`

## Context
GitHub Actions doesn't allow referencing `matrix` context in a job-level `if` condition (line 36). The `matrix` context is only available inside steps. The intent is to filter matrix entries when a specific agent is selected via `workflow_dispatch`.

## Approach: Dynamic matrix via setup job

Replace the static matrix + job-level `if` with a setup job that computes the matrix dynamically.

### C...

### Prompt 2

commit and push

