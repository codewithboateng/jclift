package rules

import (
	"regexp"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var reEven = regexp.MustCompile(`(?i)\bCOND\s*=\s*EVEN\b`)
var reOnly = regexp.MustCompile(`(?i)\bCOND\s*=\s*ONLY\b`)

func init() {
	Register(Rule{
		ID:      "EXEC-COND-FIRSTSTEP-MISUSE",
		Summary: "COND=EVEN/ONLY on first step is likely pointless or misleading.",
		Eval:    evalCondFirstStep,
	})
}

func evalCondFirstStep(job *ir.Job) []ir.Finding {
	if len(job.Steps) == 0 {
		return nil
	}
	st := job.Steps[0]
	u := strings.ToUpper(st.Conditions)
	if reEven.MatchString(u) || reOnly.MatchString(u) {
		return []ir.Finding{{
			RuleID:   "EXEC-COND-FIRSTSTEP-MISUSE",
			Type:     "RISK",
			Severity: "LOW",
			Job:      job.Name,
			Step:     st.Name,
			Message:  "First step uses COND=EVEN/ONLY; there is no prior RC to branch on.",
			Evidence: st.Conditions,
		}}
	}
	return nil
}
