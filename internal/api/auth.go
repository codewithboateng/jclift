package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/codewithboateng/jclift/internal/security"
)

const sessionCookie = "jclift_session"

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type meResp struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ... imports unchanged
// ... imports unchanged
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in loginReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		s.err(w, http.StatusBadRequest, "invalid json"); return
	}
	u, hash, err := s.UserStore.GetUserByUsername(in.Username) // <--
	if err != nil || !security.CheckPassword(hash, in.Password) {
		s.err(w, http.StatusUnauthorized, "invalid credentials"); return
	}
	tok, err := security.NewToken(32)
	if err != nil { s.err(w, http.StatusInternalServerError, "token error"); return }
	exp := time.Now().Add(s.SessionDuration)
	if err := s.UserStore.CreateSession(u.ID, tok, exp); err != nil { // <--
		s.err(w, http.StatusInternalServerError, "session error"); return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie, Value: tok, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: false, Expires: exp,
	})
	_ = s.UserStore.LogAudit(u.Username, "login", "", map[string]any{"ip": r.RemoteAddr}) // <--
	writeJSON(w, http.StatusOK, meResp{Username: u.Username, Role: u.Role})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	tok, err := readSessionCookie(r)
	if err == nil { _ = s.UserStore.DeleteSession(tok) } // <--
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", Expires: time.Unix(0,0), MaxAge: -1, HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}



func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromCtx(r.Context())
	if !ok { s.err(w, http.StatusUnauthorized, "unauthorized"); return }
	writeJSON(w, http.StatusOK, meResp{Username: u.Username, Role: u.Role})
}

func readSessionCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil || c.Value == "" { return "", errors.New("no session") }
	return c.Value, nil
}
