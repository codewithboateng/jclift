package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "IEBGENER-REDUNDANT-COPY",
		Summary: "IEBGENER copies entire dataset without filtering; may be redundant.",
		Eval:    evalIEBGENERRedundant,
	})
}

func evalIEBGENERRedundant(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		if strings.EqualFold(st.Program, "IEBGENER") {
			// Heuristic: if SYSIN DUMMY or empty (full copy)
			sysin := ""
			haveIn, haveOut := false, false
			for _, dd := range st.DD {
				n := strings.ToUpper(dd.DDName)
				if n == "SYSIN" {
					sysin = strings.TrimSpace(strings.ToUpper(dd.Content))
				}
				if n == "SYSUT1" {
					haveIn = true
				}
				if n == "SYSUT2" {
					haveOut = true
				}
			}
			if (sysin == "" || sysin == "DUMMY") && haveIn && haveOut {
				out = append(out, ir.Finding{
					RuleID:      "IEBGENER-REDUNDANT-COPY",
					Type:        "COST",
					Severity:    "LOW",
					Step:        st.Name,
					Message:     "IEBGENER appears to copy the full dataset without filtering; consider inlining or eliminating redundant copies.",
					Evidence:    "SYSIN=DUMMY or empty; SYSUT1→SYSUT2 full copy.",
					SavingsMIPS: 0.5, // placeholder; refined by cost model later
				})
			}
		}
	}
	return out
}
