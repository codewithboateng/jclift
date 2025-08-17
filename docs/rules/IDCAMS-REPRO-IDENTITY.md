---

### `docs/rules/IDCAMS-REPRO-IDENTITY.md`
```markdown
---

id: IDCAMS-REPRO-IDENTITY
type: COST
default_severity: LOW
since: mvp
docs_version: 1
summary: IDCAMS REPRO appears to copy without filtering.

---

# IDCAMS-REPRO-IDENTITY

## Why it matters

`REPRO` without `INCLUDE/EXCLUDE/KEYS/FROMKEY/TOKEY` is typically a pure copy—CPU/I-O cost with limited value.

## When it triggers

- Program = `IDCAMS`
- `SYSIN` contains `REPRO` and any of `INFILE/OUTFILE/INDATASET/OUTDATASET`
- No selection clauses (`INCLUDE`, `EXCLUDE`, `FROMKEY`, `TOKEY`, `KEYS`)

Detector: `internal/rules/idcams_repro_identity.go`

## Examples

**Flagged**

```jcl
//S1 EXEC PGM=IDCAMS
//SYSIN DD *
  REPRO INDATASET(IN.VSAM) OUTDATASET(OUT.VSAM)
/*
```
