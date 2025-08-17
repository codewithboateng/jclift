package shared

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Driver string `yaml:"driver"` // "sqlite" (default)
		DSN    string `yaml:"dsn"`    // "./jclift.db"
	} `yaml:"database"`

	Analysis struct {
		Sources   []string `yaml:"sources"`     // ["./samples/bank-small"]
		MIPSToUSD float64  `yaml:"mips_to_usd"` // 0 (optional)
	} `yaml:"analysis"`

	Reporting struct {
		OutDir string `yaml:"out_dir"` // "./reports"
	} `yaml:"reporting"`

	Logging struct {
		Format string `yaml:"format"` // "json"|"text"
		Level  string `yaml:"level"`  // "info"|"debug"|"warn"|"error"
	} `yaml:"logging"`
}

func DefaultConfig() Config {
	var c Config
	c.Database.Driver = "sqlite"
	c.Database.DSN = "./jclift.db"
	c.Reporting.OutDir = "./reports"
	c.Logging.Format = "json"
	c.Logging.Level = "info"
	return c
}

func LoadConfig(path string) (Config, error) {
	c := DefaultConfig()
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			_ = yaml.Unmarshal(b, &c)
		}
	}
	// Env overrides (simple, explicit)
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
