package perf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codewithboateng/jclift/internal/cost"
	"github.com/codewithboateng/jclift/internal/parser"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/shared"
)

const benchSample = `//B JOB (1),'B',CLASS=A,MSGCLASS=X
//S1 EXEC PGM=SORT
//SYSIN DD *
  SORT  FIELDS=COPY
/*
//SORTWK01 DD UNIT=SYSDA,SPACE=(CYL,(500,50))
//S2 EXEC PGM=IEBGENER
//SYSUT1 DD DSN=IN,DISP=SHR
//SYSUT2 DD DSN=OUT,DISP=(NEW,CATLG)
`

func BenchmarkAnalyze_Small(b *testing.B) {
	dir := b.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.jcl"), []byte(benchSample), 0o644); err != nil {
		b.Fatal(err)
	}

	cfg := shared.DefaultConfig()
	rules.SetSettings(rules.Settings{
		SeverityThreshold:         "LOW",
		Disabled:                  map[string]bool{},
		SortwkPrimaryCylThreshold: cfg.Rules.Sortwk.PrimaryCylThreshold,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run, _ := parser.Parse(dir)
		// cost annotate
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

		for j := range run.Jobs {
			for k := range run.Jobs[j].Steps {
				size := cost.EstimateSizeMB(&run.Jobs[j].Steps[k], run.Context.Geometry)
				run.Jobs[j].Steps[k].Annotations.SizeMB = size
				run.Jobs[j].Steps[k].Annotations.Cost = cost.Estimate(&run.Jobs[j].Steps[k], run.Context)
			}
		}
		run.Findings = rules.Evaluate(&run)
		if len(run.Jobs) == 0 {
			b.Fatal("no jobs parsed")
		}
	}
}
