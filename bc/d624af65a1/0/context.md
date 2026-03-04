# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Implement Cursor as an External Agent Binary

## Context

The Cursor agent is currently a built-in agent compiled into the CLI (`cmd/entire/cli/agent/cursor/`). We want to extract it into a standalone external agent binary (`entire-agent-cursor`) that communicates with the CLI via the external agent protocol (subcommand-based JSON over stdin/stdout). This makes Cursor a plugin that can be distributed and updated independently.

## Approach

Create a self...

### Prompt 2

now I would like to manually test the new setup, can you setup `/Users/nodo/work/test-1` to use 1. the newly build entire cli 2. the new entire-agent-cursor 3. add the feature flag to the repo to use external agents

### Prompt 3

can you test that for me?

### Prompt 4

[Request interrupted by user]

