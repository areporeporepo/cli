# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Session Pruning — Skip Fully-Condensed Sessions in PostCommit

## Context

PostCommit iterates over ALL sessions for the worktree, including ENDED sessions that have already been fully condensed. Profiling shows 200 sessions costs ~12.6s, with attribution at 6.49s (51%), shadow cleanup at 2.66s (21%), and hasNewContent at 1.53s (12%). Every session costs ~60ms just to exist in the pipeline (shadow resolve + hasNewContent + state machine + attribution if ...

### Prompt 2

lets runt his against our benchmark and see how it does

### Prompt 3

okay let's clean this up, and put this change specifically, skipping fullycondensed sessions on a new branch with a clear outline of the problem and the solution. then commit it and push a new PR

