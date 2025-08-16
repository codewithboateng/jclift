package cost

import "github.com/codewithboateng/jclift/internal/ir"

// Estimate applies a trivial heuristic for MVP skeleton.
func Estimate(step *ir.Step, ctx ir.Context) ir.Cost {
	// Placeholder: constant small cost so pipeline compiles
	cpu := 1.0
	mips := 1.0
	usd := 0.0
	if ctx.MIPSToUSD > 0 {
		usd = mips * ctx.MIPSToUSD
	}
	return ir.Cost{CPUSeconds: cpu, MIPS: mips, USD: usd}
}
