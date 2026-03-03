# Session Context

## User Prompts

### Prompt 1

ok we're working on the pr some more.

better approach, only use in two situations:

pre-push before we merge trees
in entire doctor, with detection and pointers to run doctor

so we can remove it from the other places and just have it be done with what is invoked from the pre-push hook. it has to be done before we do the tree merge in the pre-push. 

then in the other places where a disconnected branch may cause unexpected behaviour, like missing checkpoints, briefly explain the situation to...

### Prompt 2

push it to the branch

