---

### `docs/rules/IEBGENER-REDUNDANT-COPY.md`
```markdown
---

id: IEBGENER-REDUNDANT-COPY
type: COST
default_severity: LOW
since: mvp
docs_version: 1
summary: IEBGENER copies entire dataset without filtering; may be redundant.

---

# IEBGENER-REDUNDANT-COPY

## Why it matters

`IEBGENER` full-copy steps are often vestiges; they add elapsed time and MIPS without business value.

## When it triggers

- Program = `IEBGENER`
- `SYSIN` is `DUMMY` or empty
- Both `SYSUT1` and `SYSUT2` present (full copy)

Detector: `internal/rules/iebgener_redundant.go`

## Examples

**Flagged**

```jcl
//S2   EXEC PGM=IEBGENER
//SYSUT1 DD DSN=INPUT.FILE,DISP=SHR
//SYSUT2 DD DSN=OUTPUT.FILE,DISP=(NEW,CATLG,DELETE)
//SYSIN  DD DUMMY
```
