package rules

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var (
	registry  []Rule
	ruleIndex = map[string]int{} // UPPER(ruleID) -> index
)

func Register(r Rule) {
	registry = append(registry, r)
	ruleIndex[strings.ToUpper(strings.TrimSpace(r.ID))] = len(registry) - 1
}

func List() []Rule {
	out := make([]Rule, 0, len(registry))
	for _, r := range registry {
		if rsettings.Disabled[strings.ToUpper(r.ID)] {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func Evaluate(run *ir.Run) []ir.Finding {
	var all []ir.Finding
	rs := List()

	seen := make(map[string]struct{}) // finding IDs seen in this run
	seq := 0

	put := func(id string) bool {
		if _, ok := seen[id]; ok {
			return false
		}
		seen[id] = struct{}{}
		return true
	}

	for i := range run.Jobs {
		job := &run.Jobs[i]
		for _, rule := range rs {
			fs := rule.Eval(job)
			for k := range fs {
				// Ensure Job is set
				if fs[k].Job == "" {
					fs[k].Job = job.Name
				}
				// Compute USD from MIPS if configured
				if fs[k].SavingsUSD == 0 && fs[k].SavingsMIPS > 0 && run.Context.MIPSToUSD > 0 {
					fs[k].SavingsUSD = fs[k].SavingsMIPS * run.Context.MIPSToUSD
				}
				// Guarantee unique ID within the run
				id := fs[k].ID
				if id == "" || !put(id) {
					// Assign a fresh, run-local unique id
					for {
						seq++
						candidate := fmt.Sprintf("%s-%06d", rule.ID, seq)
						if put(candidate) {
							id = candidate
							break
						}
					}
					fs[k].ID = id
				}
			}
			all = append(all, fs...)
		}
	}

	// Stable order for reproducible outputs
	sev := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Severity == all[j].Severity {
			return all[i].ID < all[j].ID
		}
		return sev[all[i].Severity] > sev[all[j].Severity]
	})
	return all
}

func makeID(ruleID, job, step, evidence string, idx int) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d", ruleID, job, step, evidence, idx)
	sum := crc32.ChecksumIEEE([]byte(data))
	return fmt.Sprintf("%s-%08x", ruleID, sum)
}

// Get returns a rule by ID if registered (used by HTML report to link docs).
func Get(id string) (Rule, bool) {
	idx, ok := ruleIndex[strings.ToUpper(strings.TrimSpace(id))]
	if !ok || idx < 0 || idx >= len(registry) {
		return Rule{}, false
	}
	return registry[idx], true
}

