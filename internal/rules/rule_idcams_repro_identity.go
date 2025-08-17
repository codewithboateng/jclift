package rules

import (
	"regexp"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var reRepro    = regexp.MustCompile(`(?i)\bREPRO\b`)
var reSelect   = regexp.MustCompile(`(?i)\b(INCLUDE|EXCLUDE|FROMKEY|TOKEY|KEYS)\b`)
var reHasFiles = regexp.MustCompile(`(?i)\b(INFILE|OUTFILE|INDATASET|OUTDATASET)\b`)

func init() {
	Register(Rule{
		ID:              "IDCAMS-REPRO-IDENTITY",
		Summary:         "IDCAMS REPRO appears to copy without filtering.",
		Type:            "COST",
		DefaultSeverity: "LOW",
		Docs:            "docs/rules/IDCAMS-REPRO-IDENTITY.md",
		Eval:            evalIDCAMSReproIdentity,
	})
}


func evalIDCAMSReproIdentity(job *ir.Job) []ir.Finding {
	var out []ir.Finding
	for _, st := range job.Steps {
		if !strings.EqualFold(st.Program, "IDCAMS") {
			continue
		}
		sysin := ""
		for _, dd := range st.DD {
			if strings.EqualFold(dd.DDName, "SYSIN") {
				sysin = dd.Content
				break
			}
		}
		u := strings.ToUpper(sysin)
		if reRepro.MatchString(u) && reHasFiles.MatchString(u) && !reSelect.MatchString(u) {
			savings := st.Annotations.Cost.MIPS
			if savings <= 0 {
				savings = 0.5 // conservative fallback if cost couldn’t be computed
			}
			out = append(out, ir.Finding{
				RuleID:      "IDCAMS-REPRO-IDENTITY",
				Type:        "COST",
				Severity:    "LOW",
				Job:         job.Name,
				Step:        st.Name,
				Message:     "IDCAMS REPRO without selection clauses; consider eliminating or consolidating redundant copies.",
				Evidence:    "SYSIN REPRO with IN/OUT and no INCLUDE/EXCLUDE/KEYS.",
				SavingsMIPS: savings, // USD filled in Evaluate()
			})
		}
	}
	return out
}
