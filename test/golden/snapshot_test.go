package golden

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/codewithboateng/jclift/internal/cost"
	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/parser"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/shared"
)

var update = flag.Bool("update", false, "update golden snapshot")

const goldenFile = "test/golden/expected.json"

const samplePayroll = `//PAYROLL  JOB (12345),'DEMO RUN',CLASS=A,MSGCLASS=X,NOTIFY=&SYSUID
//S1       EXEC PGM=SORT
//SYSIN    DD *
  SORT  FIELDS=COPY
/*
//SORTWK01 DD UNIT=SYSDA,SPACE=(CYL,(900,50))
//SORTWK02 DD UNIT=SYSDA,SPACE=(CYL,(700,50))
//S2       EXEC PGM=IEBGENER
//SYSUT1   DD DSN=INPUT.FILE,DISP=SHR
//SYSUT2   DD DSN=OUTPUT.FILE,DISP=(NEW,CATLG,DELETE)
//SYSIN    DD DUMMY
//X1       DD DSN=SHARED.DATA.SET,DISP=OLD
//X2       DD DSN=SHARED.DATA.SET,DISP=OLD
//X3       DD DSN=SHARED.DATA.SET,DISP=OLD
`

func TestGolden_PayrollSnapshot(t *testing.T) {
	// Build a temp input dir
	dir := t.TempDir()
	in := filepath.Join(dir, "payroll.jcl")
	if err := os.WriteFile(in, []byte(samplePayroll), 0o644); err != nil {
		t.Fatalf("write sample: %v", err)
	}

	// Parse → Run
	run, _ := parser.Parse(dir)

	// Config defaults
	cfg := shared.DefaultConfig()

	// Rules settings
	rules.SetSettings(rules.Settings{
		SeverityThreshold:         "LOW",
		Disabled:                  map[string]bool{},
		SortwkPrimaryCylThreshold: cfg.Rules.Sortwk.PrimaryCylThreshold,
	})

	// Context (cost model, geometry, price)
	run.ID = "run-golden" // stable id for snapshot
	run.StartedAt = time.Time{}
	run.Source = "samples/bank-small"
	run.IRVersion = ir.Version

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
	run.Context.RuleSeverityThreshold = "LOW"

	// Cost annotate
	for i := range run.Jobs {
		for j := range run.Jobs[i].Steps {
			size := cost.EstimateSizeMB(&run.Jobs[i].Steps[j], run.Context.Geometry)
			run.Jobs[i].Steps[j].Annotations.SizeMB = size
			run.Jobs[i].Steps[j].Annotations.Cost = cost.Estimate(&run.Jobs[i].Steps[j], run.Context)
		}
	}

	// Evaluate rules
	run.Findings = rules.Evaluate(&run)

	// Normalize volatile fields before snapshot
	norm := normalize(run)

	// Serialize pretty
	got, err := json.MarshalIndent(norm, "", "  ")
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}

	if *update {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(goldenFile, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", goldenFile)
		return
	}

	want, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden (%s): %v\nRun with: go test ./test/golden -run TestGolden_PayrollSnapshot -args -update", goldenFile, err)
	}

	if !bytes.Equal(bytes.TrimSpace(got), bytes.TrimSpace(want)) {
		tmp := filepath.Join(t.TempDir(), "got.json")
		_ = os.WriteFile(tmp, got, 0o644)
		t.Fatalf("golden mismatch.\n  golden: %s\n  actual: %s\nTip: update with\n  go test ./test/golden -run TestGolden_PayrollSnapshot -count=1 -args -update", goldenFile, tmp)
	}
}

type runLite struct {
	ID        string        `json:"id"`
	StartedAt string        `json:"started_at"`
	Source    string        `json:"source,omitempty"`
	IRVersion string        `json:"ir_version,omitempty"`
	Context   ir.Context    `json:"context"`
	Jobs      []jobLite     `json:"jobs"`
	Findings  []findingLite `json:"findings"`
}

type jobLite struct {
	Name  string     `json:"name"`
	Steps []stepLite `json:"steps"`
}

type stepLite struct {
	Name        string           `json:"name"`
	Program     string           `json:"program"`
	Ordinal     int              `json:"ordinal"`
	Annotations stepAnnoLite     `json:"annotations"`
}

type stepAnnoLite struct {
	Cost   ir.Cost `json:"cost"`
	SizeMB float64 `json:"size_mb,omitempty"`
}

type findingLite struct {
	RuleID      string  `json:"rule_id"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Job         string  `json:"job"`
	Step        string  `json:"step,omitempty"`
	Message     string  `json:"message"`
	SavingsMIPS float64 `json:"savings_mips,omitempty"`
	SavingsUSD  float64 `json:"savings_usd,omitempty"`
}

// normalize removes volatile fields (IDs/timestamps) and sorts deterministically.
func normalize(run ir.Run) runLite {
	// Jobs/steps shallow copy for sorting
	jobs := make([]jobLite, 0, len(run.Jobs))
	for _, j := range run.Jobs {
		steps := make([]stepLite, 0, len(j.Steps))
		for _, s := range j.Steps {
			steps = append(steps, stepLite{
				Name:    s.Name,
				Program: s.Program,
				Ordinal: s.Ordinal,
				Annotations: stepAnnoLite{
					Cost:   s.Annotations.Cost,
					SizeMB: s.Annotations.SizeMB,
				},
			})
		}
		sort.Slice(steps, func(i, k int) bool {
			if steps[i].Ordinal == steps[k].Ordinal {
				return steps[i].Name < steps[k].Name
			}
			return steps[i].Ordinal < steps[k].Ordinal
		})
		jobs = append(jobs, jobLite{Name: j.Name, Steps: steps})
	}
	sort.Slice(jobs, func(i, k int) bool { return jobs[i].Name < jobs[k].Name })

	// Findings without volatile ID; sort by (Severity desc, RuleID, Job, Step)
	finds := make([]findingLite, 0, len(run.Findings))
	for _, f := range run.Findings {
		finds = append(finds, findingLite{
			RuleID:      f.RuleID,
			Type:        f.Type,
			Severity:    f.Severity,
			Job:         f.Job,
			Step:        f.Step,
			Message:     f.Message,
			SavingsMIPS: f.SavingsMIPS,
			SavingsUSD:  f.SavingsUSD,
		})
	}
	sevRank := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
	sort.Slice(finds, func(i, k int) bool {
		si, sk := sevRank[finds[i].Severity], sevRank[finds[k].Severity]
		if si != sk {
			return si > sk
		}
		if finds[i].RuleID != finds[k].RuleID {
			return finds[i].RuleID < finds[k].RuleID
		}
		if finds[i].Job != finds[k].Job {
			return finds[i].Job < finds[k].Job
		}
		return finds[i].Step < finds[k].Step
	})

	return runLite{
		ID:        "run-golden",
		StartedAt: "", // zeroed
		Source:    run.Source,
		IRVersion: run.IRVersion,
		Context:   run.Context,
		Jobs:      jobs,
		Findings:  finds,
	}
}
