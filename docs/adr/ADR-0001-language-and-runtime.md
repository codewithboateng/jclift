# ADR-0001: Language and Runtime

Date: 2025-08-17
Status: Accepted

## Context

We need a fast, static binary with good concurrency and simple deployment.

## Decision

Use **Go** for CLI & API.

## Consequences

- ✔ Single static binary, easy ops
- ✔ Strong stdlib (HTTP, sqlite)
- ✖ Some parsing libs less mature than in JVM
- Mitigation: keep parser small and well-tested
