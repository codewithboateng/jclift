package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/rules"
	"github.com/codewithboateng/jclift/internal/storage"
)

// Store is the minimal contract the API needs.
type Store interface {
	ListRuns(limit, offset int) ([]storage.RunRow, error)
	LoadRun(id string) (ir.Run, error)
	ListFindings(runID, minSeverity string) ([]ir.Finding, error)

	// NEW
	LoadLatestRun() (ir.Run, error)

	// Waivers (from previous commit)
	ListWaivers(activeOnly bool) ([]storage.Waiver, error)
	CreateWaiver(ruleID, job, step, pattern, reason, createdBy string, expires time.Time) (int64, error)
	RevokeWaiver(id int64, by string) error
}

// UserStore is the auth/audit contract the API uses.
type UserStore interface {
	GetUserByUsername(string) (storage.User, string, error)
	CreateSession(int64, string, time.Time) error
	GetSession(string) (storage.User, error)
	DeleteSession(string) error
	LogAudit(username, action, resource string, meta map[string]any) error

	
}

type Server struct {
	DB              Store
	UserStore       UserStore       // <-- exported so cmd can set it
	Logger          *slog.Logger
	AllowedOrigins  []string
	SessionDuration time.Duration
}


func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h(w, r)
		}
	}

	// Health
	mux.HandleFunc("GET /api/v1/health", withCORS(s.handleHealth))

	// Auth
	mux.HandleFunc("POST /api/v1/auth/login", withCORS(s.handleLogin))
	mux.HandleFunc("POST /api/v1/auth/logout", withCORS(withAuth(s, s.handleLogout, "auth:logout")))

	// Me (ONLY ONCE)
	mux.HandleFunc("GET /api/v1/me", withCORS(withAuth(s, s.handleMe, "me")))

	// Runs
	mux.HandleFunc("GET /api/v1/runs", withCORS(s.handleListRuns))
	mux.HandleFunc("GET /api/v1/runs/latest", withCORS(s.handleGetLatest))
	mux.HandleFunc("GET /api/v1/runs/{id}", withCORS(s.handleGetRun))
	mux.HandleFunc("GET /api/v1/runs/{id}/findings", withCORS(s.handleListFindings))

	// Rules inventory
	mux.HandleFunc("GET /api/v1/rules", withCORS(s.handleRules))

	// Waivers
	mux.HandleFunc("GET /api/v1/waivers", withCORS(withAuth(s, s.handleListWaivers, "waivers:list")))
	mux.HandleFunc("POST /api/v1/waivers", withCORS(withAdmin(s, s.handleCreateWaiver, "waivers:create")))
	mux.HandleFunc("POST /api/v1/waivers/{id}/revoke", withCORS(withAdmin(s, s.handleRevokeWaiver, "waivers:revoke")))

	// Fallback 404
	mux.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	return mux
}

func (s *Server) pickCORSOrigin(r *http.Request) string {
	if len(s.AllowedOrigins) == 0 {
		return ""
	}
	origin := r.Header.Get("Origin")
	for _, ao := range s.AllowedOrigins {
		if ao == "*" {
			return "*"
		}
		if origin != "" && strings.EqualFold(origin, ao) {
			return origin
		}
	}
	// Not allowed → return empty (no CORS header)
	return ""
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := clamp(parseInt(q.Get("limit"), 20), 1, 200)
	offset := parseInt(q.Get("offset"), 0)

	rows, err := s.DB.ListRuns(limit, offset)
	if err != nil {
		s.err(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": rows, "limit": limit, "offset": offset,
	})
}

func (s *Server) handleGetLatestRun(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.ListRuns(1, 0)
	if err != nil || len(rows) == 0 {
		s.err(w, http.StatusNotFound, "no runs")
		return
	}
	run, err := s.DB.LoadRun(rows[0].ID)
	if err != nil {
		s.err(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.PathValue("id"), "")
	run, err := s.DB.LoadRun(id)
	if err != nil {
		s.err(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleListFindings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	min := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("min_severity")))
	if min == "" {
		min = "LOW"
	}
	items, err := s.DB.ListFindings(id, min)
	if err != nil {
		s.err(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"run_id": id, "min_severity": min, "items": items,
	})
}

func (s *Server) handleListRules(w http.ResponseWriter, r *http.Request) {
	type rr struct {
		ID      string `json:"id"`
		Summary string `json:"summary"`
	}
	var out []rr
	for _, r := range rules.List() {
		out = append(out, rr{ID: r.ID, Summary: r.Summary})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (s *Server) err(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
func clamp(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}



// GET /api/v1/rules (IDs + summaries; no auth needed for read-only)
func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	type R struct {
		ID      string `json:"id"`
		Summary string `json:"summary"`
	}
	var out []R
	for _, rr := range rules.List() {
		out = append(out, R{ID: rr.ID, Summary: rr.Summary})
	}
	// stable order already guaranteed by rules.List()
	writeJSON(w, http.StatusOK, map[string]any{"items": out, "count": len(out)})
}

// GET /api/v1/runs/latest
func (s *Server) handleGetLatest(w http.ResponseWriter, r *http.Request) {
	run, err := s.DB.LoadLatestRun()
	if err != nil {
		s.err(w, http.StatusNotFound, "no runs")
		return
	}
	writeJSON(w, http.StatusOK, run)
}
