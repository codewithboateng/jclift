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
		if !strings.EqualFold(st.Program, "IEBGENER") {
			continue
		}
		var (
			sysin          string
			haveIn, haveOut bool
		)
		for _, dd := range st.DD {
			switch {
			case strings.EqualFold(dd.DDName, "SYSIN"):
				sysin = strings.TrimSpace(dd.Content)
			case strings.EqualFold(dd.DDName, "SYSUT1"):
				haveIn = true
			case strings.EqualFold(dd.DDName, "SYSUT2"):
				haveOut = true
			}
		}
		if (sysin == "" || strings.EqualFold(sysin, "DUMMY")) && haveIn && haveOut {
			savings := st.Annotations.Cost.MIPS
			if savings <= 0 {
				savings = 0.5 // safe fallback if cost wasn’t computed
			}
			out = append(out, ir.Finding{
				RuleID:      "IEBGENER-REDUNDANT-COPY",
				Type:        "COST",
				Severity:    "LOW",
				Job:         job.Name,
				Step:        st.Name,
				Message:     "IEBGENER appears to copy the full dataset without filtering; consider inlining or eliminating redundant copies.",
				Evidence:    "SYSIN=DUMMY or empty; SYSUT1→SYSUT2 full copy.",
				SavingsMIPS: savings, // USD auto-filled by rules.Evaluate using MIPS→USD
			})
		}
	}
	return out
}
