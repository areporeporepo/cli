# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Remove Strategy Abstraction & Add `commit_linking` Setting

## Context

The CLI has a `Strategy` interface abstraction (`strategy.go`) designed to support multiple session strategies, but only one implementation (`ManualCommitStrategy`) has ever existed. The current behavior prompts users on every commit to link/unlink the session. Per RFD-003 decisions, we need to:

1. Remove the `Strategy` interface (inline the single implementation)
2. Add a `commit_l...

### Prompt 2

commit and push this to a new PR

