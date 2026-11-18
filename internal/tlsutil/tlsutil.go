// Package tlsutil provides the HTTPS options of the server: a certificate the
// user supplies, or a self-signed one Atalaia makes and renews by itself.
package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// Validity of a generated certificate.
	Validity = 365 * 24 * time.Hour
	// RenewBefore is how long before expiry a generated certificate is replaced.
	RenewBefore = 30 * 24 * time.Hour
)

// SelfSigned makes sure dir holds a valid self-signed certificate that covers
// names (host names and IP addresses) and returns the certificate and key paths.
// An existing one is reused unless it expires soon or does not cover a name.
// The key is written 0600 in a 0700 directory.
func SelfSigned(dir string, names []string, now time.Time) (certPath, keyPath string, err error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", err
	}
	certPath, keyPath = filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	if c, err := loadLeaf(certPath); err == nil && now.Add(RenewBefore).Before(c.NotAfter) && covers(c, names) {
		if _, kerr := os.Stat(keyPath); kerr == nil {
			return certPath, keyPath, nil
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return "", "", err
	}
	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "Atalaia", Organization: []string{"Atalaia (self-signed)"}},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(Validity),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, n := range names {
		if ip := net.ParseIP(n); ip != nil {
			tpl.IPAddresses = append(tpl.IPAddresses, ip)
		} else if n != "" {
			tpl.DNSNames = append(tpl.DNSNames, n)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return "", "", err
	}
	kb, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", "", err
	}
	if err := writeAtomic(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: kb}), 0o600); err != nil {
		return "", "", err
	}
	if err := writeAtomic(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		return "", "", err
	}
	return certPath, keyPath, nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func loadLeaf(certPath string) (*x509.Certificate, error) {
	b, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(b)
	if blk == nil {
		return nil, errors.New("no certificate in file")
	}
	return x509.ParseCertificate(blk.Bytes)
}

func covers(c *x509.Certificate, names []string) bool {
	for _, n := range names {
		if n == "" {
			continue
		}
		if err := c.VerifyHostname(n); err != nil {
			return false
		}
	}
	return true
}

// Fingerprint returns the SHA-256 fingerprint of the certificate file, for the
// user to compare with what the browser shows.
func Fingerprint(certPath string) (string, error) {
	c, err := loadLeaf(certPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(c.Raw)
	h := strings.ToUpper(hex.EncodeToString(sum[:]))
	parts := make([]string, 0, len(h)/2)
	for i := 0; i < len(h); i += 2 {
		parts = append(parts, h[i:i+2])
	}
	return strings.Join(parts, ":"), nil
}

// Config returns a server TLS configuration that serves the certificate and
// key files and picks up a renewed certificate (a changed file) without a
// restart. TLS 1.2 is the minimum.
func Config(certPath, keyPath string) (*tls.Config, error) {
	r := &reloader{cert: certPath, key: keyPath}
	if _, err := r.get(nil); err != nil {
		return nil, fmt.Errorf("loading the TLS certificate: %w", err)
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: r.get}, nil
}

type reloader struct {
	cert, key string
	mu        sync.Mutex
	mtime     time.Time
	loaded    *tls.Certificate
}

func (r *reloader) get(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	st, err := os.Stat(r.cert)
	if err != nil {
		if r.loaded != nil {
			return r.loaded, nil // keep serving the last good one
		}
		return nil, err
	}
	if r.loaded == nil || st.ModTime().After(r.mtime) {
		c, err := tls.LoadX509KeyPair(r.cert, r.key)
		if err != nil {
			if r.loaded != nil {
				return r.loaded, nil
			}
			return nil, err
		}
		r.loaded, r.mtime = &c, st.ModTime()
	}
	return r.loaded, nil
}
