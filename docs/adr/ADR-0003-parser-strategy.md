# ADR-0003: Parser Strategy

Date: 2025-08-17
Status: Accepted

## Context

Full JCL grammar is large; MVP must ship quickly.

## Decision

Implement a **lightweight lexer/normalizer** that extracts JOB/EXEC/DD, DD attributes, and SYSIN content; tolerate unknowns.

## Consequences

- ✔ Fast MVP with usable signals
- ✖ Edge cases may be missed
- Mitigation: fuzz tests + SME feedback, iterate parser depth
