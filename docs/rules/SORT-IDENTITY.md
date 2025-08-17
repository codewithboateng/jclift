---
id: SORT-IDENTITY
type: COST
default_severity: MEDIUM
since: mvp
docs_version: 1
summary: SORT step appears to be a no-op (FIELDS=COPY or no effective key).
---

# SORT-IDENTITY

## Why it matters

Invoking SORT when it only performs an identity copy burns CPU/MIPS and I/O without transforming data. This is common technical debt in legacy streams.

## When it triggers

- Program = `SORT`
- `SYSIN` contains `FIELDS=COPY`, **or**
- `SYSIN` is empty/whitespace (no effective sort key)

Detector: `internal/rules/sort_identity.go`

## Examples

**Flagged**

```jcl
//S1  EXEC PGM=SORT
//SYSIN DD *
  SORT FIELDS=COPY
/*
```
