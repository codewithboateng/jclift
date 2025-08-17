package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

type diffPayload struct {
	BaseID  string         `json:"base_id"`
	HeadID  string         `json:"head_id"`
	Summary diffSummary    `json:"summary"`
	New     []diffFinding  `json:"new"`
	Removed []diffFinding  `json:"removed"`
	Changed []diffChanged  `json:"changed"`
}

type diffSummary struct {
	NewCount     int `json:"new"`
	RemovedCount int `json:"removed"`
	ChangedCount int `json:"changed"`
}

type diffFinding struct {
	RuleID   string  `json:"rule_id"`
	Job      string  `json:"job"`
	Step     string  `json:"step,omitempty"`
	Severity string  `json:"severity,omitempty"`
	Message  string  `json:"message,omitempty"`
	SavMIPS  float64 `json:"savings_mips,omitempty"`
}

type diffChanged struct {
	Key     string      `json:"key"`
	Base    diffFinding `json:"base"`
	Head    diffFinding `json:"head"`
	Changed []string    `json:"fields_changed"`
}

func WriteDiffJSON(baseID, headID, outDir string, base, head *ir.Run) (string, error) {
	path := filepath.Join(outDir, "diff_"+baseID+"__"+headID+".json")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}

	// index findings
	bm := map[string]ir.Finding{}
	hm := map[string]ir.Finding{}
	for _, f := range base.Findings {
		bm[keyOf(f)] = f
	}
	for _, f := range head.Findings {
		hm[keyOf(f)] = f
	}

	var added []diffFinding
	var removed []diffFinding
	var changed []diffChanged

	// additions & changes
	for k, hf := range hm {
		if bf, ok := bm[k]; !ok {
			added = append(added, asDiff(hf))
		} else {
			var fields []string
			if norm(bf.Severity) != norm(hf.Severity) {
				fields = append(fields, "severity")
			}
			if strings.TrimSpace(bf.Message) != strings.TrimSpace(hf.Message) {
				fields = append(fields, "message")
			}
			if bf.SavingsMIPS != hf.SavingsMIPS {
				fields = append(fields, "savings_mips")
			}
			if len(fields) > 0 {
				changed = append(changed, diffChanged{
					Key:     k,
					Base:    asDiff(bf),
					Head:    asDiff(hf),
					Changed: fields,
				})
			}
		}
	}
	// removals
	for k, bf := range bm {
		if _, ok := hm[k]; !ok {
			removed = append(removed, asDiff(bf))
		}
	}

	// stable sort
	sort.Slice(added, func(i, j int) bool { return added[i].RuleID < added[j].RuleID })
	sort.Slice(removed, func(i, j int) bool { return removed[i].RuleID < removed[j].RuleID })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Key < changed[j].Key })

	payload := diffPayload{
		BaseID: baseID, HeadID: headID,
		Summary: diffSummary{
			NewCount:     len(added),
			RemovedCount: len(removed),
			ChangedCount: len(changed),
		},
		New:     added,
		Removed: removed,
		Changed: changed,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, b, 0o644)
}

func keyOf(f ir.Finding) string {
	sb := strings.Builder{}
	sb.WriteString(norm(f.RuleID)); sb.WriteByte('|')
	sb.WriteString(norm(f.Job)); sb.WriteByte('|')
	sb.WriteString(norm(f.Step)); sb.WriteByte('|')
	// evidence drives logical identity for many rules
	sb.WriteString(norm(f.Evidence))
	return sb.String()
}

func asDiff(f ir.Finding) diffFinding {
	return diffFinding{
		RuleID:   f.RuleID,
		Job:      f.Job,
		Step:     f.Step,
		Severity: f.Severity,
		Message:  f.Message,
		SavMIPS:  f.SavingsMIPS,
	}
}

func norm(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
