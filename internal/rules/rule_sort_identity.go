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
		// Minimal stub logic until parser is richer
		if strings.EqualFold(st.Program, "SORT") {
			// Try to find SYSIN control cards
			sysin := ""
			for _, dd := range st.DD {
				if strings.EqualFold(dd.DDName, "SYSIN") {
					sysin = dd.Content
					break
				}
			}
			// Simple heuristics
			if strings.Contains(strings.ToUpper(sysin), "FIELDS=COPY") ||
				(strings.TrimSpace(sysin) == "" && len(st.DD) > 0) {
				out = append(out, ir.Finding{
					RuleID:   "SORT-IDENTITY",
					Type:     "COST",
					Severity: "MEDIUM",
					Step:     st.Name,
					Message:  "SORT appears to perform an identity copy (no effective key). Consider removing or merging upstream.",
					Evidence: snippet(sysin),
					// A small placeholder savings until cost model/size is known.
					SavingsMIPS: 1.0,
				})
			}
		}
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
