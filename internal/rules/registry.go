package rules

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var registry []Rule

func Register(r Rule) { registry = append(registry, r) }

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
	for i := range run.Jobs {
		job := &run.Jobs[i]
		for _, rule := range rs {
			fs := rule.Eval(job)
			for k := range fs {
				if fs[k].ID == "" {
					fs[k].ID = makeID(rule.ID, job.Name, fs[k].Step, fs[k].Evidence, k)
				}
				if fs[k].Job == "" {
					fs[k].Job = job.Name
				}
				// drop below-threshold findings
				if !severityOK(fs[k].Severity) {
					continue
				}
				if fs[k].SavingsUSD == 0 && fs[k].SavingsMIPS > 0 && run.Context.MIPSToUSD > 0 {
					fs[k].SavingsUSD = fs[k].SavingsMIPS * run.Context.MIPSToUSD
				}
				all = append(all, fs[k])
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Severity == all[j].Severity {
			return all[i].ID < all[j].ID
		}
		return severityRank(all[i].Severity) > severityRank(all[j].Severity)
	})
	return all
}

func makeID(ruleID, job, step, evidence string, idx int) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d", ruleID, job, step, evidence, idx)
	sum := crc32.ChecksumIEEE([]byte(data))
	return fmt.Sprintf("%s-%08x", ruleID, sum)
}
