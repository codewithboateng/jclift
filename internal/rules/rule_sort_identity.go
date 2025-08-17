package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func init() {
	Register(Rule{
		ID:      "SORT-IDENTITY",
		Summary: "SORT step appears to be a no-op (FIELDS=COPY or no effective key).",
		Eval:    evalSortIdentity,
	})
}

func evalSortIdentity(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		if !strings.EqualFold(st.Program, "SORT") {
			continue
		}
		// Pull SYSIN control cards (if any)
		sysin := ""
		for _, dd := range st.DD {
			if strings.EqualFold(dd.DDName, "SYSIN") {
				sysin = dd.Content
				break
			}
		}
		up := strings.ToUpper(sysin)
		identity := strings.Contains(up, "FIELDS=COPY") || (strings.TrimSpace(sysin) == "" && len(st.DD) > 0)
		if !identity {
			continue
		}

		// Savings = step cost (MIPS) if available; otherwise a tiny fallback
		savings := st.Annotations.Cost.MIPS
		if savings <= 0 {
			savings = 0.8
		}

		ev := strings.TrimSpace(sysin)
		if ev == "" {
			ev = "(empty SYSIN)"
		}

		out = append(out, ir.Finding{
			RuleID:      "SORT-IDENTITY",
			Type:        "COST",
			Severity:    "MEDIUM",
			Job:         job.Name,
			Step:        st.Name,
			Message:     "SORT appears to perform an identity copy (no effective key). Consider removing or merging upstream.",
			Evidence:    snippet(ev),
			SavingsMIPS: savings, // USD filled by rules.Evaluate using MIPS→USD
		})
	}
	return out
}

func snippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}
