package rules

import (
	"regexp"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var gdgRe = regexp.MustCompile(`\(([+-]?\d+)\)\s*$`)

func init() {
	Register(Rule{
		ID:      "GDG-ROLLOFF-RISK",
		Summary: "Job reads prior GDG generation and writes current; verify roll-off logic.",
		Eval:    evalGDGRollOff,
	})
}

func evalGDGRollOff(job *ir.Job) []ir.Finding {
	readMinus1 := false
	writeZero := false
	var readEv, writeEv []string

	for _, st := range job.Steps {
		for _, dd := range st.DD {
			ds := strings.ToUpper(dd.Dataset)
			if ds == "" {
				continue
			}
			m := gdgRe.FindStringSubmatch(ds)
			if len(m) == 2 {
				gen := m[1]
				if gen == "-1" {
					readMinus1 = true
					readEv = append(readEv, st.Name+"."+dd.DDName+"="+dd.Dataset)
				}
				if gen == "0" {
					// treat as write if DISP indicates NEW/CATLG or not explicitly SHR
					upDisp := strings.ToUpper(dd.DISP)
					if strings.Contains(upDisp, "NEW") || strings.Contains(upDisp, "CATLG") || upDisp == "" || strings.Contains(upDisp, "MOD") {
						writeZero = true
						writeEv = append(writeEv, st.Name+"."+dd.DDName+"="+dd.Dataset)
					}
				}
			}
		}
	}

	if readMinus1 && writeZero {
		return []ir.Finding{{
			RuleID:   "GDG-ROLLOFF-RISK",
			Type:     "RISK",
			Severity: "MEDIUM",
			Job:      job.Name,
			Message:  "Reads GDG(-1) and writes GDG(0) in same job; validate roll-off windows and restart behavior.",
			Evidence: "reads: " + strings.Join(readEv, ", ") + " | writes: " + strings.Join(writeEv, ", "),
		}}
	}
	return nil
}
