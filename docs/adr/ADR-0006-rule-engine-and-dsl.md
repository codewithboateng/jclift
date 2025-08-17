# ADR-0006: Rule Engine & DSL

Date: 2025-08-17
Status: Accepted

## Context

We need quick iteration on rules and documented guidance.

## Decision

- Built-in **Go rules** for core patterns.
- **YAML DSL** loader for additional rules; evaluated post-parse.
- Each rule has a doc at `docs/rules/<ID>.md`.

## Consequences

- ✔ Mix of performance (Go) + flexibility (DSL)
- ✖ Two paths to maintain
