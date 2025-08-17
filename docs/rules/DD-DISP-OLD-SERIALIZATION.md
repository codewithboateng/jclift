---

### `docs/rules/DD-DISP-OLD-SERIALIZATION.md`
```markdown
---

id: DD-DISP-OLD-SERIALIZATION
type: RISK
default_severity: LOW
since: mvp
docs_version: 1
summary: DISP=OLD enforces exclusive access; may block concurrency.

---

# DD-DISP-OLD-SERIALIZATION

## Why it matters

`DISP=OLD` serializes access to a dataset. Unnecessary exclusivity reduces throughput and increases contention.

## When it triggers

- Any `DD` contains `DISP=OLD`

Detector: `internal/rules/dd_disp_old_serialization.go`

## Examples

**Flagged**

```jcl
//X1 DD DSN=SHARED.DATA.SET,DISP=OLD
```
