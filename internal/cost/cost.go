package cost

import (
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

func Estimate(step *ir.Step, ctx ir.Context) ir.Cost {
	pgm := strings.ToUpper(step.Program)
	cpu := 1.0
	switch pgm {
	case "SORT":
		cpu = estimateSortCPU(step)
	case "IEBGENER":
		cpu = 2.0
	case "IDCAMS":
		cpu = 4.0
	default:
		cpu = 1.0
	}
	mips := cpu // placeholder mapping (1 CPU sec ≈ 1 MIPS unit, heuristic)
	usd := 0.0
	if ctx.MIPSToUSD > 0 {
		usd = mips * ctx.MIPSToUSD
	}
	return ir.Cost{CPUSeconds: cpu, MIPS: mips, USD: usd}
}

func estimateSortCPU(step *ir.Step) float64 {
	// Cheap heuristics: identity sort cheaper
	sysin := ""
	for _, dd := range step.DD {
		if strings.EqualFold(dd.DDName, "SYSIN") {
			sysin = strings.ToUpper(dd.Content)
			break
		}
	}
	if strings.Contains(sysin, "FIELDS=COPY") {
		return 3.0
	}
	// If many SORTWK DDs, assume larger workload
	sortwkCount := 0
	for _, dd := range step.DD {
		if strings.HasPrefix(strings.ToUpper(dd.DDName), "SORTWK") {
			sortwkCount++
		}
	}
	if sortwkCount >= 6 {
		return 12.0
	}
	return 10.0
}
