package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net/http"

	"github.com/codewithboateng/jclift/internal/api"
	"github.com/codewithboateng/jclift/internal/security"

	"github.com/codewithboateng/jclift/internal/cost"
	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/parser"
	"github.com/codewithboateng/jclift/internal/reporting"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/rulesdsl"
	"github.com/codewithboateng/jclift/internal/shared"
	"github.com/codewithboateng/jclift/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "admin":
		adminCmd(os.Args[2:])
	case "serve":
		serveCmd(os.Args[2:])
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
	fmt.Fprintf(os.Stderr, `jclift – JCL Cost/Risk Analyzer

Usage:
  jclift analyze --path <input-dir> --out <reports-dir> [--db ./jclift.db] [--mips-usd 250] [--config ./configs/jclift.yaml]
  jclift report  --run <run-id>     --out <reports-dir> [--db ./jclift.db] [--config ./configs/jclift.yaml]
  jclift diff    --base <run-id> --head <run-id> --out <reports-dir> [--db ./jclift.db] [--config ./configs/jclift.yaml]
  jclift version
`)
}

func analyzeCmd(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	configPath   := fs.String("config", "", "Path to YAML config (optional)")
	inPath       := fs.String("path", "", "Path to input JCL directory")
	outDir       := fs.String("out", "", "Output directory for reports")
	dbPath       := fs.String("db", "", "SQLite database path")
	mipsUSD      := fs.Float64("mips-usd", 0, "USD per MIPS unit (optional)")
	sevThresh    := fs.String("severity-threshold", "", "Minimum severity to report (LOW|MEDIUM|HIGH)")
	rulesDisable := fs.String("rules-disable", "", "Comma-separated rule IDs to disable")
	rulesPack    := fs.String("rules-pack", "", "Path to YAML rule pack (DSL)") // ✅ define BEFORE Parse
	failOn       := fs.Bool("fail-on-findings", false, "Exit non-zero if any findings remain after threshold/disable")
	_ = fs.Parse(args)

	// Load config + init logger
	cfg, _ := shared.LoadConfig(*configPath)
	logger := shared.InitLogger(cfg.Logging.Format, cfg.Logging.Level)
	_ = logger

	// Precedence: flags > config > defaults
	if *inPath == "" && len(cfg.Analysis.Sources) > 0 { *inPath = cfg.Analysis.Sources[0] }
	if *outDir == "" { *outDir = cfg.Reporting.OutDir }
	if *dbPath == "" { *dbPath = cfg.Database.DSN }
	if *mipsUSD == 0 && cfg.Analysis.MIPSToUSD > 0 { *mipsUSD = cfg.Analysis.MIPSToUSD }

	// Severity threshold + disabled rules
	sth := cfg.Rules.SeverityThreshold
	if *sevThresh != "" { sth = *sevThresh }
	disable := map[string]bool{}
	for _, id := range cfg.Rules.Disable { disable[strings.ToUpper(strings.TrimSpace(id))] = true }
	if *rulesDisable != "" {
		for _, id := range strings.Split(*rulesDisable, ",") {
			disable[strings.ToUpper(strings.TrimSpace(id))] = true
		}
	}
	sortwkThresh := cfg.Rules.Sortwk.PrimaryCylThreshold

	// I/O prep
	if *inPath == "" {
		fmt.Fprintln(os.Stderr, "analyze: --path (or analysis.sources in config) is required")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "analyze: cannot create out dir:", err)
		os.Exit(1)
	}

	// Configure rules engine
	rules.SetSettings(rules.Settings{
		SeverityThreshold:         sth,
		Disabled:                  disable,
		SortwkPrimaryCylThreshold: sortwkThresh,
	})

	// Parse input → build Run
	run, diags := parser.Parse(*inPath)
	if len(diags.Warnings) > 0 {
		slog.Warn("parse warnings", "warnings", diags.Warnings)
	}
	run.ID = fmt.Sprintf("run-%d", time.Now().Unix())
	run.StartedAt = time.Now().UTC()
	run.Context.MIPSToUSD = *mipsUSD
	run.Context.RuleSeverityThreshold = sth
	run.Context.DisabledRules = make([]string, 0, len(disable))
	for id := range disable { run.Context.DisabledRules = append(run.Context.DisabledRules, id) }

	// Inject geometry & model into Context
	run.Context.Geometry.TracksPerCyl  = cfg.Cost.Geometry.TracksPerCyl
	run.Context.Geometry.BytesPerTrack = cfg.Cost.Geometry.BytesPerTrack
	run.Context.Model.MIPSPerCPU = cfg.Cost.Model.MIPSPerCPU
	run.Context.Model.SortAlpha  = cfg.Cost.Model.Sort.Alpha
	run.Context.Model.SortBeta   = cfg.Cost.Model.Sort.Beta
	run.Context.Model.CopyAlpha  = cfg.Cost.Model.Copy.Alpha
	run.Context.Model.CopyBeta   = cfg.Cost.Model.Copy.Beta
	run.Context.Model.IDAlpha    = cfg.Cost.Model.IDCAMS.Alpha
	run.Context.Model.IDBeta     = cfg.Cost.Model.IDCAMS.Beta

	// ✅ Load optional DSL rules pack (after settings, before evaluation)
	if *rulesPack != "" {
		if n, err := rulesdsl.LoadAndRegister(*rulesPack); err != nil {
			slog.Warn("rules pack load error", "err", err, "path", *rulesPack)
		} else {
			slog.Info("rules pack loaded", "count", n, "path", *rulesPack)
		}
	}

	// Cost annotate (add SizeMB)
	for i := range run.Jobs {
		for j := range run.Jobs[i].Steps {
			size := cost.EstimateSizeMB(&run.Jobs[i].Steps[j], run.Context.Geometry)
			c := cost.Estimate(&run.Jobs[i].Steps[j], run.Context)
			run.Jobs[i].Steps[j].Annotations.SizeMB = size
			run.Jobs[i].Steps[j].Annotations.Cost = c
		}
	}

	// Evaluate rules
	run.Findings = rules.Evaluate(&run)

	// Persist & report
	// Persist & report (open DB earlier so we can load waivers)
	db, err := storage.OpenSQLite(*dbPath)
	if err != nil { slog.Error("db open error", "err", err); os.Exit(1) }
	defer db.Close()
	if err := db.CreateSchema(); err != nil { slog.Error("db schema error", "err", err); os.Exit(1) }

	// Apply waivers (active only)
	waivers, err := db.ListWaivers(true)
	if err != nil { slog.Warn("waiver list error", "err", err) }
	if len(waivers) > 0 {
		kept, waived := rules.ApplyWaivers(run.Findings, waivers)
		run.Findings = kept
		run.Context.WaivedCount = waived // <-- record how many were waived
		slog.Info("waivers applied", "waived", waived, "remaining", len(run.Findings))
	}

	// Save run
	if err := db.SaveRun(&run); err != nil { slog.Error("db save run error", "err", err); os.Exit(1) }

	// Reports
	jsonPath, _ := reporting.WriteJSON(run.ID, *outDir, &run)
	htmlPath, _ := reporting.WriteHTML(run.ID, *outDir, &run)

	slog.Info("analyze complete", "run", run.ID, "json", jsonPath, "html", htmlPath, "db", filepath.Clean(*dbPath))
	fmt.Printf("Analyze OK\n  Run: %s\n  JSON: %s\n  HTML: %s\n  DB: %s\n", run.ID, jsonPath, htmlPath, filepath.Clean(*dbPath))

	if *failOn && len(run.Findings) > 0 { os.Exit(3) }

}

