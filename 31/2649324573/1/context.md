# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix Droid CLI Args Passed as Prompt

## Context

Droid E2E tests intermittently time out because `StartSession` invokes droid with `exec`-only flags in interactive mode:

```go
// Current (broken) — e2e/agents/droid.go:190
NewTmuxSession(name, dir, nil, "env", "ENTIRE_TEST_TTY=0", d.Binary(), "--model", defaultDroidModel, "--skip-permissions-unsafe")
```

From `droid --help`, interactive mode only accepts `-v`, `-r`, `-h`. Everything else becomes part of `[pro...

### Prompt 2

does this handle setting the model to the custom one for the tmux mode?

### Prompt 3

still failing in ci [e2e/tests]········································✖✖··✖✖····················✖✖··✖✖✖✖
=== Failed
=== FAIL: e2e/tests TestInteractiveMultiStep/factoryai-droid (155.53s)
    interactive_test.go:17: start session: droid failed to start after 3 attempts: timed out waiting for "⛬|│ >" after 30s
        --- pane content ---
        
        
                    █████████    █████████     ████████    ███   █████████
                    ███    ███   ███    ███   ███    ███   ███  ...

### Prompt 4

[Request interrupted by user for tool use]

