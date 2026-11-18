package tlsutil

import (
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSelfSignedCreatesReusesAndRenews(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tls")
	now := time.Now()
	names := []string{"localhost", "127.0.0.1", "atalaia.lan", "192.168.1.20"}

	cert, key, err := SelfSigned(dir, names, now)
	if err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Stat(key); st.Mode().Perm() != 0o600 {
		t.Fatalf("key mode %v", st.Mode())
	}
	if st, _ := os.Stat(dir); st.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode %v", st.Mode())
	}
	c, err := loadLeaf(cert)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		if err := c.VerifyHostname(n); err != nil {
			t.Errorf("does not cover %s: %v", n, err)
		}
	}
	first, _ := os.ReadFile(cert)

	// reused while valid and covering
	if _, _, err := SelfSigned(dir, names[:2], now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(cert); string(again) != string(first) {
		t.Fatal("a valid certificate was replaced")
	}
	// a new name forces a new certificate
	if _, _, err := SelfSigned(dir, append(names, "new.lan"), now); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(cert); string(again) == string(first) {
		t.Fatal("a certificate that misses a name was kept")
	}
	// close to expiry renews
	second, _ := os.ReadFile(cert)
	if _, _, err := SelfSigned(dir, append(names, "new.lan"), now.Add(Validity-RenewBefore+time.Hour)); err != nil {
		t.Fatal(err)
	}
	if again, _ := os.ReadFile(cert); string(again) == string(second) {
		t.Fatal("an expiring certificate was kept")
	}
}

func TestFingerprintFormat(t *testing.T) {
	cert, _, err := SelfSigned(t.TempDir(), []string{"localhost"}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	fp, err := Fingerprint(cert)
	if err != nil || len(fp) != 32*3-1 || strings.ToUpper(fp) != fp || strings.Count(fp, ":") != 31 {
		t.Fatalf("%q %v", fp, err)
	}
}

func TestConfigServesAndReloadsTheCertificate(t *testing.T) {
	dir := t.TempDir()
	cert, key, _ := SelfSigned(dir, []string{"localhost", "127.0.0.1"}, time.Now())
	cfg, err := Config(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MinVersion < tls.VersionTLS12 {
		t.Fatal("TLS below 1.2 allowed")
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() { _ = c.(*tls.Conn).Handshake(); c.Close() }()
		}
	}()
	serial := func() string {
		c, err := tls.Dial("tcp", ln.Addr().String(), &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		return c.ConnectionState().PeerCertificates[0].SerialNumber.String()
	}
	before := serial()
	// renew on disk (a name change makes a new certificate) and make sure the mtime moves
	time.Sleep(20 * time.Millisecond)
	if _, _, err := SelfSigned(dir, []string{"localhost", "127.0.0.1", "other.lan"}, time.Now()); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Second)
	_ = os.Chtimes(cert, future, future)
	if after := serial(); after == before {
		t.Fatal("the renewed certificate was not picked up")
	}
	_ = net.IPv4zero
}

func TestConfigRejectsMissingFiles(t *testing.T) {
	if _, err := Config("/nonexistent/c.pem", "/nonexistent/k.pem"); err == nil {
		t.Fatal("expected an error")
	}
}
