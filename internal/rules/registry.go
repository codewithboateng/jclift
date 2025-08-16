package rules

import (
	"fmt"
	"hash/crc32"
	"sort"

	"github.com/codewithboateng/jclift/internal/ir"
)

var registry []Rule

// Register adds a rule to the in-memory registry.
// Call from individual rule files' init() functions.
func Register(r Rule) { registry = append(registry, r) }

// List returns the currently registered rules (sorted by ID for determinism).
func List() []Rule {
	out := make([]Rule, len(registry))
	copy(out, registry)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Evaluate executes all registered rules over the run and returns the merged findings.
// It also normalizes fields (IDs, USD from MIPS where possible).
func Evaluate(run *ir.Run) []ir.Finding {
	var all []ir.Finding
	rules := List()
	for i := range run.Jobs {
		job := &run.Jobs[i]
		for _, rule := range rules {
			fs := rule.Eval(job)
			for k := range fs {
				// Deterministic, unique ID per finding payload
				if fs[k].ID == "" {
					fs[k].ID = makeID(rule.ID, job.Name, fs[k].Step, fs[k].Evidence, k)
				}
				// Attach job if missing
				if fs[k].Job == "" {
					fs[k].Job = job.Name
				}
				// Compute USD from MIPS, if configured
				if fs[k].SavingsUSD == 0 && fs[k].SavingsMIPS > 0 && run.Context.MIPSToUSD > 0 {
					fs[k].SavingsUSD = fs[k].SavingsMIPS * run.Context.MIPSToUSD
				}
			}
			all = append(all, fs...)
		}
	}
	// Stable order for reproducible outputs
	sort.Slice(all, func(i, j int) bool {
		if all[i].Severity == all[j].Severity {
			return all[i].ID < all[j].ID
		}
		// HIGH > MEDIUM > LOW
		sev := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
		return sev[all[i].Severity] > sev[all[j].Severity]
	})
	return all
}

func makeID(ruleID, job, step, evidence string, idx int) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d", ruleID, job, step, evidence, idx)
	sum := crc32.ChecksumIEEE([]byte(data))
	return fmt.Sprintf("%s-%08x", ruleID, sum)
}
