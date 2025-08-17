# ADR-0002: Intermediate Representation (IR)

Date: 2025-08-17
Status: Accepted

## Context

Rules, cost, and reporting need a consistent model for jobs/steps/DD/findings.

## Decision

Define a stable **IR** in Go (`internal/ir`) and serialize to JSON. Version = `ir.Version`.

## Consequences

- ✔ Deterministic outputs (golden tests)
- ✔ Easier interop with UI/APIs
- ✖ Schema evolution must be backward-compatible
