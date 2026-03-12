# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Add `RecordError()` to Perf Span API

## Context

The perf framework currently has no way to indicate that a span ended due to an error. Users see lower-than-expected latencies in perf logs without knowing a step failed early. Following OpenTelemetry's pattern of separating error recording from span completion, we add a `RecordError(err)` method that marks a span as errored. The root span's log output then includes boolean error flags for any errored span.

**...

