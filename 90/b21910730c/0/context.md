# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Agent-Contributed Custom Checkpoint Files (Cursor Chat Export)

## Context

When Cursor creates a checkpoint, we want to also include a cursor-chat export archive (`.cursor-chat.json`) in the checkpoint metadata. This is agent-specific data — other agents don't have Cursor's `store.db`. We need a simple extensibility point so any agent can contribute custom files to checkpoints in the future.

## Approach: `CheckpointContributor` Optional Interface

The ...

### Prompt 2

it does not create custom files into the checkpoint at entire/checkpoint/v1 branch when the checkpoint commit is being created. why ?

