package cost

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

var spaceRe = regexp.MustCompile(`\b(CYL|TRK)\s*,?\s*\(\s*(\d+)`)

// EstimateSizeMB tries to infer MB processed by a step.
// Heuristics:
// - If step has SORTWKnn with SPACE, sum primaries as proxy for size
// - Else if any DD has SPACE on output (NEW/CATLG), use that primary
// - Else return a small floor (1 MB)
func EstimateSizeMB(step *ir.Step, geom ir.Geometry) float64 {
	trkPerCyl := geom.TracksPerCyl
	if trkPerCyl <= 0 { trkPerCyl = 15 }
	bytesPerTrack := geom.BytesPerTrack
	if bytesPerTrack <= 0 { bytesPerTrack = 56664 }

	cylToMB := func(cyl int) float64 {
		return float64(cyl*trkPerCyl*bytesPerTrack) / (1024.0 * 1024.0)
	}
	trkToMB := func(trk int) float64 {
		return float64(trk*bytesPerTrack) / (1024.0 * 1024.0)
	}

	sumMB := 0.0
	// Prefer SORTWK for SORT
	if strings.EqualFold(step.Program, "SORT") {
		for _, dd := range step.DD {
			if strings.HasPrefix(strings.ToUpper(dd.DDName), "SORTWK") {
				if m := spaceRe.FindStringSubmatch(strings.ToUpper(dd.Space)); len(m) >= 3 {
					n, _ := strconv.Atoi(m[2])
					if strings.EqualFold(m[1], "CYL") {
						sumMB += cylToMB(n)
					} else {
						sumMB += trkToMB(n)
					}
				}
			}
		}
		if sumMB > 0 {
			return math.Max(sumMB, 1.0)
		}
	}

	// Otherwise, try any output NEW/CATLG DD with SPACE
	for _, dd := range step.DD {
		upDisp := strings.ToUpper(dd.DISP)
		if upDisp == "" || strings.Contains(upDisp, "NEW") || strings.Contains(upDisp, "CATLG") || strings.Contains(upDisp, "MOD") {
			if m := spaceRe.FindStringSubmatch(strings.ToUpper(dd.Space)); len(m) >= 3 {
				n, _ := strconv.Atoi(m[2])
				if strings.EqualFold(m[1], "CYL") {
					return math.Max(cylToMB(n), 1.0)
				}
				return math.Max(trkToMB(n), 1.0)
			}
		}
	}

	return 1.0 // floor
}
