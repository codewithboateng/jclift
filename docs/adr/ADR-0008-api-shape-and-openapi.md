# ADR-0008: API Shape & OpenAPI

Date: 2025-08-17
Status: Accepted

## Context

UI and external tooling need a stable API contract.

## Decision

Model the HTTP API as read-only resources:

- `/runs`, `/runs/{id}`, `/runs/{id}/findings`, `/rules`, `/rules/meta`, `/waivers`, `/me`, `/auth/*`
  Define `docs/api/openapi.yaml` and publish Swagger locally.

## Consequences

- ✔ Tooling can generate clients
- ✖ Must keep spec in sync with code
