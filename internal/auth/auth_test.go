package auth

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	h, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected format %q", h)
	}
	if !VerifyPassword(h, "correct horse battery") {
		t.Fatal("correct password rejected")
	}
	if VerifyPassword(h, "correct horse batterY") || VerifyPassword(h, "") {
		t.Fatal("wrong password accepted")
	}
	h2, _ := HashPassword("correct horse battery")
	if h == h2 {
		t.Fatal("salt must be random")
	}
}

func TestVerifyRejectsMalformedAndHostileHashes(t *testing.T) {
	for _, h := range []string{
		"", "plain", "$argon2id$v=19$m=19456,t=2,p=1$AAAA", "$bcrypt$v=19$m=1,t=1,p=1$AAAA$AAAA",
		"$argon2id$v=19$m=4294967295,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAA", // would exhaust memory
		"$argon2id$v=19$m=19456,t=2,p=0$AAAAAAAAAAAAAAAAAAAAAA$AAAA",
		"$argon2id$v=19$m=19456,t=2,p=1$!!!$!!!",
	} {
		if VerifyPassword(h, "whatever") {
			t.Errorf("accepted %q", h)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if ValidatePassword("short") == nil || ValidatePassword(strings.Repeat("a", 129)) == nil {
		t.Fatal("limits not enforced")
	}
	if ValidatePassword("senha-longa-ok") != nil || ValidatePassword("ação ação ação") != nil {
		t.Fatal("valid passwords rejected (length is counted in characters)")
	}
}

func TestTokens(t *testing.T) {
	tok, hash, err := NewToken()
	if err != nil || len(tok) < 40 || hash != HashToken(tok) || hash == tok {
		t.Fatalf("token=%q hash=%q err=%v", tok, hash, err)
	}
	tok2, _, _ := NewToken()
	if tok == tok2 {
		t.Fatal("tokens must be random")
	}
	if !EqualToken("abc", "abc") || EqualToken("abc", "abd") || EqualToken("abc", "ab") {
		t.Fatal("EqualToken")
	}
}

func TestLimiterBackoffAndReset(t *testing.T) {
	var l Limiter
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 4; i++ {
		l.Fail("1.2.3.4", now)
		if ok, _ := l.Allow("1.2.3.4", now); !ok {
			t.Fatalf("blocked too early after %d failures", i+1)
		}
	}
	l.Fail("1.2.3.4", now) // 5th
	ok, wait := l.Allow("1.2.3.4", now)
	if ok || wait != time.Minute {
		t.Fatalf("ok=%v wait=%v", ok, wait)
	}
	if ok, _ := l.Allow("5.6.7.8", now); !ok {
		t.Fatal("other clients must not be affected")
	}
	later := now.Add(61 * time.Second)
	if ok, _ := l.Allow("1.2.3.4", later); !ok {
		t.Fatal("must be allowed after the wait")
	}
	l.Fail("1.2.3.4", later) // 6th: 2 minutes
	if _, wait := l.Allow("1.2.3.4", later); wait != 2*time.Minute {
		t.Fatalf("backoff = %v", wait)
	}
	for i := 0; i < 20; i++ {
		l.Fail("1.2.3.4", later)
	}
	if _, wait := l.Allow("1.2.3.4", later); wait != 15*time.Minute {
		t.Fatalf("backoff must cap at 15 min, got %v", wait)
	}
	l.Reset("1.2.3.4")
	if ok, _ := l.Allow("1.2.3.4", later); !ok {
		t.Fatal("reset must clear the block")
	}
}

func TestSetupCode(t *testing.T) {
	a, err := NewSetupCode()
	if err != nil || len(a) != 19 || a[4] != '-' || a[9] != '-' || a[14] != '-' {
		t.Fatalf("code %q err %v", a, err)
	}
	if b, _ := NewSetupCode(); a == b {
		t.Fatal("codes must be random")
	}
}

func TestAttemptIsAtomicUnderContention(t *testing.T) {
	var l Limiter
	now := time.Now()
	var wg sync.WaitGroup
	var allowed atomic.Int32
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := l.Attempt("k", now); ok {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 5 {
		t.Fatalf("exactly 5 attempts may pass, got %d", allowed.Load())
	}
}
