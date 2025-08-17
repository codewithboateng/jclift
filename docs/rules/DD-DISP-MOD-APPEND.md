---

### `docs/rules/DD-DISP-MOD-APPEND.md`
```markdown
---

id: DD-DISP-MOD-APPEND
type: RISK
default_severity: LOW
since: mvp
docs_version: 1
summary: DISP=MOD appends to an existing dataset; can cause growth and serialization.

---

# DD-DISP-MOD-APPEND

## Why it matters

Appending to data sets via `DISP=MOD` may grow files unexpectedly and serialize writers, impacting batch windows.

## When it triggers

- `DD` has `DISP=MOD`

Detector: `internal/rules/dd_disp_mod_append.go`

## Examples

**Flagged**

```jcl
//MOD1 DD DSN=SOME.DATA,DISP=MOD
```
