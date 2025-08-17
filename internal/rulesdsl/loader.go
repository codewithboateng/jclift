package rulesdsl

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/rules"
)

type dslPack struct {
	Rules []dslRule `yaml:"rules"`
}

type dslRule struct {
	ID       string `yaml:"id"`
	Summary  string `yaml:"summary"`
	Type     string `yaml:"type"`     // COST|RISK
	Severity string `yaml:"severity"` // LOW|MEDIUM|HIGH
	Message  string `yaml:"message"`

	Where struct {
		Program    string `yaml:"program"`      // regex (case-insensitive)
		DDName     string `yaml:"ddname"`       // require a DD with this name (optional)
		SysinRegex string `yaml:"sysin_regex"`  // regex on SYSIN text (optional)
	} `yaml:"where"`

	Savings struct {
		Kind string  `yaml:"kind"` // "step_cost" or "mips"
		MIPS float64 `yaml:"mips"` // used if kind=="mips"
	} `yaml:"savings"`
}

type compiled struct {
	rule        dslRule
	reProgram   *regexp.Regexp
	reSysin     *regexp.Regexp
	needDDName  string
}

func LoadAndRegister(path string) (int, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read rules pack: %w", err)
	}
	var pack dslPack
	if err := yaml.Unmarshal(b, &pack); err != nil {
		return 0, fmt.Errorf("parse yaml: %w", err)
	}
	var n int
	for _, r := range pack.Rules {
		cr, err := compile(r)
		if err != nil {
			return n, fmt.Errorf("compile rule %q: %w", r.ID, err)
		}
		registerCompiled(*cr)
		n++
	}
	return n, nil
}

func compile(r dslRule) (*compiled, error) {
	if r.ID == "" || r.Type == "" || r.Severity == "" || r.Message == "" {
		return nil, fmt.Errorf("missing required fields (id/type/severity/message)")
	}
	c := &compiled{rule: r, needDDName: strings.ToUpper(strings.TrimSpace(r.Where.DDName))}
	if r.Where.Program != "" {
		re, err := regexp.Compile("(?i)" + r.Where.Program)
		if err != nil { return nil, fmt.Errorf("program regex: %w", err) }
		c.reProgram = re
	}
	if r.Where.SysinRegex != "" {
		re, err := regexp.Compile("(?i)" + r.Where.SysinRegex)
		if err != nil { return nil, fmt.Errorf("sysin_regex: %w", err) }
		c.reSysin = re
	}
	return c, nil
}

func registerCompiled(c compiled) {
	rules.Register(rules.Rule{
		ID:      c.rule.ID,
		Summary: c.rule.Summary,
		Eval: func(job *ir.Job) []ir.Finding {
			var out []ir.Finding
			for _, st := range job.Steps {
				// program match
				if c.reProgram != nil && !c.reProgram.MatchString(st.Program) {
					continue
				}
				// DD presence (optional)
				if c.needDDName != "" {
					found := false
					for _, dd := range st.DD {
						if strings.EqualFold(dd.DDName, c.needDDName) {
							found = true; break
						}
					}
					if !found { continue }
				}
				// SYSIN regex (optional)
				if c.reSysin != nil {
					sysin := ""
					for _, dd := range st.DD {
						if strings.EqualFold(dd.DDName, "SYSIN") {
							sysin = dd.Content; break
						}
					}
					if !c.reSysin.MatchString(sysin) {
						continue
					}
				}

				// Savings
				sav := 0.0
				switch strings.ToLower(c.rule.Savings.Kind) {
				case "step_cost":
					sav = st.Annotations.Cost.MIPS
				case "mips":
					sav = c.rule.Savings.MIPS
				}
				out = append(out, ir.Finding{
					RuleID:      c.rule.ID,
					Type:        strings.ToUpper(c.rule.Type),
					Severity:    strings.ToUpper(c.rule.Severity),
					Job:         job.Name,
					Step:        st.Name,
					Message:     c.rule.Message,
					Evidence:    evidenceFor(st, c),
					SavingsMIPS: sav,
				})
			}
			return out
		},
	})
}

func evidenceFor(st ir.Step, c compiled) string {
	parts := []string{"PGM=" + st.Program}
	if c.needDDName != "" {
		parts = append(parts, "has DD="+c.needDDName)
	}
	if c.reSysin != nil {
		txt := ""
		for _, dd := range st.DD {
			if strings.EqualFold(dd.DDName, "SYSIN") { txt = dd.Content; break }
		}
		if len(strings.TrimSpace(txt)) > 80 {
			txt = strings.TrimSpace(txt[:80]) + "..."
		}
		if txt == "" { txt = "(empty SYSIN)" }
		parts = append(parts, "SYSIN~")
		parts = append(parts, txt)
	}
	return strings.Join(parts, " | ")
}
