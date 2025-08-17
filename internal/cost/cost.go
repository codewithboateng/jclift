package cost

import (
	"math"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func Estimate(step *ir.Step, ctx ir.Context) ir.Cost {
	sizeMB := EstimateSizeMB(step, ctx.Geometry)

	// Defaults if not configured
	mipsPerCPU := ctx.Model.MIPSPerCPU
	if mipsPerCPU <= 0 { mipsPerCPU = 1.0 }

	alphaS, betaS := ctx.Model.SortAlpha, ctx.Model.SortBeta
	if alphaS == 0 && betaS == 0 { alphaS, betaS = 0.3, 0.001 }

	alphaC, betaC := ctx.Model.CopyAlpha, ctx.Model.CopyBeta
	if alphaC == 0 && betaC == 0 { alphaC, betaC = 0.10, 0.0005 }

	alphaI, betaI := ctx.Model.IDAlpha, ctx.Model.IDBeta
	if alphaI == 0 && betaI == 0 { alphaI, betaI = 0.15, 0.0006 }

	cpu := 0.5 // default small base to avoid zeros
	switch strings.ToUpper(step.Program) {
	case "SORT":
		mb := math.Max(sizeMB, 1.0)
		cpu = alphaS + betaS*mb*math.Log2(mb)
	case "IEBGENER":
		mb := math.Max(sizeMB, 1.0)
		cpu = alphaC + betaC*mb
	case "IDCAMS":
		mb := math.Max(sizeMB, 1.0)
		cpu = alphaI + betaI*mb
	default:
		cpu = 0.2 + 0.0002*sizeMB
	}

	mips := cpu * mipsPerCPU
	usd := 0.0
	if ctx.MIPSToUSD > 0 {
		usd = mips * ctx.MIPSToUSD
	}

	return ir.Cost{
		CPUSeconds: cpu,
		MIPS:       mips,
		USD:        usd,
	}
}
