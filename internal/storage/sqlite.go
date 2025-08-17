package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"github.com/codewithboateng/jclift/internal/ir"
)

// DB is the concrete storage backed by SQLite.
type DB struct {
	conn *sql.DB
}

// OpenSQLite opens (and creates if missing) a SQLite DB at path.
func OpenSQLite(path string) (*DB, error) {
	// Pragmas via DSN keep it portable with the modernc driver.
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
	c, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &DB{conn: c}, nil
}

func (db *DB) Close() error { return db.conn.Close() }

// CreateSchema ensures tables (and compatibility views) exist.
func (db *DB) CreateSchema() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS runs (
  id         TEXT PRIMARY KEY,
  started_at TEXT,          -- RFC3339
  source     TEXT,
  ir_version TEXT,
  run_json   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS findings (
  id           TEXT,
  run_id       TEXT NOT NULL,
  job          TEXT,
  step         TEXT,
  rule_id      TEXT,
  type         TEXT,
  severity     TEXT,
  message      TEXT,
  evidence     TEXT,
  savings_mips REAL,
  savings_usd  REAL,
  PRIMARY KEY (id, run_id),
  FOREIGN KEY(run_id) REFERENCES runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_findings_run ON findings(run_id);
CREATE INDEX IF NOT EXISTS idx_findings_rule ON findings(rule_id);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  pass_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'viewer',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  username TEXT,
  action TEXT NOT NULL,
  resource TEXT,
  meta_json TEXT
);

CREATE TABLE IF NOT EXISTS waivers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  rule_id     TEXT NOT NULL,
  job         TEXT,              -- optional exact match; NULL = any
  step        TEXT,              -- optional exact match; NULL = any
  pattern_sub TEXT,              -- optional substring to match evidence/message
  reason      TEXT NOT NULL,
  expires_at  TEXT NOT NULL,     -- RFC3339Nano
  created_by  TEXT NOT NULL,
  created_at  TEXT NOT NULL,
  revoked_at  TEXT               -- NULL = active
);

-- ------------------------------------------------------------------
-- Compatibility views for legacy summary queries (e.g., db-summary)
-- These map expected legacy tables to the normalized schema.
-- ------------------------------------------------------------------
CREATE VIEW IF NOT EXISTS jobs AS
SELECT DISTINCT job
FROM findings
WHERE job IS NOT NULL;

CREATE VIEW IF NOT EXISTS steps AS
SELECT DISTINCT job, step
FROM findings
WHERE job IS NOT NULL AND step IS NOT NULL;
`)
	if err != nil {
		return err
	}
	return nil
}

// SaveRun upserts a run JSON and (re)writes its findings.
func (db *DB) SaveRun(run *ir.Run) error {
	b, err := json.Marshal(run)
	if err != nil {
		return err
	}
	ts := run.StartedAt.UTC().Format(time.RFC3339Nano)

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`INSERT INTO runs (id, started_at, source, ir_version, run_json)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(id) DO UPDATE SET started_at=excluded.started_at, source=excluded.source, ir_version=excluded.ir_version, run_json=excluded.run_json`,
		run.ID, ts, run.Source, run.IRVersion, string(b),
	); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM findings WHERE run_id = ?`, run.ID); err != nil {
		return err
	}
	if len(run.Findings) > 0 {
		stmt, err := tx.Prepare(`
			INSERT INTO findings
			(id, run_id, job, step, rule_id, type, severity, message, evidence, savings_mips, savings_usd)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, f := range run.Findings {
			if _, err := stmt.Exec(
				f.ID,
				run.ID,
				f.Job,
				f.Step,
				f.RuleID,
				f.Type,
				f.Severity,
				f.Message,
				f.Evidence,
				f.SavingsMIPS,
				f.SavingsUSD,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// LoadRun returns the full run (from stored JSON).
func (db *DB) LoadRun(id string) (ir.Run, error) {
	var s string
	row := db.conn.QueryRow(`SELECT run_json FROM runs WHERE id = ?`, id)
	if err := row.Scan(&s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ir.Run{}, err
		}
		return ir.Run{}, err
	}
	var run ir.Run
	if err := json.Unmarshal([]byte(s), &run); err != nil {
		return ir.Run{}, err
	}
	return run, nil
}