func reportCmd(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to YAML config (optional)")
	runID := fs.String("run", "", "Run ID")
	outDir := fs.String("out", "", "Output directory")
	dbPath := fs.String("db", "", "SQLite database path")
	_ = fs.Parse(args)

	cfg, _ := shared.LoadConfig(*configPath)
	shared.InitLogger(cfg.Logging.Format, cfg.Logging.Level)

	if *outDir == "" {
		*outDir = cfg.Reporting.OutDir
	}
	if *dbPath == "" {
		*dbPath = cfg.Database.DSN
	}
	if *runID == "" {
		fmt.Fprintln(os.Stderr, "report: --run is required")
		os.Exit(2)
	}

	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		slog.Error("db open error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	run, err := db.LoadRun(*runID)
	if err != nil {
		slog.Error("load run error", "err", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		slog.Error("cannot create out dir", "err", err)
		os.Exit(1)
	}
	jsonPath, _ := reporting.WriteJSON(run.ID, *outDir, &run)
	htmlPath, _ := reporting.WriteHTML(run.ID, *outDir, &run)
	fmt.Printf("Report OK\n  Run: %s\n  JSON: %s\n  HTML: %s\n", run.ID, jsonPath, htmlPath)
}

func diffCmd(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to YAML config (optional)")
	base := fs.String("base", "", "Base run ID")
	head := fs.String("head", "", "Head run ID")
	outDir := fs.String("out", "", "Output directory")
	dbPath := fs.String("db", "", "SQLite database path")
	_ = fs.Parse(args)

	cfg, _ := shared.LoadConfig(*configPath)
	shared.InitLogger(cfg.Logging.Format, cfg.Logging.Level)

	if *outDir == "" {
		*outDir = cfg.Reporting.OutDir
	}
	if *dbPath == "" {
		*dbPath = cfg.Database.DSN
	}
	if *base == "" || *head == "" {
		fmt.Fprintln(os.Stderr, "diff: --base and --head are required")
		os.Exit(2)
	}
	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		slog.Error("db open error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	br, err := db.LoadRun(*base)
	if err != nil {
		slog.Error("load base run error", "err", err)
		os.Exit(1)
	}
	hr, err := db.LoadRun(*head)
	if err != nil {
		slog.Error("load head run error", "err", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		slog.Error("cannot create out dir", "err", err)
		os.Exit(1)
	}
	path, _ := reporting.WriteDiffJSON(*base, *head, *outDir, &br, &hr)
	fmt.Printf("Diff OK\n  %s\n", path)
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to YAML config (optional)")
	dbPath := fs.String("db", "", "SQLite database path")
	listen := fs.String("listen", ":8080", "Listen address (e.g. :8080)")
	corsAllow := fs.String("cors-allow", "*", "Comma-separated allowed origins (use * for any)")
	_ = fs.Parse(args)

	// Config + logger
	cfg, _ := shared.LoadConfig(*configPath)
	logger := shared.InitLogger(cfg.Logging.Format, cfg.Logging.Level)

	// Resolve DB path
	if *dbPath == "" {
		*dbPath = cfg.Database.DSN
	}
	db, err := storage.OpenSQLite(*dbPath)
	if err != nil {
		slog.Error("db open error", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Parse allowed origins
	var allowed []string
	for _, p := range strings.Split(*corsAllow, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			allowed = append(allowed, p)
		}
	}
	if len(allowed) == 0 {
		allowed = []string{"*"}
	}

	// Start server
	s := &api.Server{
	DB: db,
	UserStore: db,                   // <-- set it now that it's exported
	Logger: logger,
	AllowedOrigins: allowed,
	SessionDuration: 24 * time.Hour,
}
	
	slog.Info("api listening", "addr", *listen, "cors_allow", strings.Join(allowed, ","))
	if err := http.ListenAndServe(*listen, s.Routes()); err != nil {
		slog.Error("api serve error", "err", err)
		os.Exit(1)
	}
}

func adminCmd(args []string) {
	fs := flag.NewFlagSet("admin", flag.ExitOnError)
	sub := fs.String("cmd", "", "Command: create-user")
	username := fs.String("username", "", "Username")
	password := fs.String("password", "", "Password")
	role := fs.String("role", "admin", "Role (admin|viewer)")
	dbPath := fs.String("db", "./jclift.db", "SQLite database path")
	_ = fs.Parse(args)

	if *sub != "create-user" {
		fmt.Fprintln(os.Stderr, "admin: --cmd create-user required")
		os.Exit(2)
	}
	if *username == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "admin: --username and --password are required")
		os.Exit(2)
	}

	db, err := storage.OpenSQLite(*dbPath)
	if err != nil { slog.Error("db open error", "err", err); os.Exit(1) }
	defer db.Close()
	if err := db.CreateSchema(); err != nil { slog.Error("schema", "err", err); os.Exit(1) }

	hash, err := security.HashPassword(*password)
	if err != nil { slog.Error("hash", "err", err); os.Exit(1) }

	id, err := db.CreateUser(*username, hash, *role)
	if err != nil { slog.Error("create user", "err", err); os.Exit(1) }
	fmt.Printf("User created id=%d username=%s role=%s\n", id, *username, *role)
}
