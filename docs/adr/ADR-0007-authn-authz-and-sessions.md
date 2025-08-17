# ADR-0007: AuthN/AuthZ & Sessions

Date: 2025-08-17
Status: Accepted

## Context

We need minimal protection for waivers and private runs.

## Decision

- Local users (admin/viewer) in DB
- Login via POST; HttpOnly cookie `jclift_session`
- `GET /api/v1/me`, protected endpoints require cookie

## Consequences

- ✔ Simple to run air-gapped
- ✖ Not enterprise SSO
- Roadmap: OIDC/SAML in v1.x
