package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

type Diagnostics struct {
	Warnings []string
}

func Parse(path string) (ir.Run, Diagnostics) {
	var run ir.Run
	run.IRVersion = ir.Version
	run.Source = filepath.Clean(path)
	diags := Diagnostics{}

	// Skeleton: walk files ending with .jcl or .txt and create fake jobs/steps
	_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".jcl") && !strings.HasSuffix(strings.ToLower(d.Name()), ".txt") {
			return nil
		}
		job := ir.Job{Name: strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))}
		job.Steps = append(job.Steps, ir.Step{
			Name:    "STEP1",
			Program: "UNKNOWN",
			Ordinal: 1,
		})
		run.Jobs = append(run.Jobs, job)
		return nil
	})

	if len(run.Jobs) == 0 {
		diags.Warnings = append(diags.Warnings, "no JCL-like files found")
	}
	return run, diags
}
