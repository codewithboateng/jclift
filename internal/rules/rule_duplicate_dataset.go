package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "DD-DUPLICATE-DATASET",
		Summary: "Multiple DDs reference the same dataset within a step; consider consolidation.",
		Eval:    evalDuplicateDataset,
	})
}

func evalDuplicateDataset(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		dsCounts := make(map[string]int)
		for _, dd := range st.DD {
			ds := strings.ToUpper(strings.TrimSpace(dd.Dataset))
			if ds == "" {
				continue
			}
			dsCounts[ds]++ // ← correct map indexing
		}

		// Evidence for datasets used more than once in this step
		var ev []string
		for ds, c := range dsCounts {
			if c > 1 {
				ev = append(ev, ds)
			}
		}
		if len(ev) > 0 {
			out = append(out, ir.Finding{
				RuleID:   "DD-DUPLICATE-DATASET",
				Type:     "RISK",
				Severity: "LOW",
				Job:      job.Name,
				Step:     st.Name,
				Message:  "Same dataset referenced multiple times within the step; verify necessity to avoid serialization or confusion.",
				Evidence: strings.Join(ev, ", "),
			})
		}
	}
	return out
}
