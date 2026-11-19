// Package auth handles the administrator's password, session tokens and
// brute-force protection.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP minimum profile: 19 MiB, 2 iterations, 1 lane),
// light enough for a Raspberry Pi.
const (
	argonMemory  = 19 * 1024
	argonTime    = 2
	argonThreads = 1
	argonKeyLen  = 32
	saltLen      = 16
)

// Password length limits (characters). The minimum is the only complexity rule:
// length matters more than symbols.
const (
	MinPassword = 10
	MaxPassword = 128
)

// ErrWeakPassword is returned for passwords outside the length limits.
var ErrWeakPassword = fmt.Errorf("the password must have between %d and %d characters", MinPassword, MaxPassword)

// ValidatePassword checks the length limits.
func ValidatePassword(pw string) error {
	n := utf8.RuneCountInString(pw)
	if n < MinPassword || n > MaxPassword {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword returns an encoded argon2id hash with a random salt.
func HashPassword(pw string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(pw), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

// VerifyPassword checks pw against an encoded hash in constant time.
func VerifyPassword(encoded, pw string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version, m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	// refuse absurd parameters from a tampered database (denial of service)
	if m > 1<<20 || t > 10 || p == 0 || p > 16 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(want) == 0 || len(want) > 128 {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, m, p, uint32(len(want))) // #nosec G115 -- bounded above
	return subtle.ConstantTimeCompare(got, want) == 1
}

// NewToken returns a random session token (for the cookie) and its hash (for
// the database). Only the hash is ever stored.
func NewToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token), nil
}

// HashToken returns the SHA-256 (hex) of a token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// EqualToken compares two secrets in constant time.
func EqualToken(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// Limiter slows down password guessing: after MaxFails consecutive failures
// for a key (the client address), further attempts are refused for an
// exponentially growing time (1 min, 2, 4... up to 15 min). In memory only.
type Limiter struct {
	MaxFails int           // default 5
	MaxWait  time.Duration // longest lockout, default 15 min
	mu       sync.Mutex
	state    map[string]*entry
}

type entry struct {
	fails int
	until time.Time
	last  time.Time
}

// Allow reports whether key may try now, and if not, how long to wait.
func (l *Limiter) Allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e, ok := l.state[key]; ok && now.Before(e.until) {
		return false, e.until.Sub(now)
	}
	return true, 0
}

// Fail records a failed attempt.
func (l *Limiter) Fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state == nil {
		l.state = map[string]*entry{}
	}
	max := l.MaxFails
	if max == 0 {
		max = 5
	}
	for k, e := range l.state { // forget old entries so the map cannot grow forever
		if now.Sub(e.last) > time.Hour {
			delete(l.state, k)
		}
	}
	e := l.state[key]
	if e == nil {
		e = &entry{}
		l.state[key] = e
	}
	e.fails++
	e.last = now
	if e.fails >= max {
		shift := e.fails - max
		if shift > 4 {
			shift = 4
		}
		maxWait := l.MaxWait
		if maxWait == 0 {
			maxWait = 15 * time.Minute
		}
		wait := time.Minute << shift
		if wait > maxWait {
			wait = maxWait
		}
		e.until = now.Add(wait)
	}
}

// Attempt counts an attempt before the password is checked and reports whether
// it may go ahead. Counting first means parallel guesses cannot all slip past
// the limit while the (slow) check runs; a successful login calls Reset.
func (l *Limiter) Attempt(key string, now time.Time) (bool, time.Duration) {
	if ok, wait := l.Allow(key, now); !ok {
		return false, wait
	}
	l.Fail(key, now)
	return true, 0
}

// Reset forgets a key after a successful login.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.state, key)
}

// NewSetupCode returns the one-time code that proves, on first run, that
// whoever creates the administrator account can read the server's console.
// 16 characters from a 32-symbol alphabet without look-alikes (80 bits).
func NewSetupCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, 0, 19)
	for i, x := range b {
		if i > 0 && i%4 == 0 {
			out = append(out, '-')
		}
		out = append(out, alphabet[int(x)%len(alphabet)])
	}
	return string(out), nil
}

// verifySlots bounds how many argon2id computations run at once. Each one uses
// about 19 MiB, so an unauthenticated flood of logins could otherwise exhaust
// the memory of a small machine such as a Raspberry Pi.
var verifySlots = make(chan struct{}, 2)

// VerifyPasswordLimited is VerifyPassword with at most two computations at a
// time; it gives up (false) if ctx ends while waiting for a slot.
func VerifyPasswordLimited(ctx interface{ Done() <-chan struct{} }, hash, password string) bool {
	select {
	case verifySlots <- struct{}{}:
	case <-ctx.Done():
		return false
	}
	defer func() { <-verifySlots }()
	return VerifyPassword(hash, password)
}
