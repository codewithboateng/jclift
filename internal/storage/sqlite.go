package storage

import (
	"database/sql"
	"encoding/json"
	"errors"

	_ "modernc.org/sqlite" // pure-Go SQLite driver

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
CREATE TABLE IF NOT EXISTS runs(
  id TEXT PRIMARY KEY,
  started_at TEXT,
  source TEXT,
  ir_version TEXT,
  payload JSON
);`)
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
	_, err = d.sql.Exec(`INSERT INTO runs(id, started_at, source, ir_version, payload) VALUES(?,?,?,?,?)`,
		run.ID, run.StartedAt.UTC().Format(timeLayout), run.Source, run.IRVersion, string(b))
	return err
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
