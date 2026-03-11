# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix: TestDirtyWorkingTree cursor-cli ENOENT race

## Context

CI run `22809907314` (on commit `3373091e` which already has the pre-seed fix) still failed `TestDirtyWorkingTree` with cursor-cli. The error:

```
Error: ENOENT: no such file or directory, rename '/home/runner/.config/cursor/cli-config.json.tmp' -> '/home/runner/.config/cursor/cli-config.json'
```

**Root cause:** Pre-seeding `cli-config.json` with `{}` prevents the *initial creation* race, but not...

### Prompt 2

commit and push

