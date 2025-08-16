package rules

import "github.com/codewithboateng/jclift/internal/ir"

// Rule represents a single analysis rule executed over a Job.
type Rule struct {
	ID      string
	Summary string
	// Eval inspects the job (and its steps/DDs) and returns findings.
	Eval func(job *ir.Job) []ir.Finding
}
