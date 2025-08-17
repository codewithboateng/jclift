# ADR-0005: Storage – SQLite First

Date: 2025-08-17
Status: Accepted

## Context

MVP favors portability and zero external deps.

## Decision

Use **SQLite** with a simple schema; add Postgres later for multi-user scale.

## Consequences

- ✔ Zero-config DB
- ✖ Concurrency limited
- Migration path: migrate schema to Postgres at v1.0
