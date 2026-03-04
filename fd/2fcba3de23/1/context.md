# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Realistic Commit Hook Performance Test

## Context

A user with ~95 sessions experienced 25.5s commit time vs 0.335s without Entire. Our current perf test only reaches ~1.8s for 100 sessions because it uses a shallow clone with minimal refs/objects. We need a test that measures the real overhead by comparing a control commit (no Entire) against a commit with Entire hooks active, using a realistic repo and hundreds of sessions.

## File to Modify

**`cmd/...

### Prompt 2

okay let's run it

### Prompt 3

what are the main scaling dimensions here that impact the latency and by how much?


write up your findings in a .md file and commit to the branch

