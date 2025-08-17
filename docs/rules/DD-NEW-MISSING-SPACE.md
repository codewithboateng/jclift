---

### `docs/rules/DD-NEW-MISSING-SPACE.md`
```markdown
---

id: DD-NEW-MISSING-SPACE
type: RISK
default_severity: MEDIUM
since: mvp
docs_version: 1
summary: NEW dataset allocation without explicit SPACE.

---

# DD-NEW-MISSING-SPACE

## Why it matters

Allocating a dataset with `DISP=NEW` but no `SPACE=` leaves sizing to defaults, causing extents, allocation failures, or waste.

## When it triggers

- `DD` has `DISP=(NEW,...)` or `DISP=NEW`
- No `SPACE=` parameter on the same `DD`

Detector: `internal/rules/dd_new_missing_space.go`

## Examples

**Flagged**

```jcl
//OUT DD DSN=OUTPUT.FILE,DISP=(NEW,CATLG,DELETE)
```
