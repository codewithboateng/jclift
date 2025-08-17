---

### `docs/rules/SORT-SORTWK-OVERSIZED.md`
```markdown
---

id: SORT-SORTWK-OVERSIZED
type: COST
default_severity: LOW
since: mvp
docs_version: 1
summary: SORTWK work space appears oversized; consider tuning.

---

# SORT-SORTWK-OVERSIZED

## Why it matters

Oversized `SORTWKnn` allocations inflate I/O and storage utilization and can reduce batch parallelism.

## When it triggers

- Program = `SORT`
- Any `DD` named `SORTWKnn`
- Parsed `SPACE=(CYL,(primary,...))` where `primary > rules.sortwk.primary_cyl_threshold`

Detector: `internal/rules/sort_sortwk_oversized.go`  
Threshold: `configs/jclift.yaml → rules.sortwk.primary_cyl_threshold` (default e.g., 500)

## Examples

**Flagged**

```jcl
//SORTWK01 DD UNIT=SYSDA,SPACE=(CYL,(900,50))
```
