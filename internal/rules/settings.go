package rules

import "strings"

type Settings struct {
	SeverityThreshold         string
	Disabled                  map[string]bool
	SortwkPrimaryCylThreshold int
}

var rsettings = Settings{
	SeverityThreshold:         "LOW",
	Disabled:                  map[string]bool{},
	SortwkPrimaryCylThreshold: 500,
}

func SetSettings(s Settings) {
	// fill defaults
	if s.SeverityThreshold == "" {
		s.SeverityThreshold = rsettings.SeverityThreshold
	}
	if s.Disabled == nil {
		s.Disabled = map[string]bool{}
	}
	if s.SortwkPrimaryCylThreshold == 0 {
		s.SortwkPrimaryCylThreshold = rsettings.SortwkPrimaryCylThreshold
	}
	rsettings = s
}

func severityRank(sev string) int {
	switch strings.ToUpper(strings.TrimSpace(sev)) {
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	default:
		return 1 // LOW or unknown → LOW
	}
}

func severityOK(sev string) bool {
	return severityRank(sev) >= severityRank(rsettings.SeverityThreshold)
}
