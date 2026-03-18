# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Deep Dive: Why Cursor CLI E2E Tests Are Flaky

## Context

CI run [22831463580](https://github.com/entireio/cli/actions/runs/22831463580) (commit `a258927d`) shows 2/42 cursor-cli tests failing: `TestPartialStaging` and `TestUserSplitsAgentChanges`. Additionally, the current branch HEAD (`e90e3f9c`) reverted the per-session config isolation fix, meaning the ENOENT race condition is also unfixed.

There are **two independent flakiness sources** for cursor-cli:
...

### Prompt 2

that didn;t work update the @.claude/skills/e2e/implement.md to always run the appropriate e2e tests to verify changes

### Prompt 3

[Request interrupted by user for tool use]

