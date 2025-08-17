---

### `docs/rules/GDG-ROLLOFF-RISK.md`
```markdown
---

id: GDG-ROLLOFF-RISK
type: RISK
default_severity: LOW
since: mvp
docs_version: 1
summary: Reading GDG(-1) while writing GDG(0) within the same step risks roll-off surprises.

---

# GDG-ROLLOFF-RISK

## Why it matters

Concurrent references to GDG bases in the same step can cause catalog sequencing surprises (e.g., roll-off timing vs. read consistency).

## When it triggers

- A step references `BASE.GDG(-1)` and `BASE.GDG(0)` (or other mixed generations) within the same step.

Detector: `internal/rules/gdg_rolloff_risk.go`

## Examples

**Flagged**

```jcl
//IN1  DD DSN=BASE.FILE.GDG(-1),DISP=SHR
//OUT1 DD DSN=BASE.FILE.GDG(0),DISP=(NEW,CATLG)
```
