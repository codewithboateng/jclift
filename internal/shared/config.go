package shared

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`

	Analysis struct {
		Sources   []string `yaml:"sources"`
		MIPSToUSD float64  `yaml:"mips_to_usd"`
	} `yaml:"analysis"`

	Reporting struct {
		OutDir string `yaml:"out_dir"`
	} `yaml:"reporting"`

	Logging struct {
		Format string `yaml:"format"` // json|text
		Level  string `yaml:"level"`  // debug|info|warn|error
	} `yaml:"logging"`

	Rules struct {
		SeverityThreshold string   `yaml:"severity_threshold"` // LOW|MEDIUM|HIGH
		Disable           []string `yaml:"disable"`            // ["RULE-ID", ...]
		Sortwk            struct {
			PrimaryCylThreshold int `yaml:"primary_cyl_threshold"` // default 500
		} `yaml:"sortwk"`
	} `yaml:"rules"`

	Cost struct {
		Geometry struct {
			TracksPerCyl  int     `yaml:"tracks_per_cyl"`  // default 15
			BytesPerTrack int     `yaml:"bytes_per_track"` // default 56664 (3390)
		} `yaml:"geometry"`
		Model struct {
			MIPSPerCPU float64 `yaml:"mips_per_cpu"` // default 1.0 (1 MIPS ≈ 1 CPU-sec)
			Sort struct {
				Alpha float64 `yaml:"alpha"` // base CPU-sec
				Beta  float64 `yaml:"beta"`  // β * MB * log2(MB)
			} `yaml:"sort"`
			Copy struct {
				Alpha float64 `yaml:"alpha"` // IEBGENER
				Beta  float64 `yaml:"beta"`  // β * MB
			} `yaml:"copy"`
			IDCAMS struct {
				Alpha float64 `yaml:"alpha"` // REPRO
				Beta  float64 `yaml:"beta"`  // β * MB
			} `yaml:"idcams"`
		} `yaml:"model"`
	} `yaml:"cost"`
}

func DefaultConfig() Config {
	var c Config
	c.Database.Driver = "sqlite"
	c.Database.DSN = "./jclift.db"
	c.Reporting.OutDir = "./reports"
	c.Logging.Format = "json"
	c.Logging.Level = "info"
	c.Rules.SeverityThreshold = "LOW"
	c.Rules.Sortwk.PrimaryCylThreshold = 500

	// Geometry defaults (approx 3390)
	c.Cost.Geometry.TracksPerCyl = 15
	c.Cost.Geometry.BytesPerTrack = 56664

	// Model defaults: small coefficients to keep totals reasonable
	c.Cost.Model.MIPSPerCPU = 1.0
	c.Cost.Model.Sort.Alpha = 0.3
	c.Cost.Model.Sort.Beta = 0.001
	c.Cost.Model.Copy.Alpha = 0.10
	c.Cost.Model.Copy.Beta  = 0.0005
	c.Cost.Model.IDCAMS.Alpha = 0.15
	c.Cost.Model.IDCAMS.Beta  = 0.0006
	return c
}

func LoadConfig(path string) (Config, error) {
	c := DefaultConfig()
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			_ = yaml.Unmarshal(b, &c)
		}
	}
	// Env overrides (examples)
	if v := os.Getenv("JCLIFT_DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("JCLIFT_MIPS_TO_USD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.Analysis.MIPSToUSD = f
		}
	}
	if v := os.Getenv("JCLIFT_LOG_FORMAT"); v != "" {
		c.Logging.Format = v
	}
	if v := os.Getenv("JCLIFT_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}
	if v := os.Getenv("JCLIFT_OUT_DIR"); v != "" {
		c.Reporting.OutDir = v
	}
	return c, nil
}
