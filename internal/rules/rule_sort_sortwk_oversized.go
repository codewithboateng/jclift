package rules

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

// Match CYL(900,...) or CYL,(900,...) and TRK variants.
// m[1] = unit ("CYL" or "TRK"), m[2] = primary number ("900")
var cylRe = regexp.MustCompile(`\b(CYL|TRK)\s*,?\s*\(\s*(\d+)`)

func init() {
	Register(Rule{
		ID:              "SORT-SORTWK-OVERSIZED",
		Summary:         "SORTWK work space appears oversized; consider tuning.",
		Type:            "COST",
		DefaultSeverity: "LOW",
		Docs:            "docs/rules/SORT-SORTWK-OVERSIZED.md",
		Eval:            evalSortwkOversized,
	})
}


func evalSortwkOversized(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		if !strings.EqualFold(st.Program, "SORT") {
			continue
		}
		overs := 0
		var evParts []string
		for _, dd := range st.DD {
			dn := strings.ToUpper(dd.DDName)
			if strings.HasPrefix(dn, "SORTWK") {
				s := strings.ToUpper(dd.Space)
				if m := cylRe.FindStringSubmatch(s); len(m) >= 3 {
					// m[1] is unit, m[2] is the primary number
					primary, _ := strconv.Atoi(m[2])
					if primary > rsettings.SortwkPrimaryCylThreshold {
						overs++
						evParts = append(evParts, dn+" SPACE="+dd.Space)
					}
				}
			}
		}
		if overs > 0 {
			out = append(out, ir.Finding{
				RuleID:      "SORT-SORTWK-OVERSIZED",
				Type:        "COST",
				Severity:    "LOW",
				Job:         job.Name,
				Step:        st.Name,
				Message:     "SORTWK primary cylinders exceed recommended thresholds; potential I/O/CPU waste.",
				Evidence:    strings.Join(evParts, " | "),
				SavingsMIPS: 0.8 * float64(overs), // placeholder heuristic
			})
		}
	}
	return out
}
