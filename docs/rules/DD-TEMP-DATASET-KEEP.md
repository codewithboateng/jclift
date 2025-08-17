---

### `docs/rules/DD-TEMP-DATASET-KEEP.md`
```markdown
---

id: DD-TEMP-DATASET-KEEP
type: RISK
default_severity: LOW
since: mvp
docs_version: 1
summary: Temporary dataset (&&name) is kept/cataloged; may leak storage.

---

# DD-TEMP-DATASET-KEEP

## Why it matters

Temporary datasets (`&&TMP`) are expected to vanish. Keeping/cataloging them causes catalog clutter and storage leaks.

## When it triggers

- `DD` uses a temp dataset name (`DSN=&&...`)
- `DISP=(NEW,CATLG)` or equivalent “keep” disposition

Detector: `internal/rules/dd_temp_dataset_keep.go`

## Examples

**Flagged**

```jcl
//TMP1 DD DSN=&&TMP1,DISP=(NEW,CATLG)
```
