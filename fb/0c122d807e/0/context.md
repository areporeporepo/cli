# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Add CI artifact download to e2e-triage skill

## Context

When the user says "latest CI run", the e2e-triage skill should use `scripts/download-e2e-artifacts.sh` to download artifacts from the most recent failed CI E2E run. Currently, the skill's Step L1 only mentions local artifact paths and test names — it has no guidance for fetching CI artifacts, which led to ad-hoc `gh run list` queries instead of using the existing script.

## Change

**File:** `.c...

### Prompt 2

commit

### Prompt 3

## Context

- Current git status: On branch alisha/e2e-triage-local-only
Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git restore <file>..." to discard changes in working directory)
	modified:   .claude/skills/e2e-triage/SKILL.md

no changes added to commit (use "git add" and/or "git commit -a")
- Current git diff (staged and unstaged changes): diff --git a/.claude/skills/e2e-triage/SKILL.md b/.claude/skills/e2e-triage/SKILL.md
index 3580...

