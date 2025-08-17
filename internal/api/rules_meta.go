package api

import (
	"net/http"

	"github.com/codewithboateng/jclift/internal/rules"
)

func (s *Server) handleRulesMeta(w http.ResponseWriter, r *http.Request) {
	type R struct {
		ID              string `json:"id"`
		Summary         string `json:"summary"`
		Type            string `json:"type"`
		DefaultSeverity string `json:"default_severity"`
		Docs            string `json:"docs,omitempty"`
	}
	var out []R
	for _, rr := range rules.List() {
		out = append(out, R{
			ID: rr.ID, Summary: rr.Summary, Type: rr.Type,
			DefaultSeverity: rr.DefaultSeverity, Docs: rr.Docs,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "count": len(out)})
}
