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
	return c
}

func LoadConfig(path string) (Config, error) {
	c := DefaultConfig()
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			_ = yaml.Unmarshal(b, &c)
		}
	}
	// Env overrides (example)
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
