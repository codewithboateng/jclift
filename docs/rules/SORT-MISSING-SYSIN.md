---

### `docs/rules/SORT-MISSING-SYSIN.md`
```markdown
---

id: SORT-MISSING-SYSIN
type: RISK
default_severity: MEDIUM
since: mvp
docs_version: 1
summary: SORT step missing SYSIN (control cards) may imply unintended defaults.

---

# SORT-MISSING-SYSIN

## Why it matters

Running SORT without explicit control cards can be ambiguous, leading to default behaviors that surprise operators and waste CPU.

## When it triggers

- Program = `SORT`
- No `SYSIN` DD present

Detector: `internal/rules/sort_missing_sysin.go`

## Examples

**Flagged**

```jcl
//S2 EXEC PGM=SORT
//SYSUT1 DD DSN=IN.FILE,DISP=SHR
//SYSUT2 DD DSN=OUT.FILE,DISP=(NEW,CATLG,DELETE)
//* (no SYSIN)
```
