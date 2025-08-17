# ADR-0004: Cost Model v1 → v2

Date: 2025-08-17
Status: Accepted

## Context

We must estimate CPU/MIPS cost to rank findings.

## Decision

- **v1**: Heuristic model driven by DD space hints + geometry.
- **v2**: Calibrate with SMF/RMF per site.

## Consequences

- ✔ Immediate prioritization
- ✖ Approximations can be off
- Mitigation: expose coefficients; add calibration inputs in v2
