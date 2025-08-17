package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (db *DB) CreateUser(username, passHash, role string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := db.conn.Exec(`INSERT INTO users(username, pass_hash, role, created_at) VALUES(?,?,?,?)`,
		username, passHash, role, now)
	if err != nil { return 0, err }
	return res.LastInsertId()
}

func (db *DB) GetUserByUsername(username string) (User, string, error) {
	row := db.conn.QueryRow(`SELECT id, username, role, created_at, pass_hash FROM users WHERE username=?`, username)
	var u User
	var ph string
	var created string
	if err := row.Scan(&u.ID, &u.Username, &u.Role, &created, &ph); err != nil {
		return User{}, "", err
	}
	if t, err := time.Parse(time.RFC3339Nano, created); err == nil { u.CreatedAt = t }
	return u, ph, nil
}

func (db *DB) CreateSession(userID int64, token string, expires time.Time) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return execOne(db.conn, `INSERT INTO sessions(token, user_id, expires_at, created_at) VALUES(?,?,?,?)`,
		token, userID, expires.UTC().Format(time.RFC3339Nano), now)
}

func (db *DB) GetSession(token string) (User, error) {
	row := db.conn.QueryRow(`
SELECT u.id, u.username, u.role, u.created_at
FROM sessions s JOIN users u ON s.user_id=u.id
WHERE s.token=? AND s.expires_at > ?`, token, time.Now().UTC().Format(time.RFC3339Nano))
	var u User
	var created string
	if err := row.Scan(&u.ID, &u.Username, &u.Role, &created); err != nil {
		return User{}, err
	}
	if t, err := time.Parse(time.RFC3339Nano, created); err == nil { u.CreatedAt = t }
	return u, nil
}

func (db *DB) DeleteSession(token string) error {
	return execOne(db.conn, `DELETE FROM sessions WHERE token=?`, token)
}

func (db *DB) LogAudit(username, action, resource string, meta map[string]any) error {
	b, _ := json.Marshal(meta)
	_, err := db.conn.Exec(`INSERT INTO audit(ts, username, action, resource, meta_json) VALUES(?,?,?,?,?)`,
		time.Now().UTC().Format(time.RFC3339Nano), username, action, resource, string(b))
	return err
}

func execOne(db *sql.DB, q string, args ...any) error {
	res, err := db.Exec(q, args...)
	if err != nil { return err }
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
