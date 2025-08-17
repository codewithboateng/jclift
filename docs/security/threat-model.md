# Threat Model

_Last updated: 2025-08-17_

## Assets

- JCL source (may contain environment names and DSNs)
- Findings & cost projections (sensitive spend signals)
- User accounts & sessions
- Waiver records (governance data)
- SQLite database file

## Actors

- Authenticated users (admin/viewer)
- Unauthorized users (Internet/LAN)
- Insider with local FS access

## Entry Points

- HTTP API `serve` (default :8080)
- Local filesystem (samples, reports, DB)

## Risks & Mitigations

1. **Unauthorized access to runs**
   - Mitigation: cookie auth for non-public endpoints; CORS control
2. **Session theft**
   - Mitigation: HttpOnly cookie; SameSite=Lax; configurable CORS
3. **Sensitive data exfiltration**
   - Mitigation: run air-gapped; limit datasets in samples; encrypt at rest (ops policy)
4. **Elevation via waiver APIs**
   - Mitigation: role-based checks; audit log on create/revoke
5. **Code injection in reports**
   - Mitigation: HTML escapes; serve reports statically (no script)

## Assumptions

- Reverse proxy/TLS termination handled by deployment
- No PII; only technical metadata & cost estimates
