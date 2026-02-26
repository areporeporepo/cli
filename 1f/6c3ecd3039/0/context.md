# Session Context

## User Prompts

### Prompt 1

You are an expert code reviewer. Follow these steps:

      1. If no PR number is provided in the args, run `gh pr list` to show open PRs
      2. If a PR number is provided, run `gh pr view <number>` to get PR details
      3. Run `gh pr diff <number>` to get the diff
      4. Analyze the changes and provide a thorough code review that includes:
         - Overview of what the PR does
         - Analysis of code quality and style
         - Specific suggestions for improvements
         - An...

### Prompt 2

for #10, this is by design - we wanted to have the ability to build the binary from different branches and test it against a stable set of e2e tests (and to switch and see regressions/variations). should we make this an option instead?

summarise the fixes we need please.

### Prompt 3

This session is being continued from a previous conversation that ran out of context. The summary below covers the earlier portion of the conversation.

Analysis:
Let me analyze the conversation chronologically:

1. The user invoked `/review` command without a PR number
2. I listed open PRs with `gh pr list`
3. The user specified PR #474
4. I fetched PR details and diff for PR #474 "Consolidate E2E test suite into cli repo"
5. I read through the entire diff (296KB, ~8600 lines) in multiple ch...

### Prompt 4

I guess we should default to the local directory's go build, and have the E2E_ENTIRE_BIN to override with a custom build?

