# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Plan: Replace combinatorial wrapper types with CapabilityDeclarer pattern

## Context

The `Wrap()` function in `external/capabilities.go` manually defines wrapper types for each combination of capabilities an external agent can declare. With 7 capability flags, this doesn't scale (128 possible combinations; currently only 9 are covered). The fix: introduce a `CapabilityDeclarer` interface and `As*` helper functions so a single wrapper type can implement all i...

