# ADR-0009: Waivers & Policy

Date: 2025-08-17
Status: Accepted

## Context

Teams need controlled exceptions without deleting rules.

## Decision

Introduce waivers stored in DB:

- Scope: by rule, optionally filter by job/step/pattern
- Attributes: reason, expires_at, created_by, revoked flags
- Apply during evaluation; track `Context.WaivedCount`

## Consequences

- ✔ Governance without code change
- ✖ Requires audit trail & admin role
