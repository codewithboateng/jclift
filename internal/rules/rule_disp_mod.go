package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "DD-DISP-MOD-APPEND",
		Summary: "DD uses DISP=MOD (append); verify it’s intentional.",
		Eval:    evalDispMod,
	})
}

func evalDispMod(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		for _, dd := range st.DD {
			up := strings.ToUpper(dd.DISP)
			if strings.Contains(up, "MOD") {
				out = append(out, ir.Finding{
					RuleID:   "DD-DISP-MOD-APPEND",
					Type:     "RISK",
					Severity: "LOW",
					Job:      job.Name,
					Step:     st.Name,
					Message:  "DISP=MOD appends to dataset; can cause unexpected growth and serialization.",
					Evidence: dd.DDName + " DISP=" + dd.DISP,
				})
			}
		}
	}
	return out
}
