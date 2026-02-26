# Session Context

## User Prompts

### Prompt 1

can you take a look at mise.toml, we have a few commands which are basically shell scripts, I'd like to extract them out, but first: is there a linter for mise.toml that we can run to prevent this happening in the future?

### Prompt 2

can we add a script in mise-tasks/lint "mise" that does the multi line check, and then also add it to lint/_default?

### Prompt 3

can we now extract the multi line scripts into individual files in mise-tasks

