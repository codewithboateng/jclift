package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "DD-NEW-MISSING-SPACE",
		Summary: "NEW allocation without SPACE specified.",
		Eval:    evalDDNewMissingSpace,
	})
}

func evalDDNewMissingSpace(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		for _, dd := range st.DD {
			disp := strings.ToUpper(dd.DISP)
			if strings.Contains(disp, "NEW") && strings.TrimSpace(dd.Space) == "" {
				out = append(out, ir.Finding{
					RuleID:   "DD-NEW-MISSING-SPACE",
					Type:     "RISK",
					Severity: "LOW",
					Job:      job.Name,
					Step:     st.Name,
					Message:  "DD allocates NEW dataset without SPACE=…; verify SMS defaults or specify SPACE to avoid abends/waste.",
					Evidence: dd.DDName + " DISP=" + dd.DISP,
				})
			}
		}
	}
	return out
}
