package api

import (
	"context"
	"net/http"

	"github.com/codewithboateng/jclift/internal/storage"
)

type ctxKey int
const userKey ctxKey = 1

func withAuth(s *Server, next http.HandlerFunc, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := readSessionCookie(r)
	if err != nil {
			s.err(w, http.StatusUnauthorized, "unauthorized"); return
		}
		u, err := s.UserStore.GetSession(tok) // <--
		if err != nil {
			s.err(w, http.StatusUnauthorized, "unauthorized"); return
		}
		_ = s.UserStore.LogAudit(u.Username, action, r.URL.Path, map[string]any{"method": r.Method}) // <--
		ctx := context.WithValue(r.Context(), userKey, u)
		next(w, r.WithContext(ctx))
	}
}

func userFromCtx(ctx context.Context) (storage.User, bool) {
	u, ok := ctx.Value(userKey).(storage.User)
	return u, ok
}
