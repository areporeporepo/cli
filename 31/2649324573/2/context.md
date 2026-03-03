# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix Droid E2E Remaining Timeout Failures

## Context

The tmux resize fix (already applied in `e2e/agents/droid.go`) solved the **startup** timeout issue (8/37 → 4/37 failures). The remaining 4 failures are a different problem: **agent processing timeouts**.

### Failure analysis

| Test | Duration | Timeout | Root Cause |
|------|----------|---------|------------|
| TestShadowBranchCleanedAfterAgentCommit | 112s | WaitFor 90s | Hardcoded 90s WaitFor not scale...

