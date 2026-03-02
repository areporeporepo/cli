# Session Context

## User Prompts

### Prompt 1

Base directory for this skill: /Users/alex/workspace/cli/.worktrees/3/.claude/skills/agent-integration

# Agent Integration — Full Pipeline

Run all three phases of agent integration in a single session. Parameters are collected once and reused across all phases.

## Parameters

Collect these before starting (ask the user if not provided):

| Parameter | Description | How to derive |
|-----------|-------------|---------------|
| `AGENT_NAME` | Human-readable name (e.g., "Gemini CLI") | User p...

### Prompt 2

Generate a PR title and description based on the work done in this session. 

  Instructions:
  1. Review the conversation history to understand:
     - What the user asked for
     - What was implemented
     - Key decisions and trade-offs made
     - Any issues encountered and how they were resolved

  2. Run `git diff main...HEAD` to confirm the actual file changes

  3. Generate:
     - A concise PR title (50-72 chars, imperative mood)
     - A PR description (use markdown) with:
       -...

### Prompt 3

ok let's do it

