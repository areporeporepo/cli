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

I believe what we are trying to guard against is reaching outside the agent interface and building to internals that may break later. is it possible to institute more formal checks (linting?) to enforce this instead of relying on an agent instruction?

### Prompt 3

can we try 3 out?

### Prompt 4

let's put this into a different branch

### Prompt 5

okay, so this is what I'd like to do:

go back to the other agent skill PR, fix the typo and then approve/merge it

then we can come back to this one, update to main and relax the agent instructions - does that sound like a plan?

### Prompt 6

and we are behind the remote now...

### Prompt 7

I've done it on my end. let's switch back across now

### Prompt 8

oh, it was just waiting to merge in the CI pipeline - let's update again now please

### Prompt 9

yes please

