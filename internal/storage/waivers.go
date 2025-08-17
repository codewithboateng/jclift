package storage

import (
	"database/sql"
	"time"
)

type Waiver struct {
	ID         int64     `json:"id"`
	RuleID     string    `json:"rule_id"`
	Job        string    `json:"job,omitempty"`
	Step       string    `json:"step,omitempty"`
	PatternSub string    `json:"pattern_sub,omitempty"`
	Reason     string    `json:"reason"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
}

func (db *DB) CreateWaiver(ruleID, job, step, pattern, reason, createdBy string, expires time.Time) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := db.conn.Exec(`
INSERT INTO waivers(rule_id, job, step, pattern_sub, reason, expires_at, created_by, created_at)
VALUES(?,?,?,?,?,?,?,?)`,
		ruleID, nz(job), nz(step), nz(pattern), reason, expires.UTC().Format(time.RFC3339Nano), createdBy, now)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func (db *DB) RevokeWaiver(id int64, by string) error {
	// we store revoker in audit; waivers table only has revoked_at
	_, err := db.conn.Exec(`UPDATE waivers SET revoked_at=? WHERE id=? AND revoked_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339Nano), id)
	return err
}

func (db *DB) ListWaivers(activeOnly bool) ([]Waiver, error) {
	q := `
SELECT id, rule_id, COALESCE(job,''), COALESCE(step,''), COALESCE(pattern_sub,''),
       reason, expires_at, created_by, created_at, revoked_at
FROM waivers`
	args := []any{}
	if activeOnly {
		q += ` WHERE (revoked_at IS NULL) AND (expires_at > ?)`
		args = append(args, time.Now().UTC().Format(time.RFC3339Nano))
	}
	q += ` ORDER BY id DESC`
	rows, err := db.conn.Query(q, args...)
	if err != nil { return nil, err }
	defer rows.Close()

	var out []Waiver
	for rows.Next() {
		var (
			w Waiver
			exp, ca, ra sql.NullString
			job, step, pat string
		)
		if err := rows.Scan(&w.ID, &w.RuleID, &job, &step, &pat, &w.Reason, &exp, &w.CreatedBy, &ca, &ra); err != nil {
			return nil, err
		}
		w.Job, w.Step, w.PatternSub = job, step, pat
		if exp.Valid { if t, e := time.Parse(time.RFC3339Nano, exp.String); e == nil { w.ExpiresAt = t } }
		if ca.Valid  { if t, e := time.Parse(time.RFC3339Nano, ca.String); e == nil { w.CreatedAt = t } }
		if ra.Valid  { if t, e := time.Parse(time.RFC3339Nano, ra.String); e == nil { w.RevokedAt = &t } }
		out = append(out, w)
	}
	return out, rows.Err()
}

func nz(s string) any {
	if s == "" { return nil }
	return s
}
