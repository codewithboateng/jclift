package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/codewithboateng/jclift/internal/ir"
)

type DB struct{ sql *sql.DB }

func OpenSQLite(path string) (*DB, error) {
	s, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	return &DB{sql: s}, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) CreateSchema() error {
	_, err := d.sql.Exec(`
PRAGMA journal_mode=WAL;

CREATE TABLE IF NOT EXISTS runs(
  id TEXT PRIMARY KEY,
  started_at TEXT,
  source TEXT,
  ir_version TEXT,
  payload JSON
);

CREATE TABLE IF NOT EXISTS jobs(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL,
  name TEXT,
  class TEXT,
  owner TEXT
);

CREATE TABLE IF NOT EXISTS steps(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  job_id INTEGER NOT NULL,
  ordinal INTEGER,
  name TEXT,
  program TEXT,
  conditions TEXT,
  annotations_json JSON
);

CREATE TABLE IF NOT EXISTS findings(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL,
  job_name TEXT,
  step_name TEXT,
  rule_id TEXT,
  type TEXT,
  severity TEXT,
  message TEXT,
  evidence TEXT,
  savings_mips REAL,
  savings_usd REAL,
  metadata_json JSON
);

CREATE INDEX IF NOT EXISTS idx_findings_run ON findings(run_id);
CREATE INDEX IF NOT EXISTS idx_findings_rule ON findings(rule_id);
`)
	return err
}

func (d *DB) SaveRun(run *ir.Run) error {
	if run == nil || run.ID == "" {
		return errors.New("invalid run")
	}
	b, err := json.Marshal(run)
	if err != nil {
		return err
	}
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`INSERT INTO runs(id, started_at, source, ir_version, payload) VALUES(?,?,?,?,?)`,
		run.ID, run.StartedAt.UTC().Format(timeLayout), run.Source, run.IRVersion, string(b)); err != nil {
		return err
	}

	// Insert jobs/steps
	jobIDs := make([]int64, len(run.Jobs))
	for i, j := range run.Jobs {
		res, err := tx.Exec(`INSERT INTO jobs(run_id, name, class, owner) VALUES(?,?,?,?)`,
			run.ID, j.Name, j.Class, j.Owner)
		if err != nil {
			return err
		}
		jobID, _ := res.LastInsertId()
		jobIDs[i] = jobID

		// steps
		for _, s := range j.Steps {
			ann, _ := json.Marshal(s.Annotations)
			if _, err := tx.Exec(`INSERT INTO steps(job_id, ordinal, name, program, conditions, annotations_json)
				VALUES(?,?,?,?,?,?)`,
				jobID, s.Ordinal, s.Name, s.Program, s.Conditions, string(ann)); err != nil {
				return err
			}
		}
	}

	// Insert findings
	for _, f := range run.Findings {
		meta, _ := json.Marshal(f.Metadata)
		if _, err := tx.Exec(`INSERT INTO findings(run_id, job_name, step_name, rule_id, type, severity, message, evidence, savings_mips, savings_usd, metadata_json)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			run.ID, f.Job, f.Step, f.RuleID, f.Type, f.Severity, f.Message, f.Evidence, f.SavingsMIPS, f.SavingsUSD, string(meta)); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (d *DB) LoadRun(id string) (ir.Run, error) {
	var r ir.Run
	row := d.sql.QueryRow(`SELECT payload FROM runs WHERE id = ?`, id)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return r, err
	}
	if err := json.Unmarshal([]byte(payload), &r); err != nil {
		return r, err
	}
	return r, nil
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

func (d *DB) Summary(runID string) (string, int, int, error) {
	var jobs, finds int
	row := d.sql.QueryRow(`SELECT COUNT(1) FROM jobs WHERE run_id = ?`, runID)
	if err := row.Scan(&jobs); err != nil {
		return "", 0, 0, err
	}
	row = d.sql.QueryRow(`SELECT COUNT(1) FROM findings WHERE run_id = ?`, runID)
	if err := row.Scan(&finds); err != nil {
		return "", 0, 0, err
	}
	return fmt.Sprintf("run=%s jobs=%d findings=%d", runID, jobs, finds), jobs, finds, nil
}
