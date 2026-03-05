# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Switch Kiro CI auth from SIGV4 to device-flow token

## Context

SIGV4 auth was never ported to `kiro-cli-chat` — the binary ignores `AMAZON_Q_SIGV4` entirely (upstream issues: kirodotdev/Kiro#5938, aws/amazon-q-developer-cli#1051 both open). The workaround is device-flow auth: run `kiro-cli login --license free --use-device-flow` locally, then store the resulting token JSON as a GitHub secret (`KIRO_AUTH_TOKEN`). In CI, write it to `~/.aws/sso/cache/kiro-auth...

### Prompt 2

KIRO_AUTH_TOKEN is the access token or the full json?

### Prompt 3

commit and push

