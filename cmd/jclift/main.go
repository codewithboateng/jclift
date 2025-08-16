package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codewithboateng/jclift/internal/cost"
	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/parser"
	"github.com/codewithboateng/jclift/internal/reporting"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "analyze":
		analyzeCmd(os.Args[2:])
	case "report":
		reportCmd(os.Args[2:])
	case "diff":
		diffCmd(os.Args[2:])
	case "version":
		fmt.Println("jclift (MVP skeleton) IR:", ir.Version)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `jclift – JCL Cost/Risk Analyzer (Go-only skeleton)

Usage:
  jclift analyze --path <input-dir> --out <reports-dir> [--db ./jclift.db]
  jclift report  --run <run-id>     --out <reports-dir> [--db ./jclift.db]
  jclift diff    --base <run-id> --head <run-id> --out <reports-dir> [--db ./jclift.db]
  jclift version

`)
}

func analyzeCmd(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	inPath := fs.String("path", "", "Path to input JCL directory")
	outDir := fs.String("out", "./reports", "Output directory for reports")
	dbPath := fs.String("db", "./jclift.db", "SQLite database path")
	_ = fs.Parse(args)

	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "analyze: --path is required")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "analyze: cannot create out dir:", err)
		os.Exit(1)
	}

	// Parse JCL → IR
	run, diags := parser.Parse(*inPath)
	if len(diags.Warnings) > 0 {
		fmt.Fprintln(os.Stderr, "parse warnings:", diags.Warnings)
	}
	run.ID = fmt.Sprintf("run-%d", time.Now().Unix())
	run.StartedAt = time.Now().UTC()

	// Cost annotate (heuristic placeholder)
	for i := range run.Jobs {
		for j := range run.Jobs[i].Steps {
			c := cost.Estimate(&run.Jobs[i].Steps[j], run.Context)
			run.Jobs[i].Steps[j].Annotations.Cost = c
		}
	}

	// Persist
	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open error:", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.CreateSchema(); err != nil {
		fmt.Fprintln(os.Stderr, "db schema error:", err)
		os.Exit(1)
	}
	if err := db.SaveRun(&run); err != nil {
		fmt.Fprintln(os.Stderr, "db save run error:", err)
		os.Exit(1)
	}

	// Reports (JSON + HTML placeholders)
	jsonPath, _ := reporting.WriteJSON(run.ID, *outDir, &run)
	htmlPath, _ := reporting.WriteHTML(run.ID, *outDir, &run)

	fmt.Printf("Analyze OK\n  Run: %s\n  JSON: %s\n  HTML: %s\n  DB: %s\n",
		run.ID, jsonPath, htmlPath, filepath.Clean(*dbPath))

	// Evaluate rules → findings
	run.Findings = rules.Evaluate(&run)
}

func reportCmd(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	runID := fs.String("run", "", "Run ID")
	outDir := fs.String("out", "./reports", "Output directory")
	dbPath := fs.String("db", "./jclift.db", "SQLite database path")
	_ = fs.Parse(args)

	if *runID == "" {
		fmt.Fprintln(os.Stderr, "report: --run is required")
		os.Exit(2)
	}

	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open error:", err)
		os.Exit(1)
	}
	defer db.Close()

	run, err := db.LoadRun(*runID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load run error:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "cannot create out dir:", err)
		os.Exit(1)
	}
	jsonPath, _ := reporting.WriteJSON(run.ID, *outDir, &run)
	htmlPath, _ := reporting.WriteHTML(run.ID, *outDir, &run)
	fmt.Printf("Report OK\n  Run: %s\n  JSON: %s\n  HTML: %s\n", run.ID, jsonPath, htmlPath)
}

func diffCmd(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	base := fs.String("base", "", "Base run ID")
	head := fs.String("head", "", "Head run ID")
	outDir := fs.String("out", "./reports", "Output directory")
	dbPath := fs.String("db", "./jclift.db", "SQLite database path")
	_ = fs.Parse(args)

	if *base == "" || *head == "" {
		fmt.Fprintln(os.Stderr, "diff: --base and --head are required")
		os.Exit(2)
	}
	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open error:", err)
		os.Exit(1)
	}
	defer db.Close()

	br, err := db.LoadRun(*base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load base run error:", err); os.Exit(1)
	}
	hr, err := db.LoadRun(*head)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load head run error:", err); os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "cannot create out dir:", err); os.Exit(1)
	}
	path, _ := reporting.WriteDiffJSON(*base, *head, *outDir, &br, &hr)
	fmt.Printf("Diff OK\n  %s\n", path)
}
