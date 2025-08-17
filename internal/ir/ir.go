package ir

import "time"

const Version = "1.0"

type Run struct {
	ID        string    `json:"id"`
	StartedAt time.Time `json:"started_at"`
	Source    string    `json:"source,omitempty"`
	IRVersion string    `json:"ir_version,omitempty"`

	Context  Context   `json:"context"`
	Jobs     []Job     `json:"jobs"`
	Findings []Finding `json:"findings,omitempty"`
}
type Context struct {
	MIPSToUSD             float64  `json:"mips_to_usd,omitempty"`
	RuleSeverityThreshold string   `json:"rule_severity_threshold,omitempty"`
	DisabledRules         []string `json:"disabled_rules,omitempty"`
}

type Job struct {
	Name          string `json:"name"`
	Class         string `json:"class,omitempty"`
	Owner         string `json:"owner,omitempty"`
	ProcsResolved bool   `json:"procs_resolved,omitempty"`
	Steps         []Step `json:"steps"`
}

type Step struct {
	Name        string `json:"name"`
	Program     string `json:"program"`
	Ordinal     int    `json:"ordinal"`
	DD          []DD   `json:"dd,omitempty"`
	Conditions  string `json:"conditions,omitempty"`
	Annotations Anno   `json:"annotations"`
}

type DD struct {
	DDName  string `json:"ddname"`
	Dataset string `json:"dataset,omitempty"`
	DISP    string `json:"disp,omitempty"`
	Space   string `json:"space,omitempty"`
	DCB     string `json:"dcb,omitempty"`
	Content string `json:"content,omitempty"` // SYSIN text
	Temp    bool   `json:"temp,omitempty"`
}

type Anno struct {
	Cost Cost `json:"cost,omitempty"`
}

type Cost struct {
	CPUSeconds float64 `json:"cpu_seconds,omitempty"`
	MIPS       float64 `json:"mips,omitempty"`
	USD        float64 `json:"usd,omitempty"`
}

type Finding struct {
	ID          string         `json:"id"`
	Job         string         `json:"job"`
	Step        string         `json:"step,omitempty"`
	RuleID      string         `json:"rule_id"`
	Type        string         `json:"type"`     // COST|RISK
	Severity    string         `json:"severity"` // LOW|MEDIUM|HIGH
	Message     string         `json:"message"`
	Evidence    string         `json:"evidence,omitempty"`
	SavingsMIPS float64        `json:"savings_mips,omitempty"`
	SavingsUSD  float64        `json:"savings_usd,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
