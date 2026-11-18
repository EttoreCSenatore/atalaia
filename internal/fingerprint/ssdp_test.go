package fingerprint

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/EttoreCSenatore/atalaia/internal/scope"
)

const tvXML = `<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0"><device>
<deviceType>urn:schemas-upnp-org:device:MediaRenderer:1</deviceType>
<friendlyName>Sala TV</friendlyName><manufacturer>Samsung Electronics</manufacturer>
<modelName>UN55 Series</modelName></device></root>`

func TestParseDescription(t *testing.T) {
	d := parseDescription([]byte(tvXML))
	if d.DeviceType != "MediaRenderer" || d.Manufacturer != "Samsung Electronics" || d.Model != "UN55 Series" || d.Friendly != "Sala TV" {
		t.Fatalf("%+v", d)
	}
	// hostile XML: entities are not expanded, garbage gives nothing
	bomb := `<?xml version="1.0"?><!DOCTYPE x [<!ENTITY a "AAAAAAAAAA"><!ENTITY b "&a;&a;&a;&a;&a;&a;&a;&a;">]><root><device><friendlyName>&b;</friendlyName></device></root>`
	if d := parseDescription([]byte(bomb)); len(d.Friendly) > 50 {
		t.Fatalf("an entity was expanded: %q", d.Friendly)
	}
	if d := parseDescription([]byte("not xml")); d != (SSDPInfo{}) {
		t.Fatalf("%+v", d)
	}
}

func TestLocationMustPointBackAtTheAnswerer(t *testing.T) {
	from := netip.MustParseAddr("192.168.1.40")
	for raw, want := range map[string]bool{
		"http://192.168.1.40:8080/desc.xml": true,
		"http://192.168.1.41/desc.xml":      false, // another host: not fetched (SSRF guard)
		"http://169.254.169.254/latest":     false,
		"https://192.168.1.40/desc.xml":     false,
		"http://evil.example/desc.xml":      false,
		"http://user:pw@192.168.1.40/":      false,
		"ftp://192.168.1.40/":               false,
	} {
		if got := locationOK(raw, from); got != want {
			t.Errorf("%s: %v want %v", raw, got, want)
		}
	}
}

// A unicast "device" stands in for the multicast group: it answers the M-SEARCH
// with a LOCATION that points to a real local HTTP server.
func TestDiscoverEndToEnd(t *testing.T) {
	web := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(tvXML)) }))
	defer web.Close()
	dev, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()
	go func() {
		buf := make([]byte, 1500)
		for {
			_, from, err := dev.ReadFromUDP(buf)
			if err != nil {
				return
			}
			resp := "HTTP/1.1 200 OK\r\nST: upnp:rootdevice\r\nSERVER: Linux UPnP/1.0\r\nLOCATION: " + web.URL + "/desc.xml\r\n\r\n"
			_, _ = dev.WriteToUDP([]byte(resp), from)
		}
	}()
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")})
	_ = sc
	got, err := SSDP{Dest: dev.LocalAddr().(*net.UDPAddr), Wait: 500 * time.Millisecond, Scope: nil}.Discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	info, ok := got[netip.MustParseAddr("127.0.0.1")]
	if !ok || info.Model != "UN55 Series" || info.DeviceType != "MediaRenderer" || !strings.Contains(info.Server, "UPnP") {
		t.Fatalf("%+v %v", info, ok)
	}
}

func TestDiscoverIgnoresOutOfScopeAnswerers(t *testing.T) {
	dev, _ := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	defer dev.Close()
	go func() {
		buf := make([]byte, 1500)
		_, from, err := dev.ReadFromUDP(buf)
		if err == nil {
			_, _ = dev.WriteToUDP([]byte("HTTP/1.1 200 OK\r\nSERVER: x\r\n\r\n"), from)
		}
	}()
	sc := scope.New([]netip.Prefix{netip.MustParsePrefix("192.168.1.0/24")}) // loopback is not in scope
	got, _ := SSDP{Dest: dev.LocalAddr().(*net.UDPAddr), Wait: 300 * time.Millisecond, Scope: sc}.Discover(context.Background())
	if len(got) != 0 {
		t.Fatalf("an out-of-scope answerer was accepted: %+v", got)
	}
}
