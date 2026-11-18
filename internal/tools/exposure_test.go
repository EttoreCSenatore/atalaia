package tools

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"testing"
	"time"
)

func TestAssessRanksAndIgnoresSafePorts(t *testing.T) {
	open := []Port{{22, "ssh"}, {445, "smb"}, {80, "http"}, {2375, "docker-api"}, {23, "telnet"}, {443, "https"}}
	got := Assess(open)
	if len(got) != 3 || got[0].Service != "docker-api" || got[0].Severity != SevCritical || got[1].Service != "telnet" || got[2].Service != "smb" {
		t.Fatalf("%+v", got)
	}
	if got := Assess([]Port{{22, "ssh"}, {80, "http"}}); len(got) != 0 {
		t.Fatalf("safe ports flagged: %+v", got)
	}
}

func TestEveryRiskyServiceExistsInTheCommonPorts(t *testing.T) {
	known := map[string]bool{}
	for _, s := range CommonPorts {
		known[s] = true
	}
	for s := range riskyServices {
		if !known[s] {
			t.Errorf("risky service %q is not a scanned port, so it can never be reported", s)
		}
	}
}

func hostPort(t *testing.T, raw string) (netip.Addr, int) {
	t.Helper()
	h, p, _ := net.SplitHostPort(raw[len("http://"):])
	port, _ := strconv.Atoi(p)
	return netip.MustParseAddr(h), port
}

func TestWebProbeReadsTitleServerAndAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "lighttpd/1.4\r")
		w.Header().Set("WWW-Authenticate", `Basic realm="x"`)
		w.WriteHeader(401)
		_, _ = w.Write([]byte("<html><head><TITLE>\n Router   Admin\x1b[0m </TITLE></head>"))
	}))
	defer ts.Close()
	ip, port := hostPort(t, ts.URL)
	info, ok := WebProbe(context.Background(), ip, port, false, nil)
	if !ok || info.Status != 401 || !info.Auth || info.Server != "lighttpd/1.4" || info.Title != "Router Admin [0m" {
		t.Fatalf("%+v %v", info, ok)
	}
	if _, ok := WebProbe(context.Background(), netip.MustParseAddr("127.0.0.1"), 1, false, nil); ok {
		t.Fatal("a closed port answered")
	}
}

func TestWebProbeDoesNotFollowRedirects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/", 302)
	}))
	defer ts.Close()
	ip, port := hostPort(t, ts.URL)
	info, ok := WebProbe(context.Background(), ip, port, false, nil)
	if !ok || info.Status != 302 || info.Redirect != "http://169.254.169.254/" {
		t.Fatalf("%+v", info)
	}
}

func selfSigned(t *testing.T, notAfter time.Time) tls.Certificate {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "nas.lan"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: notAfter}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

func TestTLSCheckDescribesASelfSignedCertificate(t *testing.T) {
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{selfSigned(t, time.Now().Add(10*24*time.Hour))}, MinVersion: tls.VersionTLS12})
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
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(p)
	info, ok := TLSCheck(context.Background(), netip.MustParseAddr("127.0.0.1"), port, nil, time.Now())
	if !ok || !info.SelfSigned || info.Expired || info.DaysLeft < 9 || info.DaysLeft > 10 || info.Version != "TLS 1.3" || info.LegacyTLS {
		t.Fatalf("%+v %v", info, ok)
	}
	if info.Subject != "CN=nas.lan" {
		t.Fatalf("subject %q", info.Subject)
	}
}

func TestTLSCheckFlagsExpired(t *testing.T) {
	ln, _ := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{selfSigned(t, time.Now().Add(-48*time.Hour))}})
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
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(p)
	info, ok := TLSCheck(context.Background(), netip.MustParseAddr("127.0.0.1"), port, nil, time.Now())
	if !ok || !info.Expired || info.DaysLeft >= 0 {
		t.Fatalf("%+v", info)
	}
}

func TestTLSCheckOnAPlainPortFails(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_, _ = c.Write([]byte("SSH-2.0-test\r\n"))
			c.Close()
		}
	}()
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(p)
	if _, ok := TLSCheck(context.Background(), netip.MustParseAddr("127.0.0.1"), port, nil, time.Now()); ok {
		t.Fatal("a non-TLS port reported TLS")
	}
}
