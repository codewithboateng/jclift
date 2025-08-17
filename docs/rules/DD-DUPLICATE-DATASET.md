---

### `docs/rules/DD-DUPLICATE-DATASET.md`
```markdown
---

id: DD-DUPLICATE-DATASET
type: RISK
default_severity: LOW
since: mvp
docs_version: 1
summary: Same dataset referenced multiple times within a step.

---

# DD-DUPLICATE-DATASET

## Why it matters

Multiple `DD` statements pointing to the same dataset in one step can signal confusion, potential serialization, or accidental overwrite.

## When it triggers

- Within a single step, the same `DSN=` appears on multiple DDs.

Detector: `internal/rules/dd_duplicate_dataset.go`

## Examples

**Flagged**

```jcl
//X1 DD DSN=SHARED.DATA.SET,DISP=OLD
//X2 DD DSN=SHARED.DATA.SET,DISP=OLD
//X3 DD DSN=SHARED.DATA.SET,DISP=OLD
```
