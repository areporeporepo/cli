# Session Context

## User Prompts

### Prompt 1

Implement the following plan:

# Enable BM25 Full-Text Search + Hybrid Search in Turbopuffer

## Context

Checkpoint search is currently semantic-only (OpenAI embeddings → Turbopuffer cosine similarity). There's a TODO in `turbopuffer.ts:4` to add BM25 full-text search. BM25 excels at exact keyword matches (file names, function names, error messages) where semantic search can miss. Adding both and combining via rank fusion gives the best results.

Turbopuffer supports BM25 natively via a `sch...

### Prompt 2

do we need to re-run the backfill or is it just going to work as-is?

### Prompt 3

okay, we can do the re-index. in that case, while we're at it. let's set up the necessary plumbing to use the right namespace instead of the hardcoded one right now.

### Prompt 4

[Request interrupted by user for tool use]

