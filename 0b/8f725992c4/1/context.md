# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Resume Only From Latest Checkpoint on Squash Merges

## Context

When a feature branch is squash-merged, all `Entire-Checkpoint` trailers from the original commits end up in the squash commit. PR #534 added logic to restore sessions from **every** checkpoint, deduplicating by session ID. But this restores stale/ended sessions that the user will never continue. We should only resume from the **latest** checkpoint — the one representing the final state of ...

