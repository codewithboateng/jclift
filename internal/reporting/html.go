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

	totalCPU, totalMIPS, totalUSD := 0.0, 0.0, 0.0
	for _, j := range run.Jobs {
		for _, s := range j.Steps {
			totalCPU += s.Annotations.Cost.CPUSeconds
			totalMIPS += s.Annotations.Cost.MIPS
			totalUSD += s.Annotations.Cost.USD
		}
	}

	fmt.Fprintf(f, "<!doctype html><html><head><meta charset='utf-8'><title>%s</title>", runID)
	fmt.Fprint(f, "<style>body{font-family:system-ui,Arial,sans-serif;padding:20px} table{border-collapse:collapse} td,th{border:1px solid #ddd;padding:6px} .dim{color:#666}</style>")
	fmt.Fprint(f, "</head><body>")
	fmt.Fprintf(f, "<h1>jclift report – %s</h1>", html.EscapeString(runID))
	fmt.Fprintf(f, "<p>Jobs: %d &nbsp; Findings: %d</p>", len(run.Jobs), len(run.Findings))
	fmt.Fprintf(f, "<p><b>Estimated totals</b>: CPU=%.1fs &nbsp; MIPS=%.1f &nbsp; USD=%.2f <span class='dim'>(heuristic)</span></p>", totalCPU, totalMIPS, totalUSD)

	// Top offenders (by SavingsUSD first, then MIPS)
	type tf struct {
		ir.Finding
		usd float64
	}
	var tops []tf
	for _, fd := range run.Findings {
		usd := fd.SavingsUSD
		tops = append(tops, tf{fd, usd})
	}
	sort.Slice(tops, func(i, j int) bool {
		if tops[i].usd == tops[j].usd {
			return tops[i].SavingsMIPS > tops[j].SavingsMIPS
		}
		return tops[i].usd > tops[j].usd
	})
	if len(tops) > 0 {
		fmt.Fprint(f, "<h2>Top Offenders</h2><table><tr><th>Rule</th><th>Job</th><th>Step</th><th>Projected USD</th><th>Projected MIPS</th><th>Message</th></tr>")
		limit := len(tops)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			fd := tops[i].Finding
			fmt.Fprintf(f, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%.2f</td><td>%.2f</td><td>%s</td></tr>",
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
	}

	fmt.Fprint(f, "</body></html>")
	return path, nil
}
