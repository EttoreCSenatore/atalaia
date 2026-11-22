package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ErrUserExists is returned by CreateUser when the administrator already exists.
var ErrUserExists = errors.New("the administrator account already exists")

// HasUser reports whether the administrator account has been created.
func (s *Store) HasUser(ctx context.Context) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n > 0, err
}

// CreateUser stores the administrator's password hash. It fails with
// ErrUserExists if there already is one, so two simultaneous setups cannot
// both succeed.
func (s *Store) CreateUser(ctx context.Context, hash string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO users (id, password_hash, created_at) VALUES (1, ?, ?)`, hash, at.Unix())
	if err != nil && isConstraint(err) {
		return ErrUserExists
	}
	return err
}

// PasswordHash returns the administrator's password hash ("" if none).
func (s *Store) PasswordHash(ctx context.Context) (string, error) {
	var h string
	err := s.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = 1`).Scan(&h)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return h, err
}

// ChangePassword replaces the password hash and ends every session except keep
// (pass "" to end them all).
func (s *Store) ChangePassword(ctx context.Context, hash, keepTokenHash string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = 1`, hash); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash <> ?`, keepTokenHash); err != nil {
		return err
	}
	return tx.Commit()
}

// CreateSession records a session (by the hash of its token) valid for ttl.
func (s *Store) CreateSession(ctx context.Context, tokenHash string, now time.Time, ttl time.Duration) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO sessions (token_hash, created_at, expires_at) VALUES (?, ?, ?)`,
		tokenHash, now.Unix(), now.Add(ttl).Unix())
	return err
}

// SessionValid reports whether the session exists and has not expired.
func (s *Store) SessionValid(ctx context.Context, tokenHash string, now time.Time) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE token_hash = ? AND expires_at > ?`, tokenHash, now.Unix()).Scan(&n)
	return n > 0, err
}

// DeleteSession ends a session.
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// ResetAccount removes the administrator account and every session, so the
// next `atalaia serve` asks for a new account with a one-time setup code.
// Devices, history and settings are kept. It reports whether there was an
// account to remove.
func (s *Store) ResetAccount(ctx context.Context) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions`); err != nil {
		return false, err
	}
	r, err := tx.ExecContext(ctx, `DELETE FROM users`)
	if err != nil {
		return false, err
	}
	n, _ := r.RowsAffected()
	return n > 0, tx.Commit()
}

// PurgeSessions removes expired sessions.
func (s *Store) PurgeSessions(ctx context.Context, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, now.Unix())
	return err
}

// Setting returns a stored setting or def.
func (s *Store) Setting(ctx context.Context, key, def string) (string, error) {
	var v string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return def, nil
	}
	return v, err
}

// SettingInt is Setting for integers; an unparsable value yields def.
func (s *Store) SettingInt(ctx context.Context, key string, def int) (int, error) {
	v, err := s.Setting(ctx, key, strconv.Itoa(def))
	if err != nil {
		return def, err
	}
	n, perr := strconv.Atoi(v)
	if perr != nil {
		return def, nil
	}
	return n, nil
}

// SetSetting stores a setting.
func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

// PurgeEvents deletes events older than before and returns how many.
func (s *Store) PurgeEvents(ctx context.Context, before time.Time) (int64, error) {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM audit WHERE at < ?`, before.Unix()); err != nil {
		return 0, err
	}
	r, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE at < ?`, before.Unix())
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

func isConstraint(err error) bool {
	// modernc.org/sqlite reports "constraint failed" in the message.
	return err != nil && (strings.Contains(err.Error(), "constraint failed") || strings.Contains(err.Error(), "UNIQUE"))
}
