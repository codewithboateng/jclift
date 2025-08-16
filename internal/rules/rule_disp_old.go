package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "DD-DISP-OLD-SERIALIZATION",
		Summary: "DISP=OLD can over-serialize dataset usage; verify if needed.",
		Eval:    evalDispOld,
	})
}

func evalDispOld(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		for _, dd := range st.DD {
			if strings.Contains(strings.ToUpper(dd.DISP), "OLD") {
				out = append(out, ir.Finding{
					RuleID:   "DD-DISP-OLD-SERIALIZATION",
					Type:     "RISK",
					Severity: "LOW",
					Job:      job.Name,
					Step:     st.Name,
					Message:  "DD uses DISP=OLD which enforces exclusive access; consider DISP=SHR if safe to improve concurrency.",
					Evidence: dd.DDName + " DISP=" + dd.DISP,
				})
			}
		}
	}
	return out
}
