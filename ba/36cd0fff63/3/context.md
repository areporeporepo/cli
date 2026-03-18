# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Fix Kiro CI auth: inject SQLite auth DB instead of SSO cache file

## Context

The previous commit wrote a token to `~/.aws/sso/cache/kiro-auth-token.json`, but `kiro-cli-chat` does **not** read from that path. Investigation shows:

- `kiro-cli-chat` stores auth in a **SQLite database** at a platform-specific path:
  - macOS: `~/Library/Application Support/kiro-cli/data.sqlite3`
  - Linux (CI): `~/.local/share/kiro-cli/data.sqlite3`
- The `auth_kv` table has t...

### Prompt 2

can I test this works locally? can you write a script that does what the gh action does to test that auth works?

### Prompt 3

[Request interrupted by user for tool use]

