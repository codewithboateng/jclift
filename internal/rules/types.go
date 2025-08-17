package rules

import "github.com/codewithboateng/jclift/internal/ir"

// Rule represents a single analysis rule executed over a Job.
type Rule struct {
	ID               string
	Summary          string
	Type             string // "COST" | "RISK" (advisory)
	DefaultSeverity  string // "LOW" | "MEDIUM" | "HIGH" (advisory)
	Docs             string // URL or repo path to docs for this rule
	Eval             func(job *ir.Job) []ir.Finding
}