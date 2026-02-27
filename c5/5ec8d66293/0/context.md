# Session Context

## User Prompts

### Prompt 1

Can you check PR 531 it's already merged but I'd like to especially focus on the changes in manual_hooks_commit.go and askConfirmTTY I'm trying to understand how the behavior changed to before the PR

### Prompt 2

can you review this again from the point of view of no TTY, like agent commits mitdurn?

### Prompt 3

but isn't the comment then on askConfirmTTY wrong?

### Prompt 4

I'm mostly meaning this comment:

// If TTY is unavailable, returns ttyConfirmYes when defaultYes is true, ttyConfirmNo otherwise.

### Prompt 5

oh wait and "ttyConfirmNo" is "don't ask for confirmation" ?

### Prompt 6

how can we improve this from a readability point of view, any suggestions?

### Prompt 7

yes, let's do 2 + 3

