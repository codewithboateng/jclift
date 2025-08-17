package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "SORT-MISSING-SYSIN",
		Summary: "SORT step missing SYSIN; intent unclear.",
		Eval:    evalSortMissingSYSIN,
	})
}

func evalSortMissingSYSIN(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		if !strings.EqualFold(st.Program, "SORT") {
			continue
		}
		found := false
		for _, dd := range st.DD {
			if strings.EqualFold(dd.DDName, "SYSIN") {
				found = true
				break
			}
		}
		if !found {
			out = append(out, ir.Finding{
				RuleID:   "SORT-MISSING-SYSIN",
				Type:     "RISK",
				Severity: "LOW",
				Job:      job.Name,
				Step:     st.Name,
				Message:  "SORT has no SYSIN; verify default behavior vs. intended transform.",
			})
		}
	}
	return out
}
