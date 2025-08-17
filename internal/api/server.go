package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/codewithboateng/jclift/internal/ir"
	"github.com/codewithboateng/jclift/internal/storage"
)

// Store is the minimal contract the API needs.
// Your concrete storage type just has to implement these.
type Store interface {
	ListRuns(limit, offset int) ([]storage.RunRow, error)
	LoadRun(id string) (ir.Run, error)
	ListFindings(runID, minSeverity string) ([]ir.Finding, error)
}

type Server struct {
	DB     Store
	Logger *slog.Logger
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// CORS preflight + basic CORS on all routes
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h(w, r)
		}
	}

	mux.HandleFunc("GET /api/v1/health", withCORS(s.handleHealth))
	mux.HandleFunc("GET /api/v1/runs", withCORS(s.handleListRuns))
	mux.HandleFunc("GET /api/v1/runs/{id}", withCORS(s.handleGetRun))
	mux.HandleFunc("GET /api/v1/runs/{id}/findings", withCORS(s.handleListFindings))

	// fallback
	mux.HandleFunc("/", withCORS(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	return mux
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
		"items": rows,
		"limit": limit, "offset": offset,
	})
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
