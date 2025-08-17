package storage

import "time"

// RunRow is a lightweight listing row for /runs.
type RunRow struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Source    string    `json:"source,omitempty"`
	IRVersion string    `json:"ir_version,omitempty"`
	Findings  int       `json:"findings"`
}
