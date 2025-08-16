package reporting

import (
	"fmt"
	"html"
	"os"
	"path/filepath"

	"github.com/codewithboateng/jclift/internal/ir"
)

func WriteHTML(runID, outDir string, run *ir.Run) (string, error) {
	path := filepath.Join(outDir, runID+".html")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fmt.Fprintf(f, "<!doctype html><html><head><meta charset='utf-8'><title>%s</title>", runID)
	fmt.Fprint(f, "<style>body{font-family:system-ui,Arial,sans-serif;padding:20px} table{border-collapse:collapse} td,th{border:1px solid #ddd;padding:6px}</style>")
	fmt.Fprint(f, "</head><body>")
	fmt.Fprintf(f, "<h1>jclift report – %s</h1>", html.EscapeString(runID))
	fmt.Fprintf(f, "<p>Jobs: %d &nbsp; Findings: %d</p>", len(run.Jobs), len(run.Findings))

	if len(run.Findings) > 0 {
		fmt.Fprint(f, "<h2>Findings</h2><table><tr><th>Severity</th><th>Rule</th><th>Job</th><th>Step</th><th>Message</th></tr>")
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
