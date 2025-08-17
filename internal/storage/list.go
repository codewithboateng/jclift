package storage

import (
	"database/sql"
	"time"

	"github.com/codewithboateng/jclift/internal/ir"
)


// ListRuns returns a lightweight list of runs with counts.
func (db *DB) ListRuns(limit, offset int) ([]RunRow, error) {
	const q = `
		SELECT r.id, r.started_at, r.source, r.ir_version,
		       (SELECT COUNT(1) FROM findings f WHERE f.run_id = r.id) AS findings
		  FROM runs r
		 ORDER BY r.started_at DESC, r.id DESC
		 LIMIT ? OFFSET ?`
	rows, err := db.conn.Query(q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RunRow
	for rows.Next() {
		var rr RunRow
		var startedAtStr string
		if err := rows.Scan(&rr.ID, &startedAtStr, &rr.Source, &rr.IRVersion, &rr.Findings); err != nil {
			return nil, err
		}
		// Parse RFC3339Nano first, fallback to RFC3339
		if t, err := time.Parse(time.RFC3339Nano, startedAtStr); err == nil {
			rr.StartedAt = t
		} else if t2, err2 := time.Parse(time.RFC3339, startedAtStr); err2 == nil {
			rr.StartedAt = t2
		} else {
			// leave zero time if unparsable (shouldn't happen)
			rr.StartedAt = time.Time{}
		}
		out = append(out, rr)
	}
	return out, rows.Err()
}

// ListFindings returns findings for a run at or above a minimum severity.
func (db *DB) ListFindings(runID, minSeverity string) ([]ir.Finding, error) {
	const q = `
		SELECT id, job, step, rule_id, type, severity, message, evidence, savings_mips, savings_usd
		  FROM findings
		 WHERE run_id = ?
		   AND (CASE severity WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 ELSE 1 END)
		       >= (CASE ? WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 ELSE 1 END)
		 ORDER BY
		       (CASE severity WHEN 'HIGH' THEN 3 WHEN 'MEDIUM' THEN 2 ELSE 1 END) DESC,
		       rule_id, job, step, id`
	rows, err := db.conn.Query(q, runID, minSeverity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ir.Finding
	for rows.Next() {
		var f ir.Finding
		if err := rows.Scan(&f.ID, &f.Job, &f.Step, &f.RuleID, &f.Type, &f.Severity, &f.Message, &f.Evidence, &f.SavingsMIPS, &f.SavingsUSD); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Optional helper used by future endpoints.
func (db *DB) HasRun(id string) (bool, error) {
	const q = `SELECT 1 FROM runs WHERE id = ? LIMIT 1`
	var one int
	err := db.conn.QueryRow(q, id).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
