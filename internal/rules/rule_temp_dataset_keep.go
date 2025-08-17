package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "DD-TEMP-DATASET-KEEP",
		Summary: "Temporary dataset (&&) is kept/cataloged; potential leakage.",
		Eval:    evalTempKeep,
	})
}

func evalTempKeep(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		for _, dd := range st.DD {
			ds := strings.TrimSpace(dd.Dataset)
			if strings.HasPrefix(ds, "&&") {
				up := strings.ToUpper(dd.DISP)
				if strings.Contains(up, "KEEP") || strings.Contains(up, "CATLG") {
					out = append(out, ir.Finding{
						RuleID:   "DD-TEMP-DATASET-KEEP",
						Type:     "RISK",
						Severity: "LOW",
						Job:      job.Name,
						Step:     st.Name,
						Message:  "Temporary dataset (&&name) marked KEEP/CATLG; verify lifecycle to avoid catalog clutter/leaks.",
						Evidence: dd.DDName + " " + dd.DISP,
					})
				}
			}
		}
	}
	return out
}
