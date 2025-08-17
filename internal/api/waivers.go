package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type waiverCreateReq struct {
	RuleID     string `json:"rule_id"`
	Job        string `json:"job,omitempty"`
	Step       string `json:"step,omitempty"`
	PatternSub string `json:"pattern_sub,omitempty"`
	Reason     string `json:"reason"`
	ExpiresAt  string `json:"expires_at"` // ISO8601
}

func (s *Server) handleListWaivers(w http.ResponseWriter, r *http.Request) {
	active := r.URL.Query().Get("active")
	only := active == "1" || active == "true" || active == "yes"
	ws, err := s.DB.ListWaivers(only)
	if err != nil {
		s.err(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": ws, "active_only": only})
}

func (s *Server) handleCreateWaiver(w http.ResponseWriter, r *http.Request) {
	var in waiverCreateReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		s.err(w, http.StatusBadRequest, "invalid json"); return
	}
	if in.RuleID == "" || in.Reason == "" || in.ExpiresAt == "" {
		s.err(w, http.StatusBadRequest, "rule_id, reason, expires_at required"); return
	}
	exp, err := time.Parse(time.RFC3339Nano, in.ExpiresAt)
	if err != nil {
		// try looser format
		exp, err = time.Parse(time.RFC3339, in.ExpiresAt)
		if err != nil {
			s.err(w, http.StatusBadRequest, "bad expires_at (use RFC3339)"); return
		}
	}
	u, ok := userFromCtx(r.Context()); if !ok { s.err(w, http.StatusUnauthorized, "unauthorized"); return }
	id, err := s.DB.CreateWaiver(in.RuleID, in.Job, in.Step, in.PatternSub, in.Reason, u.Username, exp)
	if err != nil {
		s.err(w, http.StatusInternalServerError, "db error: "+err.Error()); return
	}
	_ = s.UserStore.LogAudit(u.Username, "waiver:create", "", map[string]any{"id": id, "rule": in.RuleID})
	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) handleRevokeWaiver(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		s.err(w, http.StatusBadRequest, "invalid id"); return
	}
	u, ok := userFromCtx(r.Context()); if !ok { s.err(w, http.StatusUnauthorized, "unauthorized"); return }
	if err := s.DB.RevokeWaiver(id, u.Username); err != nil {
		s.err(w, http.StatusInternalServerError, "db error: "+err.Error()); return
	}
	_ = s.UserStore.LogAudit(u.Username, "waiver:revoke", "", map[string]any{"id": id})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
