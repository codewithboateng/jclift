package rules

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/storage"
)

// ApplyWaivers filters out findings that match any active waiver.
// Returns (kept, waivedCount)
func ApplyWaivers(in []ir.Finding, waivers []storage.Waiver) ([]ir.Finding, int) {
	if len(waivers) == 0 || len(in) == 0 {
		return in, 0
	}
	var out []ir.Finding
	waived := 0
nextFinding:
	for _, f := range in {
		for _, w := range waivers {
			if !eqCI(f.RuleID, w.RuleID) { continue }
			if w.Job != ""  && !eqCI(f.Job,  w.Job)  { continue }
			if w.Step != "" && !eqCI(f.Step, w.Step) { continue }
			if w.PatternSub != "" {
				ps := strings.ToUpper(w.PatternSub)
				if !strings.Contains(strings.ToUpper(f.Evidence), ps) &&
				   !strings.Contains(strings.ToUpper(f.Message),  ps) {
					continue
				}
			}
			// matched → waive it
			waived++
			continue nextFinding
		}
		out = append(out, f)
	}
	return out, waived
}

func eqCI(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }
