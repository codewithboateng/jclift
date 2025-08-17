package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codewithboateng/jclift/internal/cost"
	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/parser"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/shared"
)


func analyzeStrings(t *testing.T, files map[string]string, severity string) ir.Run {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	run, _ := parser.Parse(dir)

	cfg := shared.DefaultConfig()

	// rules setup
	disabled := map[string]bool{}
	rules.SetSettings(rules.Settings{
		SeverityThreshold:         strings.ToUpper(severity),
		Disabled:                  disabled,
		SortwkPrimaryCylThreshold: cfg.Rules.Sortwk.PrimaryCylThreshold,
	})

	// context: cost model + geometry + price
	run.Context.MIPSToUSD = 250
	run.Context.Geometry.TracksPerCyl = cfg.Cost.Geometry.TracksPerCyl
	run.Context.Geometry.BytesPerTrack = cfg.Cost.Geometry.BytesPerTrack
	run.Context.Model.MIPSPerCPU = cfg.Cost.Model.MIPSPerCPU
	run.Context.Model.SortAlpha = cfg.Cost.Model.Sort.Alpha
	run.Context.Model.SortBeta = cfg.Cost.Model.Sort.Beta
	run.Context.Model.CopyAlpha = cfg.Cost.Model.Copy.Alpha
	run.Context.Model.CopyBeta = cfg.Cost.Model.Copy.Beta
	run.Context.Model.IDAlpha = cfg.Cost.Model.IDCAMS.Alpha
	run.Context.Model.IDBeta = cfg.Cost.Model.IDCAMS.Beta

	// cost annotate
	for i := range run.Jobs {
		for j := range run.Jobs[i].Steps {
			size := cost.EstimateSizeMB(&run.Jobs[i].Steps[j], run.Context.Geometry)
			run.Jobs[i].Steps[j].Annotations.SizeMB = size
			run.Jobs[i].Steps[j].Annotations.Cost = cost.Estimate(&run.Jobs[i].Steps[j], run.Context)
		}
	}

	run.Findings = rules.Evaluate(&run)
	return run
}

func TestSample_LowSeverity_ContainsKeyFindings(t *testing.T) {
	run := analyzeStrings(t, map[string]string{"payroll.jcl": samplePayroll}, "LOW")

	counts := map[string]int{}
	withUSD := 0
	for _, f := range run.Findings {
		counts[f.RuleID]++
		if f.SavingsUSD > 0 {
			withUSD++
		}
	}

	// Presence checks for the MVP rules on our sample
	required := []string{
		"SORT-IDENTITY",
		"IEBGENER-REDUNDANT-COPY",
		"DD-DISP-OLD-SERIALIZATION",
		"DD-DUPLICATE-DATASET",
		"SORT-SORTWK-OVERSIZED",
		"DD-NEW-MISSING-SPACE",
	}
	for _, id := range required {
		if counts[id] == 0 {
			t.Fatalf("expected at least 1 finding for %s; got 0; counts=%v", id, counts)
		}
	}
	if withUSD == 0 {
		t.Fatalf("expected at least one finding with SavingsUSD>0 (MIPS→USD projection); got none")
	}
}

func TestSample_MediumSeverity_FiltersLowFindings(t *testing.T) {
	runLow := analyzeStrings(t, map[string]string{"payroll.jcl": samplePayroll}, "LOW")
	runMed := analyzeStrings(t, map[string]string{"payroll.jcl": samplePayroll}, "MEDIUM")

	if len(runMed.Findings) >= len(runLow.Findings) {
		t.Fatalf("expected MEDIUM to have fewer findings than LOW; got MEDIUM=%d LOW=%d",
			len(runMed.Findings), len(runLow.Findings))
	}
	// SORT-IDENTITY is MEDIUM → should remain
	found := false
	for _, f := range runMed.Findings {
		if f.RuleID == "SORT-IDENTITY" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected SORT-IDENTITY to remain at MEDIUM threshold")
	}
}
