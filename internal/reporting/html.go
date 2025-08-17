package reporting

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"

	"github.com/codewithboateng/jclift/internal/ir"
)

func WriteHTML(runID, outDir string, run *ir.Run) (string, error) {
	path := filepath.Join(outDir, runID+".html")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Totals
	var totalCPU, totalMIPS, totalUSD float64
	for _, j := range run.Jobs {
		for _, s := range j.Steps {
			totalCPU += s.Annotations.Cost.CPUSeconds
			totalMIPS += s.Annotations.Cost.MIPS
			totalUSD += s.Annotations.Cost.USD
		}
	}

	// Head + styles
	fmt.Fprintf(f, "<!doctype html><html><head><meta charset='utf-8'><title>%s</title>", html.EscapeString(runID))
	fmt.Fprint(f, "<style>body{font-family:system-ui,Arial,sans-serif;padding:20px;line-height:1.4} table{border-collapse:collapse;margin:8px 0} td,th{border:1px solid #ddd;padding:6px} h1,h2{margin:6px 0 4px} .dim{color:#666} .mono{font-family:ui-monospace,Menlo,Consolas,monospace}</style>")
	fmt.Fprint(f, "</head><body>")

	// Title + summary
	fmt.Fprintf(f, "<h1>jclift report – <span class='mono'>%s</span></h1>", html.EscapeString(runID))
	fmt.Fprintf(f, "<p>Jobs: %d &nbsp; Findings: %d</p>", len(run.Jobs), len(run.Findings))
	fmt.Fprintf(f, "<p><b>Estimated totals</b>: CPU=%.1fs &nbsp; MIPS=%.2f &nbsp; USD=%.2f <span class='dim'>(heuristic)</span></p>", totalCPU, totalMIPS, totalUSD)

	// Rate (MIPS→USD)
	if run.Context.MIPSToUSD > 0 {
		fmt.Fprintf(f, "<p class='dim'>Rate: 1 MIPS ≈ %.2f USD</p>", run.Context.MIPSToUSD)
	}

	// Severity/disabled banner
	fmt.Fprintf(f, "<p class='dim'>Severity threshold: %s", html.EscapeString(run.Context.RuleSeverityThreshold))
	if n := len(run.Context.DisabledRules); n > 0 {
		fmt.Fprintf(f, " &nbsp; Disabled rules: %d", n)
	}
	fmt.Fprint(f, "</p>")

	// Geometry/model banner
	if run.Context.Geometry.TracksPerCyl > 0 || run.Context.Geometry.BytesPerTrack > 0 || run.Context.Model.MIPSPerCPU > 0 {
		fmt.Fprintf(f, "<p class='dim'>Geometry: %d trk/cyl, %d bytes/trk &nbsp; Model: MIPS/CPU=%.2f</p>",
			run.Context.Geometry.TracksPerCyl,
			run.Context.Geometry.BytesPerTrack,
			run.Context.Model.MIPSPerCPU,
		)
	}

	// Top Offenders (by USD desc, then MIPS)
	type tf struct {
		ir.Finding
		usd float64
	}
	if len(run.Findings) > 0 {
		var tops []tf
		for _, fd := range run.Findings {
			tops = append(tops, tf{Finding: fd, usd: fd.SavingsUSD})
		}
		sort.Slice(tops, func(i, j int) bool {
			if tops[i].usd == tops[j].usd {
				return tops[i].SavingsMIPS > tops[j].SavingsMIPS
			}
			return tops[i].usd > tops[j].usd
		})
		if len(tops) > 0 {
			fmt.Fprint(f, "<h2>Top Offenders</h2><table><tr><th>Severity</th><th>Rule</th><th>Job</th><th>Step</th><th>Projected USD</th><th>Projected MIPS</th><th>Message</th></tr>")
			limit := len(tops)
			if limit > 20 {
				limit = 20
			}
			for i := 0; i < limit; i++ {
				fd := tops[i].Finding
				fmt.Fprintf(f, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%.2f</td><td>%.2f</td><td>%s</td></tr>",
					html.EscapeString(fd.Severity),
					html.EscapeString(fd.RuleID),
					html.EscapeString(fd.Job),
					html.EscapeString(fd.Step),
					fd.SavingsUSD,
					fd.SavingsMIPS,
					html.EscapeString(fd.Message),
				)
			}
			fmt.Fprint(f, "</table>")
		}
	}

	// All findings
	if len(run.Findings) > 0 {
		fmt.Fprint(f, "<h2>All Findings</h2><table><tr><th>Severity</th><th>Rule</th><th>Job</th><th>Step</th><th>Message</th></tr>")
		for _, fd := range run.Findings {
			fmt.Fprintf(f, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
				html.EscapeString(fd.Severity),
				html.EscapeString(fd.RuleID),
				html.EscapeString(fd.Job),
				html.EscapeString(fd.Step),
				html.EscapeString(fd.Message),
			)
		}
		fmt.Fprint(f, "</table>")
	} else {
		fmt.Fprint(f, "<h2>All Findings</h2><p class='dim'>No findings at or above the configured threshold.</p>")
	}

	fmt.Fprint(f, "</body></html>")
	return path, nil
}
